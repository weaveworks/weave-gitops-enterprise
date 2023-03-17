package objectscollector

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/go-logr/logr"
)

var log logr.Logger
var g *WithT

func TestObjectsCollector(t *testing.T) {
	// TODD: We need to test the objects collector with a "running" cluster using the envtest library
}
