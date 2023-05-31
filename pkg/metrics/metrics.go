package metrics

import "github.com/go-logr/logr"

type Options struct {
	Enabled       bool
	ServerAddress string
	Log           logr.Logger
}
