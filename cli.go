package broccli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"
)

// Broccli is main CLI application definition.
// It has a name, description, author which are printed out to the screen in the usage syntax.
// Each CLI have commands (represented by Command).  Optionally, it is possible to require environment
// variables.
type Broccli struct {
	name        string
	usage       string
	author      string
	commands    map[string]*Command
	env         map[string]*param
	parsedFlags map[string]string
	parsedArgs  map[string]string
}

// NewBroccli returns pointer to a new Broccli instance.  Name, usage and author are displayed on the syntax screen.
func NewBroccli(name, usage, author string) *Broccli {
	cli := &Broccli{
		name:        name,
		usage:       usage,
		author:      author,
		commands:    map[string]*Command{},
		env:         map[string]*param{},
		parsedFlags: map[string]string{},
		parsedArgs:  map[string]string{},
	}

	return cli
}

// Command returns pointer to a new command with specified name, usage and handler.  Handler is a function that
// gets called when command is executed.
// Additionally, there is a set of options that can be passed as arguments.  Search for commandOption for more info.
func (c *Broccli) Command(
	name, usage string,
	handler func(ctx context.Context, cli *Broccli) int,
	opts ...CommandOption,
) *Command {
	c.commands[name] = &Command{
		name:    name,
		usage:   usage,
		flags:   map[string]*param{},
		args:    map[string]*param{},
		env:     map[string]*param{},
		handler: handler,
		options: commandOptions{},
	}
	for _, opt := range opts {
		opt(&(c.commands[name].options))
	}

	return c.commands[name]
}

// Env returns pointer to a new environment variable that is required to run every command.
// Method requires name, eg. MY_VAR, and usage.
func (c *Broccli) Env(name string, usage string) {
	c.env[name] = &param{
		name:    name,
		usage:   usage,
		flags:   IsRequired,
		options: paramOptions{},
	}
}

// Flag returns value of flag.
func (c *Broccli) Flag(name string) string {
	return c.parsedFlags[name]
}

// Arg returns value of arg.
func (c *Broccli) Arg(name string) string {
	return c.parsedArgs[name]
}

// Run parses the arguments, validates them and executes command handler.
// In case of invalid arguments, error is printed to stderr and 1 is returned.  Return value should be treated as exit
// code.
func (c *Broccli) Run(ctx context.Context) int {
	// display help, first arg is binary filename
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		c.printHelp()

		return 0
	}

	for _, commandName := range c.sortedCommands() {
		if commandName != os.Args[1] {
			continue
		}
		// display command help
		if len(os.Args) > 2 && (os.Args[2] == "-h" || os.Args[2] == "--help") {
			c.commands[commandName].printHelp()

			return 0
		}

		// check required environment variables
		if len(c.env) > 0 {
			for env, param := range c.env {
				envValue := os.Getenv(env)
				param.flags |= IsRequired

				err := param.validateValue(envValue)
				if err != nil {
					fmt.Fprintf(
						os.Stderr,
						"ERROR: %s %s: %s\n",
						c.getParamTypeName(ParamEnvVar),
						param.name,
						err.Error(),
					)
					c.printHelp()

					return 1
				}
			}
		}

		// parse and validate all the flags and args
		exitCode := c.parseFlags(c.commands[commandName])
		if exitCode > 0 {
			return exitCode
		}

		return c.commands[commandName].handler(ctx, c)
	}

	// command not found
	c.printInvalidCommand(os.Args[1])

	return 1
}

func (c *Broccli) sortedCommands() []string {
	commandNames := reflect.ValueOf(c.commands).MapKeys()

	commandNamesSorted := make([]string, len(commandNames))

	for i, cmd := range commandNames {
		commandNamesSorted[i] = cmd.String()
	}

	sort.Strings(commandNamesSorted)

	return commandNamesSorted
}

func (c *Broccli) sortedEnv() []string {
	envNames := reflect.ValueOf(c.env).MapKeys()

	envNamesSorted := make([]string, len(envNames))

	for i, ev := range envNames {
		envNamesSorted[i] = ev.String()
	}

	sort.Strings(envNamesSorted)

	return envNamesSorted
}

