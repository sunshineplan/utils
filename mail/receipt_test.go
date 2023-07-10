package mail

import (
	"encoding/json"
	"flag"
	"testing"
)

func TestFlag(t *testing.T) {
	fs := flag.NewFlagSet("Test", 0)
	var r Receipts
	fs.TextVar(&r, "rcpt", Receipts(nil), "")
	for _, testcase := range []struct {
		data string
		res  string
	}{
		{"a@b.c", "<a@b.c>"},
		{"a@b.c,c@b.a", "<a@b.c>, <c@b.a>"},
	} {
		fs.Parse([]string{"-rcpt", testcase.data})
		if res := r.String(); res != testcase.res {
			t.Errorf("expected %q; got %q", testcase.res, res)
		}
	}
}

func TestJson(t *testing.T) {
	for _, testcase := range []struct {
		data string
		res  string
	}{
		{`"a@b.c"`, "<a@b.c>"},
		{`"a@b.c,c@b.a"`, "<a@b.c>, <c@b.a>"},
		{`["a@b.c","c@b.a"]`, "<a@b.c>, <c@b.a>"},
	} {
		var r Receipts
		if err := json.Unmarshal([]byte(testcase.data), &r); err != nil {
			t.Error(err)
			continue
		}
		if res := r.String(); res != testcase.res {
			t.Errorf("expected %q; got %q", testcase.res, res)
		}
	}
}
