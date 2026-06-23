package flags

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"testing"
)

func TestParse(t *testing.T) {
	var0 := flag.String("var0", "", "")
	var1 := flag.String("var1", "", "")
	var2 := flag.String("var2", "2", "")
	config = "test_config.ini"
	if err := Parse(); err != nil {
		t.Error(err)
	}
	if *var0 != "0" {
		t.Errorf("expected %q; got %q", "0", *var0)
	}
	if *var1 != "1" {
		t.Errorf("expected %q; got %q", "1", *var1)
	}
	if *var2 != "" {
		t.Errorf("expected %q; got %q", "", *var2)
	}
	config = ""
}

func TestParsePanicOnError(t *testing.T) {
	flag.CommandLine.Init("", flag.PanicOnError)
	os.Args = []string{"cmd", "-badflag"}
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("gave no panic; want panic")
		}
	}()
	Parse()
}

func TestParseHelpExits(t *testing.T) {
	if os.Getenv("TEST_PARSE_HELP_EXIT") == "1" {
		flag.CommandLine.Init("", flag.ContinueOnError)
		os.Args = []string{"cmd", "-help"}
		Parse()
		t.Fatal("Parse did not exit")
	}

	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(exe, "-test.run=TestParseHelpExits")
	cmd.Env = append(os.Environ(), "TEST_PARSE_HELP_EXIT=1")
	if err := cmd.Run(); err != nil {
		if ee, ok := errors.AsType[*exec.ExitError](err); ok {
			if ee.ExitCode() != 0 {
				t.Fatalf("expected exit code 0; got %d", ee.ExitCode())
			}
		} else {
			t.Fatalf("expected ExitError or nil; got %T", err)
		}
	}
}

func TestParseHelpDoesNotExitWhenHelpExitFalse(t *testing.T) {
	HelpExit = false
	flag.CommandLine.Init("", flag.ContinueOnError)
	os.Args = []string{"cmd", "-help"}
	if err := Parse(); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("expected flag.ErrHelp; got %v", err)
	}
	HelpExit = true
}
