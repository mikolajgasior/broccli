package broccli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var (
	errFileNotExist       = errors.New("file does not exist")
	errFileInfo           = errors.New("file cannot be opened for stat info")
	errFileExist          = errors.New("file already exists")
	errFileNotRegularFile = errors.New("file is not a regular file")
	errFileNotDirectory   = errors.New("file is not a directory")
	errFileOpen           = errors.New("file cannot be opened")
	errFileNotValidJSON   = errors.New("file is not a valid JSON")
	errParamValueMissing  = errors.New("param value missing")
	errParamValueInvalid  = errors.New("param value invalid")
	errParamTypeInvalid   = errors.New("param type invalid")
)

func errFileNotExistInPath(path string) error {
	return fmt.Errorf("%w: %s", errFileNotExist, path)
}

func errFileInfoInPath(path string) error {
	return fmt.Errorf("%w: %s", errFileInfo, path)
}

func errFileExistInPath(path string) error {
	return fmt.Errorf("%w: %s", errFileExist, path)
}

func errFileNotRegularFileInPath(path string) error {
	return fmt.Errorf("%w: %s", errFileNotRegularFile, path)
}

func errFileNotDirectoryInPath(path string) error {
	return fmt.Errorf("%w: %s", errFileNotDirectory, path)
}

func errFileOpenInPath(reason string, path string) error {
	return fmt.Errorf("%s %w: %s", reason, errFileOpen, path)
}

func errFileNotValidJSONInPath(path string) error {
	return fmt.Errorf("%w: %s", errFileNotValidJSON, path)
}

// param represends a value and it is used for flags, args and environment variables.
// It has a name, alias, usage, value that is shown when printing help, specific type (eg. TypeBool or TypeInt),
// If more than one value shoud be allowed, eg. '1,2,3' means "multiple integers" and the separator here is ','.
// Additional characters are used with type of TypeAlphanumeric to allow dots, underscore etc.  Hence, the value of that
// arg could be '._-'.
type param struct {
	name             string
	alias            string
	valuePlaceholder string
	usage            string
	valueType        int64
	flags            int64
	options          paramOptions
}

// helpLine returns param usage info that is used when printing help.
func (p *param) helpLine() string {
	usageLine := " "
	if p.alias == "" {
		usageLine += " \t"
	} else {
		usageLine += fmt.Sprintf(" -%s,\t", p.alias)
	}

	usageLine += fmt.Sprintf(" --%s %s \t%s\n", p.name, p.valuePlaceholder, p.usage)

	return usageLine
}

func (p *param) validatePathFile(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if p.flags&IsExistent > 0 {
				return errFileNotExistInPath(path)
			}

			return nil
		}

		return errFileInfoInPath(path)
	}

	if p.flags&IsNotExistent > 0 {
		return errFileExistInPath(path)
	}

	if !fileInfo.Mode().IsRegular() && (p.flags&IsRegularFile > 0) {
		return errFileNotRegularFileInPath(path)
	}

	if !fileInfo.Mode().IsDir() && (p.flags&IsDirectory > 0) {
		return errFileNotDirectoryInPath(path)
	}

	if (p.flags&IsRegularFile > 0) && (p.flags&IsValidJSON > 0) {
		dat, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return errFileOpenInPath("validate json", path)
		}

		if !json.Valid(dat) {
			return errFileNotValidJSONInPath(path)
		}
	}

	return nil
}

//nolint:funlen
func (p *param) validateValue(paramValue string) error {
	// empty, for every time except bool
	if p.valueType != TypeBool && (p.flags&IsRequired > 0) && paramValue == "" {
		return errParamValueMissing
	}

	// string does not need any additional checks apart from the above one
	if p.valueType == TypeString {
		return nil
	}

	// if param is not required or not empty
	if p.flags&IsRequired <= 0 && paramValue == "" {
		return nil
	}

	// if flag is a file (regular file, directory, ...)
	if p.valueType == TypePathFile {
		errValidatePathFile := p.validatePathFile(paramValue)
		if errValidatePathFile != nil {
			return fmt.Errorf("file path validation failed: %w", errValidatePathFile)
		}

		return nil
	}

	// int, float, alphanumeric - single or many, separated by various chars
	var (
		reType  string
		reValue string
	)
	// set regexp part just for the type (eg. int, float, anum)

	switch p.valueType {
	case TypeInt:
		reType = "[0-9]+"
	case TypeFloat:
		reType = "[0-9]{1,16}\\.[0-9]{1,16}"
	case TypeAlphanumeric:
		reExtraChars := ""
		if p.flags&AllowUnderscore > 0 {
			reExtraChars += "_"
		}

		if p.flags&AllowDots > 0 {
			reExtraChars += "\\."
		}

		if p.flags&AllowHyphen > 0 {
			reExtraChars += "\\-"
		}

		reType = fmt.Sprintf("[0-9a-zA-Z%s]+", reExtraChars)
	default:
		return errParamTypeInvalid
	}

	// create the final regexp depending on if single or many values are allowed
	if p.flags&AllowMultipleValues > 0 {
		var delimeter string
		//nolint:gocritic
		if p.flags&SeparatorColon > 0 {
			delimeter = ":"
		} else if p.flags&SeparatorSemiColon > 0 {
			delimeter = ";"
		} else {
			delimeter = ","
		}

		reValue = "^" + reType + "(" + delimeter + reType + ")*$"
	} else {
		reValue = "^" + reType + "$"
	}

	m, err := regexp.MatchString(reValue, paramValue)
	if err != nil || !m {
		return errParamValueInvalid
	}

	return nil
}
