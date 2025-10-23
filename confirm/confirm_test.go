package confirm

import (
	"bytes"
	"strings"
	"testing"
)

func TestDo_YesResponses(t *testing.T) {
	tests := []string{"y\n", "Y\n", "yes\n", " YES \n"}
	for _, input := range tests {
		t.Run(strings.TrimSpace(input), func(t *testing.T) {
			in := strings.NewReader(input)
			var out bytes.Buffer
			got := do("Confirm?", 1, &out, in)
			if !got {
				t.Errorf("expected true for input %q, got false", input)
			}
		})
	}
}

func TestDo_NoResponses(t *testing.T) {
	tests := []string{"n\n", "N\n", "no\n", " NO \n"}
	for _, input := range tests {
		t.Run(strings.TrimSpace(input), func(t *testing.T) {
			in := strings.NewReader(input)
			var out bytes.Buffer
			got := do("Confirm?", 1, &out, in)
			if got {
				t.Errorf("expected false for input %q, got true", input)
			}
		})
	}
}

func TestDo_InvalidThenYes(t *testing.T) {
	in := strings.NewReader("maybe\nYES\n")
	var out bytes.Buffer
	got := do("Confirm?", 2, &out, in)
	if !got {
		t.Errorf("expected true after invalid input then yes, got false")
	}
	if !strings.Contains(out.String(), "Please type 'yes' or 'no':") {
		t.Errorf("expected retry prompt, got %q", out.String())
	}
}

func TestDo_MaxRetriesExceeded(t *testing.T) {
	in := strings.NewReader("maybe\nidk\nnope\n")
	var out bytes.Buffer
	got := do("Confirm?", 3, &out, in)
	if got {
		t.Errorf("expected false after max retries, got true")
	}
	if !strings.Contains(out.String(), "Max retries exceeded.") {
		t.Errorf("expected max retries message, got %q", out.String())
	}
}

func TestDo_DefaultPromptAndAttempts(t *testing.T) {
	in := strings.NewReader("y\n")
	var out bytes.Buffer
	got := do("", 0, &out, in)
	if !got {
		t.Errorf("expected true with default prompt, got false")
	}
	if !strings.Contains(out.String(), "Are you sure? (yes/no):") {
		t.Errorf("expected default prompt, got %q", out.String())
	}
}
