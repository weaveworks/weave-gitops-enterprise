package configuration

import "time"

type RetentionPolicy time.Duration

const (
	NoRetentionPolicy RetentionPolicy = 0
)
