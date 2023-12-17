package steps

// ModesConfig contains the different modes that the bootstrap supports.
type ModesConfig struct {
	// Silent instruct to do best effort to take decisions based on existing information.
	Silent bool
	// Export instruct to generate resources but to do not mutate any system but to write resources to stdout.
	Export bool
}
