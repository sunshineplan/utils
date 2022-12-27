package flags

import (
	"flag"
	"testing"
)

var (
	var0 = flag.String("var0", "", "")
	var1 = flag.String("var1", "", "")
	var2 = flag.String("var2", "2", "")
)

func TestParse(t *testing.T) {
	config = "test_config.ini"
	Parse()
	if *var0 != "0" {
		t.Errorf("expected %q; got %q", "0", *var0)
	}
	if *var1 != "1" {
		t.Errorf("expected %q; got %q", "1", *var1)
	}
	if *var2 != "" {
		t.Errorf("expected %q; got %q", "", *var2)
	}
}
