package estimation

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/tonglil/buflogr"
)

func TestCSVPricer_ListPrices(t *testing.T) {
	queryTests := []struct {
		filter map[string]string
		want   []float32
	}{
		{
			filter: map[string]string{"regionCode": "us-east-1", "instanceType": "t3.medium"},
			want:   []float32{0.08},
		},
		{
			filter: map[string]string{"regionCode": "us-east-1", "instanceType": "t3.medium", "TenancyCode": "testing"},
			want:   []float32{},
		},
		{
			filter: map[string]string{"regionCode": "us-east-1", "instanceType": "t3.large"},
			want:   []float32{0.1},
		},
		{
			filter: map[string]string{"regionCode": "us-east-1"},
			want:   []float32{0.1, 0.08},
		},
	}

	data, err := os.Open("testdata/test_prices.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()

	pricer, err := NewCSVPricer(logr.Discard(), data)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range queryTests {
		t.Run(fmt.Sprintf("query %s", tt.filter), func(t *testing.T) {
			results, err := pricer.ListPrices(context.TODO(), "AmazonEC2", "USD", tt.filter)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, results, compareFloat32); diff != "" {
				t.Fatalf("failed to get results:\n%s", diff)
			}
		})
	}
}

func TestCSVPricer_ListPrices_insensitive_parsing(t *testing.T) {
	queryTests := []struct {
		filter map[string]string
		want   []float32
	}{
		{
			filter: map[string]string{"regionCode": "us-east-1", "instanceType": "t3.medium"},
			want:   []float32{0.08},
		},
		{
			filter: map[string]string{"regionCode": "us-east-1", "instanceType": "t3.medium", "TenancyCode": "testing"},
			want:   []float32{},
		},
		{
			filter: map[string]string{"regionCode": "us-east-1", "instanceType": "t3.large"},
			want:   []float32{0.1},
		},
		{
			filter: map[string]string{"regionCode": "us-east-1"},
			want:   []float32{0.1, 0.08},
		},
	}

	pricer, err := NewCSVPricerFromFile(logr.Discard(), "testdata/test_prices_case_insensitive.csv")
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range queryTests {
		t.Run(fmt.Sprintf("query %s", tt.filter), func(t *testing.T) {
			results, err := pricer.ListPrices(context.TODO(), "AmazonEC2", "USD", tt.filter)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, results, compareFloat32); diff != "" {
				t.Fatalf("failed to get results:\n%s", diff)
			}
		})
	}
}

func TestCSVPricer_ListPrices_logging(t *testing.T) {
	pricingTests := []struct {
		name    string
		data    string
		wantOut string
	}{
		{
			name:    "invalid price",
			data:    "serviceCode,currency,regionCode,instanceType,price\nAmazonEC2,USD,us-east-1,t3.large,1x\nAmazonEC2,USD,us-east-1,t3.medium,0.08\n",
			wantOut: "parsing \"1x\": invalid syntax failed to parse pricing data",
		},
	}

	for _, tt := range pricingTests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.Buffer{}
			data := strings.NewReader(tt.data)
			pricer, err := NewCSVPricer(buflogr.NewWithBuffer(&buf), data)
			if err != nil {
				t.Fatal(err)
			}
			_, err = pricer.ListPrices(context.TODO(), "AmazonEC2", "USD", map[string]string{"regionCode": "us-east-1", "instanceType": "t3.large"})
			if err != nil {
				t.Fatal(err)
			}

			if msg := buf.String(); !strings.Contains(msg, tt.wantOut) {
				t.Fatalf("got output %q, want %q", msg, tt.wantOut)
			}
		})
	}
}
