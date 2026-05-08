package broccli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func initTestCLI(t *testing.T) (*os.File, *os.File) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Error("error creating temporary file")

		return nil, nil
	}

	devNull, err := os.OpenFile("/dev/null", os.O_APPEND, 0o600)
	if err != nil {
		t.Error("error opening temporary file")

		err := os.Remove(tmpFile.Name())
		if err != nil {
			t.Error("error removing temporary file")
		}

		err = tmpFile.Close()
		if err != nil {
			t.Error("error closing temporary file")
		}

		return nil, nil
	}

	os.Stdout = devNull
	os.Stderr = devNull

	return tmpFile, devNull
}

func removeTestFiles(t *testing.T, tmpFile *os.File, devNull *os.File) {
	err := os.Remove(tmpFile.Name())
	if err != nil {
		t.Error("error removing temporary file")
	}

	err = tmpFile.Close()
	if err != nil {
		t.Error("error closing temporary file")
	}

	err = devNull.Close()
	if err != nil {
		t.Error("error closing /dev/null")
	}
}

// TestCLIStringFlag tests a CLI instance with single flag instance.
func TestCLIStringFlag(t *testing.T) {
	t.Parallel()

	tmpFile, devNull := initTestCLI(t)
	defer func() {
		removeTestFiles(t, tmpFile, devNull)
	}()

	broccli := NewBroccli("Example", "App", "Author <a@example.com>")
	cmd1 := broccli.Command("cmd", "Prints out a string", func(_ context.Context, c *Broccli) int {
		text := c.Flag("text")
		if text == "tekst" {
			return 2
		}

		return 3
	})
	cmd1.Flag("text", "t", "Text", "Text to check", TypeString, IsRequired)

	os.Args = []string{"test", "wrongcmd"}
	got := broccli.Run(context.Background())
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd"}
	got = broccli.Run(context.Background())
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd", "-t", "tekst"}
	got = broccli.Run(context.Background())
	if got != 2 {
		t.Errorf("CLI.Run() should have returned 2 instead of %d", got)
	}

	os.Args = []string{"test", "cmd", "-t", "return3"}
	got = broccli.Run(context.Background())
	if got != 3 {
		t.Errorf("CLI.Run() should have returned 3 instead of %d", got)
	}

	os.Args = []string{"test", "cmd", "--text", "tekst"}
	got = broccli.Run(context.Background())
	if got != 2 {
		t.Errorf("CLI.Run() should have returned 2 instead of %d", got)
	}

	os.Args = []string{"test", "cmd", "--text", "return3"}
	got = broccli.Run(context.Background())
	if got != 3 {
		t.Errorf("CLI.Run() should have returned 3 instead of %d", got)
	}
}

// TestCLIStringFlagNoAlias tests a CLI instance with single flag that does not have an alias.
func TestCLIStringFlagNoAlias(t *testing.T) {
	t.Parallel()

	tmpFile, devNull := initTestCLI(t)
	defer func() {
		removeTestFiles(t, tmpFile, devNull)
	}()

	broccli := NewBroccli("Example", "App", "Author <a@example.com>")
	cmd1 := broccli.Command("cmd", "Prints out a string", func(_ context.Context, c *Broccli) int {
		text := c.Flag("text")
		text2 := c.Flag("text2")
		if text == "tekst" {
			return 2
		}
		if text2 == "tekst2" {
			return 4
		}

		return 3
	})
	cmd1.Flag("text", "", "Text", "Text to check", TypeString, IsRequired)
	cmd1.Flag("text2", "", "Text2", "Text2 to check", TypeString, IsRequired)

	os.Args = []string{"test", "wrongcmd"}
	got := broccli.Run(context.Background())
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd"}
	got = broccli.Run(context.Background())
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd", "--text", "tekst", "--text2", "tekst2"}
	got = broccli.Run(context.Background())
	if got != 2 {
		t.Errorf("CLI.Run() should have returned 2 instead of %d", got)
	}

	os.Args = []string{"test", "cmd", "--text", "return3", "--text2", "return3"}
	got = broccli.Run(context.Background())
	if got != 3 {
		t.Errorf("CLI.Run() should have returned 3 instead of %d", got)
	}
}

// TestCLIVariousFlags tests a CLI with various types of flags
func TestCLIVariousFlags(t *testing.T) {
	t.Parallel()

	tmpFile, devNull := initTestCLI(t)
	defer func() {
		removeTestFiles(t, tmpFile, devNull)
	}()

	c := NewBroccli("Example", "App", "Author <a@example.com>")
	cmd1 := c.Command("cmd1", "Prints out a string", func(_ context.Context, c *Broccli) int {
		_, _ = fmt.Fprintf(tmpFile, "TESTVALUE:%s%s\n\n", c.Flag("tekst"), c.Flag("alphanumdots"))

		if c.Flag("bool") == "true" {
			_, _ = fmt.Fprintf(tmpFile, "BOOL:true")
		}

		return 2
	})
	cmd1.Flag("tekst", "t", "Text", "Text to print", TypeString, IsRequired)
	cmd1.Flag("alphanumdots", "a", "Alphanum with dots", "Can have dots", TypeAlphanumeric, AllowDots)
	cmd1.Flag("make-required", "r", "", "Make alphanumdots required", TypeBool, 0,
		OnTrue(func(c *Command) {
			c.flags["alphanumdots"].flags |= IsRequired
		}),
	)
	// Boolean should work fine even when the optional OnTrue is not passed
	cmd1.Flag("bool", "b", "", "Bool value", TypeBool, 0)

	os.Args = []string{"test", "cmd1"}
	got := c.Run(context.Background())
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "-t", ""}
	got = c.Run(context.Background())
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "--tekst", "Tekst123", "--alphanumdots"}
	got = c.Run(context.Background())
	if got != 2 {
		t.Errorf("CLI.Run() should have returned 2 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "--tekst", "Tekst123", "-r"}
	got = c.Run(context.Background())
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "--tekst", "Tekst123", "--alphanumdots", "aZ0-9"}
	got = c.Run(context.Background())
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "--tekst", "Tekst123", "--alphanumdots", "aZ0.9", "-b"}
	got = c.Run(context.Background())
	if got != 2 {
		t.Errorf("CLI.Run() should have returned 2 instead of %d", got)
	}

	f2, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Error("error opening temporary file")
	}

	defer func() {
		err := f2.Close()
		if err != nil {
			t.Error("error closing temporary file")
		}
	}()

	b, err := io.ReadAll(f2)
	if err != nil {
		t.Error("error reading output file contents")
	}

	if !strings.Contains(string(b), "TESTVALUE:Tekst123aZ0.9") {
		t.Errorf("Cmd handler failed to work")
	}

	if !strings.Contains(string(b), "BOOL:true") {
		t.Errorf("Cmd handler failed to work")
	}
}
