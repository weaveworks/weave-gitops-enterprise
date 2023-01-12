package estimation

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
)

func TestCSVPricer_ListPrices(t *testing.T) {
	queryTests := []struct {
		filter map[string]string
		want   []float32
	}{
		{
			filter: map[string]string{"locationCode": "us-east-1", "instanceType": "t3.medium"},
			want:   []float32{0.08},
		},
		{
			filter: map[string]string{"locationCode": "us-east-1", "instanceType": "t3.medium", "TenancyCode": "testing"},
			want:   []float32{},
		},
		{
			filter: map[string]string{"locationCode": "us-east-1", "instanceType": "t3.large"},
			want:   []float32{0.1},
		},
		{
			filter: map[string]string{"locationCode": "us-east-1"},
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
