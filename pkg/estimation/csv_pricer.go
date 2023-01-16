package estimation

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
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

// NewCSVPricerFromFile parses the CSV pricing data by opening the provided
// filename.
func NewCSVPricerFromFile(l logr.Logger, filename string) (*CSVPricer, error) {
	data, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open csv pricing data: %w", err)
	}
	defer data.Close()

	return NewCSVPricer(l, data)
}

// NewCSVPricer creates and returns a new CSVPricer ready for use.
func NewCSVPricer(l logr.Logger, in io.Reader) (*CSVPricer, error) {
	reader := csv.NewReader(in)

	count := 0
	headers := []string{}
	records := []caseInsensitiveMap{}
	for {
		row, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("failed to read pricing CSV: %w", err)
		}

		count++

		if count == 1 {
			headers = row
			continue
		}

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
			p.log.Error(err, "failed to parse pricing data", "row", row)
			continue
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
			matches++
		}
	}

	return matches
}
