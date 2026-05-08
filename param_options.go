package broccli

type paramOptions struct {
	onTrue func(command *Command)
}

// ParamOption defines an optional configuration function for args and flags, intended for specific use cases.
// It should not be created manually; use one of the predefined functions below.
type ParamOption func(opts *paramOptions)

// OnTrue executes a specified function when boolean flag is true.
func OnTrue(fn func(command *Command)) ParamOption {
	return func(opts *paramOptions) {
		opts.onTrue = fn
	}
}
