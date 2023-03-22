package main

import (
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
)

const (
	defaultTimeout  = time.Second * 10
	defaultInterval = time.Second
)

type serverKey struct{}

var log logr.Logger
var g *WithT

func TestServer(t *testing.T) {
	g = NewWithT(t)
	g.SetDefaultEventuallyTimeout(defaultTimeout)
	g.SetDefaultEventuallyPollingInterval(defaultInterval)

	log = testr.New(t)

	main()

}
