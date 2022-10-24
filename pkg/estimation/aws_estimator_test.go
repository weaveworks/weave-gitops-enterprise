package estimation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

func TestAWSClusterEstimator_Estimate(t *testing.T) {
	estimationTests := []struct {
		filename string
		want     *CostEstimate
	}{
		{
			// We have 3 instances of t3.medium in the controlPLane
			// and 5 t3.large in the machineDeployment
			// MonthlyHours == 750
			// regionCode = us-iso-east-1
			// controlPlane = 5 * 750.0 * 0.04, 0.08, 0.09 = 150.0, 300.0, 337.5
			// infrastructure = 3 * 750.0 * 0.02, 0.04, 0.05 = 45.0, 90.0, 112.5
			// max = 337.5+112.5
			// min = 150+45
			filename: "testdata/cluster-template.yaml",
			want:     &CostEstimate{High: 450.00, Low: 195.00, Currency: "USD"},
		},
		{
			// We have 6 instances of t3.medium in the controlPLane
			// and 10 t3.large in the machineDeployment
			// MonthlyHours == 750
			// regionCode = us-iso-west-1
			// controlPlane = 6 * 750.0 * 0.03, 0.06, 0.07 = 135.0, 270.0, 315.0
			// infrastructure = 10 * 750.0 * 0.03, 0.06, 0.07 = 225.0, 450.0, 525.0
			// max = 315.0+525.0
			// min = 135.0+225.0
			filename: "testdata/cluster-template-machinepool.yaml",
			want:     &CostEstimate{High: 840.00, Low: 360.00, Currency: "USD"},
		},
	}

	for _, tt := range estimationTests {
		t.Run(tt.filename, func(t *testing.T) {
			pricer := newFakeAWSPricer()
			pricer.addPrices("AmazonEC2", "USD", map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-east-1",
				"instanceType":    "t3.large",
			}, []float32{0.04, 0.08, 0.09})
			pricer.addPrices("AmazonEC2", "USD", map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-east-1",
				"instanceType":    "t3.medium",
			}, []float32{0.02, 0.04, 0.05})
			pricer.addPrices("AmazonEC2", "USD", map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-west-1",
				"instanceType":    "t3.medium",
			}, []float32{0.03, 0.06, 0.07})
			pricer.addPrices("AmazonEC2", "USD", map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-west-1",
				"instanceType":    "t3.large",
			}, []float32{0.03, 0.06, 0.07})

			estimator := NewAWSClusterEstimator(pricer, map[string]string{
				"operatingSystem": "Linux",
			})
			price, err := estimator.Estimate(context.TODO(), testParseMultiDoc(t, tt.filename))
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, price, compareFloat32); diff != "" {
				t.Fatalf("failed to calculate price:\n%s", diff)
			}
		})
	}
}

func TestAWSClusterEstimator_Estimate_errors(t *testing.T) {
	estimationTests := []struct {
		filename string
		wantErr  string
	}{
		{
			filename: "testdata/cluster-template.yaml",
			wantErr:  "no price data returned for instanceType t3.medium in region us-iso-east-1",
		},
		{
			filename: "testdata/incomplete-cluster.yaml",
			wantErr:  "could not find infrastructure infrastructure.cluster.x-k8s.io/v1beta2, Kind=AWSCluster:test-cluster",
		},
		{
			filename: "testdata/cluster-template-machinepool.yaml",
			wantErr:  "error getting prices for estimation: failed to query",
		},
		{
			filename: "testdata/invalid-cluster.yaml",
			wantErr:  "failed to parse Cluster infrastructureRef \"test-cluster\": missing reference: spec.infrastructureRef",
		},
	}

	for _, tt := range estimationTests {
		t.Run(tt.filename, func(t *testing.T) {
			pricer := newFakeAWSPricer()
			pricer.addPrices("AmazonEC2", "USD", map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-west-1",
				"instanceType":    "t3.medium",
			}, []float32{0.03, 0.06, 0.07})
			pricer.addPricesError("AmazonEC2", "USD", map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-west-1",
				"instanceType":    "t3.large",
			}, errors.New("failed to query"))

			estimator := NewAWSClusterEstimator(pricer, map[string]string{
				"operatingSystem": "Linux",
			})
			_, err := estimator.Estimate(context.TODO(), testParseMultiDoc(t, tt.filename))
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func testParseMultiDoc(t *testing.T, filename string) []*unstructured.Unstructured {
	t.Helper()
	b, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	manifests := bytes.Split(b, []byte("\n---\n"))
	parsed := []*unstructured.Unstructured{}

	for _, v := range manifests {
		dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		uns := &unstructured.Unstructured{}
		_, _, err := dec.Decode(v, nil, uns)
		if err != nil {
			t.Fatalf("failed to decode the YAML: %s", err)
		}
		parsed = append(parsed, uns)
	}

	return parsed
}

func newFakeAWSPricer() *fakeAWSPricer {
	return &fakeAWSPricer{
		prices: map[string][]float32{},
		errors: map[string]error{},
	}
}

type fakeAWSPricer struct {
	prices map[string][]float32
	errors map[string]error
}

func (f *fakeAWSPricer) ListPrices(ctx context.Context, service, currency string, filters map[string]string) ([]float32, error) {
	err, ok := f.errors[fmt.Sprintf("%s:%s:%v", service, currency, filters)]
	if ok {
		return nil, err
	}

	p, ok := f.prices[fmt.Sprintf("%s:%s:%v", service, currency, filters)]
	if !ok {
		return []float32{}, nil
	}
	return p, nil
}

func (f *fakeAWSPricer) addPrices(service, currency string, filters map[string]string, prices []float32) {
	f.prices[fmt.Sprintf("%s:%s:%v", service, currency, filters)] = prices
}

func (f *fakeAWSPricer) addPricesError(service, currency string, filters map[string]string, err error) {
	f.errors[fmt.Sprintf("%s:%s:%v", service, currency, filters)] = err
}

var compareFloat32 cmp.Option = cmp.Comparer(func(x, y float32) bool {
	delta := math.Abs(float64(x - y))
	mean := math.Abs(float64(x+y)) / 2.0

	return delta/mean < 0.00001
})
