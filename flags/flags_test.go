package flags

import (
	"flag"
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
	flag.CommandLine.Init("", flag.PanicOnError)
	defer func() {
		if err := recover(); err == nil {
			t.Error("gave no panic; want panic")
		}
	}()
	Parse()
}
