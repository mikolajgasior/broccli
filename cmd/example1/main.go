// main is an example broccli usage.
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikolajgasior/broccli/v3"
)

func main() {
	cli := broccli.NewBroccli("example1", "Example app", "author@example.com")

	printCmd := cli.Command("print", "Prints a hello message", printHandler)

	printCmd.Arg(
		"first-name",
		"FIRST_NAME",
		"First name of the person to welcome",
		broccli.TypeString,
		broccli.IsRequired,
	)
	printCmd.Arg("last-name", "LAST_NAME", "Optional last name", broccli.TypeString, 0)

	printCmd.Flag(
		"language-file",
		"l",
		"PATH_TO_FILE",
		"File containing 'hello' in many languages",
		broccli.TypePathFile,
		broccli.IsRegularFile|broccli.IsExistent|broccli.IsRequired,
	)
	printCmd.Flag("alternative", "a", "", "Use alternative welcoming", broccli.TypeBool, 0)

	os.Exit(cli.Run(context.Background()))
}

func printHandler(_ context.Context, cli *broccli.Broccli) int {
	langFile := cli.Flag("language-file")

	file, err := os.Open(filepath.Clean(langFile))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file %s: %s", langFile, err.Error())

		return 1
	}

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
	}

	i, _ := rand.Int(rand.Reader, big.NewInt(int64(len(lines)-1)))
	messageArr := strings.Split(lines[i.Int64()], ":")

	message := messageArr[0]

	if cli.Flag("alternative") == "true" {
		message = messageArr[1]
	}

	firstName := cli.Arg("first-name")

	lastName := ""
	if cli.Arg("last-name") != "" {
		lastName = " " + cli.Arg("last-name")
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s, %s%s!", message, firstName, lastName)

	return 0
}