func (c *Broccli) printHelp() {
	var helpMessage strings.Builder

	_, _ = fmt.Fprintf(
		&helpMessage,
		"%s by %s\n%s\n\nUsage: %s COMMAND\n\n",
		c.name,
		c.author,
		c.usage,
		path.Base(os.Args[0]),
	)

	if len(c.env) > 0 {
		_, _ = fmt.Fprintf(&helpMessage, "Required environment variables:\n")

		tabFormatter := new(tabwriter.Writer)
		tabFormatter.Init(
			&helpMessage,
			tabWriterMinWidth,
			tabWriterTabWidth,
			tabWriterPadding,
			tabWriterPadChar,
			0,
		)

		for _, n := range c.sortedEnv() {
			_, _ = fmt.Fprintf(tabFormatter, "%s\t%s\n", n, c.env[n].usage)
		}

		_ = tabFormatter.Flush()
	}

	_, _ = fmt.Fprintf(&helpMessage, "Commands:\n")

	tabFormatter := new(tabwriter.Writer)
	tabFormatter.Init(
		&helpMessage,
		tabWriterMinWidthForCommand,
		tabWriterTabWidth,
		tabWriterPadding,
		tabWriterPadChar,
		0,
	)

	for _, commandName := range c.sortedCommands() {
		_, _ = fmt.Fprintf(&helpMessage, "  %s\t%s\n", commandName, c.commands[commandName].usage)
	}

	_ = tabFormatter.Flush()

	_, _ = fmt.Fprintf(
		&helpMessage,
		"\nRun '%s COMMAND --help' for command syntax.\n",
		path.Base(os.Args[0]),
	)

	_, err := fmt.Fprint(os.Stdout, helpMessage.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Unable to build help message")
	}
}

func (c *Broccli) printInvalidCommand(cmd string) {
	fmt.Fprintf(os.Stderr, "Invalid command: %s\n\n", cmd)
	c.printHelp()
}

// getFlagSetPtrs creates flagset instance, parses flags and returns list of pointers to results of parsing the flags.
func (c *Broccli) getFlagSetPtrs(
	cmd *Command,
) (map[string]interface{}, map[string]interface{}, []string) {
	fset := flag.NewFlagSet("flagset", flag.ContinueOnError)
	// nothing should come out of flagset
	fset.Usage = func() {}
	fset.SetOutput(io.Discard)

	flagNamePtrs := make(map[string]interface{})
	flagAliasPtrs := make(map[string]interface{})

	flagNamesSorted := cmd.sortedFlags()
	for _, flagName := range flagNamesSorted {
		flagInstance := cmd.flags[flagName]
		if flagInstance.valueType == TypeBool {
			flagNamePtrs[flagName] = fset.Bool(flagName, false, "")
			if flagInstance.alias != "" {
				flagAliasPtrs[flagInstance.alias] = fset.Bool(flagInstance.alias, false, "")
			}
		} else {
			flagNamePtrs[flagName] = fset.String(flagName, "", "")
			if flagInstance.alias != "" {
				flagAliasPtrs[flagInstance.alias] = fset.String(flagInstance.alias, "", "")
			}
		}
	}

	err := fset.Parse(os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Unable to parse flags: %s", err.Error())
	}

	return flagNamePtrs, flagAliasPtrs, fset.Args()
}

func (c *Broccli) checkEnv(cmd *Command) int {
	if len(cmd.env) == 0 {
		return 0
	}

	for envName, envVar := range cmd.env {
		envValue := os.Getenv(envName)
		envVar.flags |= IsRequired

		err := envVar.validateValue(envValue)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"ERROR: %s %s: %s\n",
				c.getParamTypeName(ParamEnvVar),
				envVar.name,
				err.Error(),
			)
			cmd.printHelp()

			return 1
		}
	}

	return 0
}

