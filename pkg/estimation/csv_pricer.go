package estimation

import (
	"context"
	"encoding/csv"
	"io"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
)

type caseInsensitiveMap map[string]string

func (c caseInsensitiveMap) set(key, value string) {
	c[strings.ToLower(key)] = value
}

func (c caseInsensitiveMap) get(key string) string {
	return c[strings.ToLower(key)]
}

// NewCSVPricer creates and returns a new CSVPricer ready for use.
func NewCSVPricer(l logr.Logger, in io.Reader) (*CSVPricer, error) {
	reader := csv.NewReader(in)
	rawRecords, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	headers := rawRecords[0]

	records := []caseInsensitiveMap{}
	for _, row := range rawRecords[1:] {
		record := caseInsensitiveMap{}
		for i := range headers {
			record.set(headers[i], row[i])
		}
		records = append(records, record)
	}

	return &CSVPricer{
		Records: records,
		log:     l,
	}, nil
}

// CSVPricer is an implementation of the Pricer that can use a CSV file as the
// source of pricing data.
type CSVPricer struct {
	log     logr.Logger
	Records []caseInsensitiveMap
}

func (p *CSVPricer) ListPrices(ctx context.Context, service, currency string, filter map[string]string) ([]float32, error) {
	queryFilter := map[string]string{}
	for k, v := range filter {
		queryFilter[k] = v
	}
	queryFilter["serviceCode"] = service
	queryFilter["currency"] = currency

	matchingRows := []caseInsensitiveMap{}
	for i := range p.Records {
		if count := matchingKeyCount(queryFilter, p.Records[i]); count == len(queryFilter) {
			matchingRows = append(matchingRows, p.Records[i])
		}
	}

	results := []float32{}
	for _, row := range matchingRows {
		price, err := strconv.ParseFloat(row.get("price"), 32)
		if err != nil {
			continue // what to do?
		}
		results = append(results, float32(price))
	}

	return results, nil
}

// returns the number of keys in the row that match the filter.
func matchingKeyCount(filter map[string]string, row caseInsensitiveMap) int {
	matches := 0
	for k, v := range filter {
		if row.get(k) == v {
			matches += 1
		}
	}

	return matches
}
