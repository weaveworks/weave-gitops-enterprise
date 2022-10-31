package estimation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/go-logr/logr"
)

var _ Pricer = (*AWSPricer)(nil)

const invalidPrice float32 = -1

// NewAWSPricer creates and returns a new AWSPricer ready for use.
func NewAWSPricer(l logr.Logger, c pricing.GetProductsAPIClient) *AWSPricer {
	return &AWSPricer{client: c, log: l}
}

// AWSPricer is an implementation of the Pricer that can use the AWS
// Product API to get the price for a Product.
type AWSPricer struct {
	client pricing.GetProductsAPIClient
	log    logr.Logger
}

// ListPrices implements the Pricer interface by querying AWS.
func (a *AWSPricer) ListPrices(ctx context.Context, service, currency string, filters map[string]string) ([]float32, error) {
	paginator := pricing.NewGetProductsPaginator(a.client, &pricing.GetProductsInput{
		ServiceCode: aws.String(service),
		Filters:     filtersFromMap(filters),
	})

	pages := 0
	items := 0
	prices := []float32{}
	for paginator.HasMorePages() {
		a.log.V(4).Info("loading pricing page", "count", pages)
		resp, err := paginator.NextPage(context.TODO())
		if err != nil {
			a.log.Error(err, "failed to list products")
			return nil, err
		}
		for _, v := range resp.PriceList {
			items += 1
			price, err := parseUnitPrice(a.log, v, currency)
			if err != nil {
				a.log.Error(err, "failed to parse price data")
				return nil, err
			}
			if price > invalidPrice {
				prices = append(prices, price)
			}
		}
		pages += 1
	}

	return prices, nil
}

// productResult is parsed from the response to GetProducts.
type productResult struct {
	ProductFamily string                                       `json:"productFamily"`
	ServiceCode   string                                       `json:"serviceCode"`
	Terms         map[string]map[string]map[string]interface{} `json:"terms"`
	Product       struct {
		Attributes map[string]string `json:"attributes"`
	} `json:"product"`
}

func parseUnitPrice(log logr.Logger, body, currency string) (float32, error) {
	var decoded productResult
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		return invalidPrice, fmt.Errorf("failed to unmarshal product response: %w", err)
	}

	ondemandTerms, ok := decoded.Terms["OnDemand"]
	if !ok {
		return invalidPrice, nil
	}

	for _, term := range ondemandTerms {
		dim, ok := term["priceDimensions"]
		if !ok {
			return invalidPrice, errors.New("unable to find on-demand pricing dimensions")
		}

		for _, v := range dim.(map[string]interface{}) {
			pricePerUnit, ok := v.(map[string]interface{})["pricePerUnit"]
			if !ok {
				return invalidPrice, errors.New("unable to find on-demand pricing dimensions")
			}
			log.V(4).Info("pricing", "pricePerUnit", pricePerUnit)
			amount, err := priceFromPricePerUnitMap(pricePerUnit.(map[string]interface{}), currency)
			if err != nil {
				return invalidPrice, err
			}
			return amount, nil
		}
	}

	return invalidPrice, errors.New("failed to parse unit price data")
}

func priceFromPricePerUnitMap(pricePerUnit map[string]any, currency string) (float32, error) {
	keys := []string{}
	for k := range pricePerUnit {
		keys = append(keys, k)
	}
	if len(keys) == 1 {
		currency := keys[0]
		rawPrice := pricePerUnit[currency]
		price, err := strconv.ParseFloat(rawPrice.(string), 64)
		if err != nil {
			return invalidPrice, fmt.Errorf("unable to parse unit price %q: %w", rawPrice, err)
		}
		return float32(price), nil
	}

	return invalidPrice, errors.New("unable to parse the price")
}