func (c *Broccli) processOnTrue(
	cmd *Command,
	flagNames []string,
	nflags map[string]interface{},
	aflags map[string]interface{},
) {
	for _, name := range flagNames {
		if cmd.flags[name].valueType != TypeBool {
			continue
		}

		if cmd.flags[name].options.onTrue == nil {
			continue
		}

		// OnTrue is called when a flag is true
		//nolint:forcetypeassert
		if *(nflags[name]).(*bool) || *(aflags[cmd.flags[name].alias]).(*bool) {
			cmd.flags[name].options.onTrue(cmd)
		}
	}
}

func (c *Broccli) processFlags(
	cmd *Command,
	flagNames []string,
	nflags map[string]interface{},
	aflags map[string]interface{},
) int {
	for _, name := range flagNames {
		flag := cmd.flags[name]

		if flag.valueType == TypeBool {
			c.parsedFlags[name] = "false"
			//nolint:forcetypeassert
			if *(nflags[name]).(*bool) || (cmd.flags[name].alias != "" && *(aflags[cmd.flags[name].alias]).(*bool)) {
				c.parsedFlags[name] = "true"
			}

			continue
		}

		//nolint:forcetypeassert
		aliasValue := ""
		if flag.alias != "" {
			aliasValue = *(aflags[flag.alias]).(*string)
		}
		//nolint:forcetypeassert
		nameValue := *(nflags[name]).(*string)

		if nameValue != "" && aliasValue != "" {
			fmt.Fprintf(os.Stderr, "ERROR: Both -%s and --%s passed", flag.alias, flag.name)

			return 1
		}

		flagValue := aliasValue
		if nameValue != "" {
			flagValue = nameValue
		}

		err := flag.validateValue(flagValue)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"ERROR: %s %s: %s\n",
				c.getParamTypeName(ParamFlag),
				name,
				err.Error(),
			)
			cmd.printHelp()

			return 1
		}

		c.parsedFlags[name] = flagValue
	}

	return 0
}

func (c *Broccli) processArgs(cmd *Command, argNamesSorted []string, args []string) int {
	for argIdx, argName := range argNamesSorted {
		argValue := ""
		if len(args) >= argIdx+1 {
			argValue = args[argIdx]
		}

		err := cmd.args[argName].validateValue(argValue)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"ERROR: %s %s: %s\n",
				c.getParamTypeName(ParamArg),
				cmd.args[argName].valuePlaceholder,
				err.Error(),
			)
			cmd.printHelp()

			return 1
		}

		c.parsedArgs[argName] = argValue
	}

	return 0
}

func (c *Broccli) processOnPostValidation(cmd *Command) int {
	if cmd.options.onPostValidation == nil {
		return 0
	}

	err := cmd.options.onPostValidation(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		cmd.printHelp()

		return 1
	}

	return 0
}

func (c *Broccli) parseFlags(cmd *Command) int {
	// check required environment variables
	if exitCode := c.checkEnv(cmd); exitCode != 0 {
		return exitCode
	}

	flags := cmd.sortedFlags()
	flagNamePtrs, flagAliasPtrs, args := c.getFlagSetPtrs(cmd)

	// Loop through boolean flags and execute onTrue() hook if exists.  That function might be used to change behaviour
	// of other flags, eg. when -e is added, another flag or argument might become required (or obsolete).
	// Bool fields will be parsed out in this loop so no reason to process them again in the next one.
	c.processOnTrue(cmd, flags, flagNamePtrs, flagAliasPtrs)

	if exitCode := c.processFlags(cmd, flags, flagNamePtrs, flagAliasPtrs); exitCode != 0 {
		return exitCode
	}

	argsNamesSorted := cmd.sortedArgs()
	if exitCode := c.processArgs(cmd, argsNamesSorted, args); exitCode != 0 {
		return exitCode
	}

	if exitCode := c.processOnPostValidation(cmd); exitCode != 0 {
		return exitCode
	}

	return 0
}

func (c *Broccli) getParamTypeName(t int8) string {
	if t == ParamArg {
		return "Argument"
	}

	if t == ParamEnvVar {
		return "Env var"
	}

	return "Flag"
}
