package estimation

import (
	"context"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/go-logr/logr"
)

func TestAWSPricer_ListPrices(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	skipIfNoCreds(t, cfg)

	svc := pricing.NewFromConfig(cfg)
	filters1 := map[string]string{
		"operatingSystem": "Linux",
		"regionCode":      "us-east-1",
		"instanceType":    "t3.large",
	}

	p := NewAWSPricer(logr.Discard(), svc)
	prices1, err := p.ListPrices(context.TODO(), "AmazonEC2", "USD", filters1)
	if err != nil {
		t.Fatal(err)
	}

	filters2 := map[string]string{
		"operatingSystem": "Linux",
		"regionCode":      "us-east-1",
		"instanceType":    "t3.large",
		"tenancy":         "Dedicated",
		"capacitystatus":  "UnusedCapacityReservation",
		"operation":       "RunInstances",
	}

	prices2, err := p.ListPrices(context.TODO(), "AmazonEC2", "USD", filters2)
	if err != nil {
		t.Fatal(err)
	}

	// As this is really talking to AWS, it's non-trivial to test the results.
	// prices2 is from a tighter set of filters, so it should return fewer
	// results than prices1
	if !(len(prices2) < len(prices1)) {
		t.Fatalf("ListPrices(%v) got %d prices, but ListPrices(%v) got %d prices", filters1, len(filters1), filters2, len(filters2))
	}
}

func TestPricing_ranged_result(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	skipIfNoCreds(t, cfg)

	svc := pricing.NewFromConfig(cfg)
	filters := map[string]string{
		"operatingSystem": "Linux",
		"regionCode":      "us-east-1",
		"instanceType":    "t3.large",
		"tenancy":         "Dedicated",
		"capacitystatus":  "UnusedCapacityReservation",
	}

	p := NewAWSPricer(logr.Discard(), svc)

	prices, err := p.ListPrices(context.TODO(), "AmazonEC2", "USD", filters)
	if err != nil {
		t.Fatal(err)
	}

	// As this is really talking to AWS, it's non-trivial to test the results.
	if l := len(prices); l != 2 {
		t.Fatalf("got %v prices, want %v", l, 2)
	}
}

func skipIfNoCreds(t *testing.T, cfg aws.Config) {
	t.Helper()
	_, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		t.Skip("could not load AWS credentials for pricing API")
	}
}
