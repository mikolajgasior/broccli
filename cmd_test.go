package broccli

import (
	"fmt"
	"testing"
)

// TestCommandParams creates a dummy Command instance and tests attaching flags, args and environment variables.
func TestCommandParams(t *testing.T) {
	t.Parallel()

	c := &Command{}
	c.Flag("flag1", "f1", "int", "Flag 1", TypeInt, IsRequired)
	c.Flag("flag2", "f2", "path", "Flag 2", TypePathFile, IsRegularFile)
	c.Flag("flag3", "f3", "", "Flag 3", TypeBool, 0, OnTrue(func(c *Command) {
		c.flags["flag2"].flags |= IsExistent
	}))
	c.Arg("arg1", "ARG1", "Arg 1", TypeInt, IsRequired)
	c.Arg("arg2", "ARG2", "Arg 2", TypeAlphanumeric, 0)
	c.Env("ENVVAR1", "Env var 1", TypeInt, 0)

	sa := c.sortedArgs()
	sf := c.sortedFlags()
	se := c.sortedEnv()

	if len(sa) != 2 || len(sf) != 3 || len(se) != 1 {
		t.Errorf("Invalid args or flags or env vars added")
	}

	for i, a := range sa {
		if a != fmt.Sprintf("arg%d", (i+1)) {
			t.Errorf("Invalid arg was added")
		}
	}

	for i, f := range sf {
		if f != fmt.Sprintf("flag%d", (i+1)) {
			t.Errorf("Invalid flag was added")
		}
	}

	for i, e := range se {
		if e != fmt.Sprintf("ENVVAR%d", (i+1)) {
			t.Errorf("Invalid env var was added")
		}
	}
}
