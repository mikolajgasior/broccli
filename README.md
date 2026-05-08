# broccli

## ⚠️ Project Status: Unmaintained

This project is no longer actively maintained. While it may still work for existing users, no new features or bug fixes will be added.

Feel free to fork this repository if you'd like to continue development.

[broccli](broccli.png "broccli")

The `broccli/v3` package simplifies command line interface management. It allows you to define commands complete with arguments and flags, and attach handlers to them. The package handles all the parsing automatically.

:warning: `v3` is not compatible with `v2`. The latest version requires a context in handlers and when calling `Run` method. Also
methods for adding commands, args etc. have been shortened to `Command`, `Arg`, `Flag` and `Env`.


## Table of Contents

* [Sample code](#sample-code)
* [Structs explained](#structs-explained)
  * [CLI](#cli)
  * [Commands](#commands)
  * [Flags and Arguments](#flags-and-arguments)
  * [Environment variables to check](#environment-variables-to-check)
  * [Accessing flag and arg values](#accessing-flag-and-arg-values)
* [Features + Roadmap](#features)

## Sample code
Example dummy application can be found in `cmd/example1` directory.

However, the following code snippet from another tiny project shows how the module can be used.

Import the code with the following URL:
```go
import "github.com/mikolajgasior/broccli/v3"
```

```go
// create new CLI object
cli := broccli.NewBroccli("snakey-letters", "Classic snake but with letters and words!", "")

// add a command and attach a function to it
cmd := cli.Command("start", "Starts the game", startHandler)
// add a flag to the command
cmd.Flag("words", "f", "", 
    "Text file with wordlist", 
    // should be a path to a file
    broccli.TypePathFile,
    // flag is required (cannot be empty) and path must exist
    broccli.IsExistent|broccli.IsRequired,
)

// add another command to print version number
_ = cli.Command("version", "Shows version", versionHandler)

// run cli
os.Exit(cli.Run(context.Background()))

// handlers for each command
func startHandler(ctx context.Context, c *broccli.CLI) int {
	fmt.Fprint(os.Stdout, "Starting with file %s...", c.Flag("words"))
	return 0
}

func versionHandler(ctx context.Context, c *broccli.CLI) int {
    fmt.Fprintf(os.Stdout, VERSION+"\n")
    return 0
}
```

## Structs explained
### Broccli
The main `Broccli` object has three arguments such as name, usage and author. These guys are displayed when syntax is printed out.

### Commands
Method `AddCmd` creates a new command which has the following properties.

* a `name`, used to call it
* short `usage` - few words to display what the command does on the syntax screen
* `handler` function that is executed when the command is called and all its flags and argument are valid

#### Command Options
Optionally, after command flags and arguments are successfully validated, and just before the execution of `handler`, additional code (func) can be executed. This can be passed as a last argument.

```go
cmd := cli.Command("start", "Starts the game", startHandler, 
    broccli.OnPostValidation(func(c *broccli.Cmd) {
        // do something, even with the command
    }),
)
```

See `cmd_options.go` for all available options.

### Flags and Arguments
Each command can have arguments and flags, as shown below.

```txt
program some-command -f flag1 -g flag2 ARGUMENT1 ARGUMENT2
```

To setup a flag in a command, method `AddFlag` is used. It takes the following arguments:

* `name` and `alias` that are used to call the flag (eg. `--help` and `-h`, without the hyphens in the func args)
* `valuePlaceholder`, a placeholder that is printed out on the syntax screen, eg. in `-f PATH_TO_FILE` it is the `PATH_TO_FILE`
* `usage` - few words telling what the command does (syntax screen again)
* `types`, an int64 value that defines the value type, currently one of `TypeString`, `TypeBool`, `TypeInt`, `TypeFloat`, `TypeAlphanumeric` or `TypePathFile` (see `flags.go` for more information)
* `flags`, an int64 value containing validation requirement, eg. `IsRequired|IsExistent|IsDirectory` could be used with `TypePathFile` to require the flag to be a non-empty path to an existing directory (again, navigate to `flags.go` for more detailed information)

Optionally, a function can be attached to a boolean flag that is triggered when a flag is true. The motivation behind that was a use case when setting a certain flag to true would make another string flag required. However, it's not recommended to be used.

To add an argument for a command, method `Arg` shall be used. It has almost the same arguments, apart from the fact that `alias` is not there.

### Environment variables to check
Command may require environment variables. `Env` can be called to setup environment variables that should be verified before running the command. For example, a variable might need to contain a path to an existing regular file.

### Accessing flag and arg values
See sample code that does that below.

```go
func startHandler(ctx context.Context, c *broccli.CLI) int {
	fmt.Fprint(os.Stdout, "Starting with level %s...", c.Arg("level"))
    fmt.Fprint(os.Stdout, "Writing moves to file %s...", c.Flag("somefile"))
	return 0
}
```

`level` and `somefile` are `name`s of the argument (sometimes they are uppercase) and flag.

## Features
- [X] Flags and arguments support
- [X] Validation for basic value types such as integer, float, string, bool
- [X] Additional value types of alpha-numeric and file path
- [X] Validation for multiple values, separated with colon or semicolon, eg. `-t val1,val2,val3`
- [X] Check for file existence and its type (can be directory, regular file or other)
- [X] Post validation hook
- [X] Boolean flag on-true hook before validation
- [X] Handlers require context
