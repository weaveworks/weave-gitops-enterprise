package test

import (
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/wks/cmd/event-writer/converter"
	"github.com/weaveworks/wks/common/database/models"
)

// Read file at path and return bytes
func readTestFile(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Create a models.Event object from an Event exported to a JSON file
func dbEventFromFile(t *testing.T, path string) (models.Event, error) {
	data, err := readTestFile(path)
	assert.NoError(t, err)
	event, err := converter.DeserializeJSONToEvent(data)
	assert.NoError(t, err)

	dbEvent, err := converter.ConvertEvent(*event)
	assert.NoError(t, err)
	return dbEvent, nil
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// StringWithCharset creates a random string given a length and charset
func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// RandomString creates a random string of a given length using the default charset
func RandomString(length int) string {
	return StringWithCharset(length, charset)
}
