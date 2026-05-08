package broccli

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"
)

// Command represent a command which has a name (used in args when calling app), usage, a handler that is called.
// Such command can have flags and arguments.  In addition to that, required environment variables can be set.
type Command struct {
	name      string
	usage     string
	flags     map[string]*param
	args      map[string]*param
	argsOrder []string
	argsIdx   int
	env       map[string]*param
	handler   func(context.Context, *Broccli) int
	options   commandOptions
}

// Flag adds a flag to a command and returns a pointer to Param instance.
// Method requires name (eg. 'data' for '--data', alias (eg. 'd' for '-d'), placeholder for the value displayed on the
// 'help' screen, usage, type of the value and additional validation that is set up with bit flags, eg. IsRequired
// or AllowMultipleValues.  If no additional flags are required, 0 should be used.
func (c *Command) Flag(
	name, alias, valuePlaceholder, usage string,
	types, flags int64,
	opts ...ParamOption,
) {
	if c.flags == nil {
		c.flags = map[string]*param{}
	}

	c.flags[name] = &param{
		name:             name,
		alias:            alias,
		usage:            usage,
		valuePlaceholder: valuePlaceholder,
		valueType:        types,
		flags:            flags,
		options:          paramOptions{},
	}
	for _, o := range opts {
		o(&(c.flags[name].options))
	}
}

// Arg adds an argument to a command and returns a pointer to Param instance.  It is the same as adding flag except
// it does not have an alias.
func (c *Command) Arg(
	name, valuePlaceholder, usage string,
	types, flags int64,
	opts ...ParamOption,
) {
	if c.argsIdx > maxArgs-1 {
		log.Fatal("Only 10 arguments are allowed")
	}

	if c.args == nil {
		c.args = map[string]*param{}
	}

	c.args[name] = &param{
		name:             name,
		usage:            usage,
		valuePlaceholder: valuePlaceholder,
		valueType:        types,
		flags:            flags,
		options:          paramOptions{},
	}
	if c.argsOrder == nil {
		c.argsOrder = make([]string, maxArgs)
	}

	c.argsOrder[c.argsIdx] = name

	c.argsIdx++

	for _, opt := range opts {
		opt(&(c.args[name].options))
	}
}

// Env adds a required environment variable to a command and returns a pointer to Param.  It's arguments are very
// similar to ones in previous AddArg and AddFlag methods.
func (c *Command) Env(name, usage string, types, flags int64, _ ...ParamOption) {
	if c.env == nil {
		c.env = map[string]*param{}
	}

	c.env[name] = &param{
		name:      name,
		usage:     usage,
		valueType: types,
		flags:     flags,
		options:   paramOptions{},
	}
}

func (c *Command) sortedArgs() []string {
	argNamesSorted := make([]string, c.argsIdx)
	idx := 0

	// required args first
	for argIdx := range c.argsIdx {
		argOrderedName := c.argsOrder[argIdx]

		arg := c.args[argOrderedName]

		if arg.flags&IsRequired > 0 {
			argNamesSorted[idx] = argOrderedName
			idx++
		}
	}

	// optional args
	for argIdx := range c.argsIdx {
		argOrderedName := c.argsOrder[argIdx]

		arg := c.args[argOrderedName]

		if arg.flags&IsRequired == 0 {
			argNamesSorted[idx] = argOrderedName
			idx++
		}
	}

	return argNamesSorted
}

func (c *Command) sortedFlags() []string {
	flagNames := reflect.ValueOf(c.flags).MapKeys()

	flagNamesSorted := make([]string, len(flagNames))

	for i, flagName := range flagNames {
		flagNamesSorted[i] = flagName.String()
	}

	sort.Strings(flagNamesSorted)

	return flagNamesSorted
}

func (c *Command) sortedEnv() []string {
	envNames := reflect.ValueOf(c.env).MapKeys()

	envNamesSorted := make([]string, len(envNames))

	for i, envName := range envNames {
		envNamesSorted[i] = envName.String()
	}

	sort.Strings(envNamesSorted)

	return envNamesSorted
}

// PrintHelp prints command usage information to stdout file.
func (c *Command) printHelp() {
	var helpMessage strings.Builder

	_, _ = fmt.Fprintf(&helpMessage, "\nUsage:  %s %s [FLAGS]%s\n\n", path.Base(os.Args[0]), c.name,
		c.argsHelpLine())
	_, _ = fmt.Fprintf(&helpMessage, "%s\n", c.usage)

	if len(c.env) > 0 {
		_, _ = fmt.Fprintf(&helpMessage, "\nRequired environment variables:\n")

		tabFormatter := new(tabwriter.Writer)
		tabFormatter.Init(
			&helpMessage,
			tabWriterMinWidth,
			tabWriterTabWidth,
			tabWriterPadding,
			tabWriterPadChar,
			0,
		)

		for _, envName := range c.sortedEnv() {
			_, _ = fmt.Fprintf(tabFormatter, "%s\t%s\n", envName, c.env[envName].usage)
		}

		_ = tabFormatter.Flush()
	}

	tabFormatter := new(tabwriter.Writer)
	tabFormatter.Init(
		&helpMessage,
		tabWriterMinWidth,
		tabWriterTabWidth,
		tabWriterPadding,
		tabWriterPadChar,
		0,
	)

	var requiredFlags string
	var optionalFlags string

	for _, flagName := range c.sortedFlags() {
		flag := c.flags[flagName]
		if flag.flags&IsRequired > 0 {
			requiredFlags += flag.helpLine()
		} else {
			optionalFlags += flag.helpLine()
		}
	}

	if requiredFlags != "" {
		_, _ = fmt.Fprintf(tabFormatter, "\nRequired flags: \n")
		_, _ = fmt.Fprintf(tabFormatter, "%s", requiredFlags)
		_ = tabFormatter.Flush()
	}

	if optionalFlags != "" {
		_, _ = fmt.Fprintf(tabFormatter, "\nOptional flags: \n")
		_, _ = fmt.Fprintf(tabFormatter, "%s", optionalFlags)
		_ = tabFormatter.Flush()
	}

	_, err := fmt.Fprint(os.Stdout, helpMessage.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Unable to build help message")
	}
}

func (c *Command) argsHelpLine() string {
	argsRequired := ""
	argsOptional := ""

	if c.argsIdx > 0 {
		for argIdx := range c.argsIdx {
			flagOrderedName := c.argsOrder[argIdx]

			arg := c.args[flagOrderedName]
			if arg.flags&IsRequired > 0 {
				argsRequired += " " + arg.valuePlaceholder
			} else {
				argsOptional += " [" + arg.valuePlaceholder + "]"
			}
		}
	}

	return argsRequired + argsOptional
}
