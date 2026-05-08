package broccli

type commandOptions struct {
	onPostValidation func(c *Command) error
}

// CommandOption defines an optional configuration function for commands, intended for specific use cases.
// It should not be created manually; use one of the predefined functions below.
type CommandOption func(opts *commandOptions)

// OnPostValidation attaches a function that is called once args, flags and env vars are validated.
func OnPostValidation(fn func(c *Command) error) CommandOption {
	return func(opts *commandOptions) {
		opts.onPostValidation = fn
	}
}
