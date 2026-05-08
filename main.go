// Package broccli is meant to make handling command line interface easier.
// Define commands with arguments, flags, attach a handler to it and package will do all the parsing.
package broccli

// Command param type.
const (
	_ = iota * 1
	// ParamFlag sets param to be a command flag.
	ParamFlag
	// ParamFlag sets param to be a command arg.
	ParamArg
	// ParamFlag sets param to be an environment variable.
	ParamEnvVar
)

// Value types.
const (
	// TypeString requires param to be a string.
	TypeString = iota * 1
	// TypeBool requires param to be a boolean.
	TypeBool
	// TypeInt requires param to be an integer.
	TypeInt
	// TypeFloat requires param to be a float.
	TypeFloat
	// TypeAlphanumeric requires param to contain numbers and latin letters only.
	TypeAlphanumeric
	// TypePathFile requires param to be a path to a file.
	TypePathFile
)

// Validation.
const (
	_ = 1 << iota
	// IsRequired means that the value is required.
	IsRequired
	// IsExistent is used with TypePathFile and requires file to exist.
	IsExistent
	// IsNotExistent is used with TypePathFile and requires file not to exist.
	IsNotExistent
	// IsDirectory is used with TypePathFile and requires file to be a directory.
	IsDirectory
	// IsRegularFile is used with TypePathFile and requires file to be a regular file.
	IsRegularFile
	// IsValidJSON is used with TypeString or TypePathFile with RegularFile to check if the contents are a valid JSON.
	IsValidJSON

	// AllowDots can be used only with TypeAlphanumeric and additionally allows flag to have dots.
	AllowDots
	// AllowUnderscore can be used only with TypeAlphanumeric and additionally allows flag to have underscore chars.
	AllowUnderscore
	// AllowHyphen can be used only with TypeAlphanumeric and additionally allows flag to have hyphen chars.
	AllowHyphen

	// AllowMultipleValues allows param to have more than one value separated by comma by default.
	// For example: AllowMany with TypeInt allows values like: 123 or 123,455,666 or 12,222
	// AllowMany works only with TypeInt, TypeFloat and TypeAlphanumeric.
	AllowMultipleValues
	// SeparatorColon works with AllowMultipleValues and sets colon to be the value separator, instead of colon.
	SeparatorColon
	// SeparatorSemiColon works with AllowMultipleValues and sets semi-colon to be the value separator.
	SeparatorSemiColon
)

const (
	tabWriterMinWidth           = 8
	tabWriterMinWidthForCommand = 10
	tabWriterTabWidth           = 8
	tabWriterPadding            = 8
	tabWriterPadChar            = '\t'
)

const (
	maxArgs = 10
)
