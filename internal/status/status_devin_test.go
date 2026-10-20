package status

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func devinFrame(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "devin", name+".txt"))
	if err != nil {
		t.Fatal(err)
	}
	// A Windows checkout with core.autocrlf holds the fixtures with CRLF
	// endings; the panes the engine matches use LF.
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}

// devin 3000.11.92 fixtures are captures of the real TUI in a 200x50 pane:
// the launch banner, an autocomplete popup, modal dialogs (trust, resume
// and model pickers), a running turn, a plain-text question, a resting
// composer after a tip, and the echoed rows a cancelled or submitted
// numbered draft leaves behind — the shape a status rule must never
// mistake for a menu.
func TestDevinPanes(t *testing.T) {
	engine := defaultEngine(t)
	for _, tc := range []struct{ frame, want string }{
		{"launched", Finished},
		{"autocomplete", Finished},
		{"numbered-draft", Finished},
		{"echoed-numbered-draft", Finished},
		{"resting-tip", Finished},
		{"working", Working},
		{"trust", Waiting},
		{"resume-picker", Waiting},
		{"model-picker", Waiting},
	} {
		t.Run(tc.frame, func(t *testing.T) {
			if got, _ := engine.Match("devin", devinFrame(t, tc.frame)); got != tc.want {
				t.Fatalf("Match(%s) = %q want %q", tc.frame, got, tc.want)
			}
		})
	}
}

// A composer holding a numbered draft, or a transcript keeping the echoed
// rows of one, is not a dialog: send_session must still type into it. The
// dialogs do hold.
func TestDevinTypingHold(t *testing.T) {
	engine := defaultEngine(t)
	for _, tc := range []struct{ frame, want string }{
		{"launched", ""},
		{"autocomplete", ""},
		{"numbered-draft", ""},
		{"echoed-numbered-draft", ""},
		{"resting-tip", ""},
		{"question", ""},
		{"working", Working},
		{"trust", Waiting},
		{"resume-picker", Waiting},
		{"model-picker", Waiting},
	} {
		t.Run(tc.frame, func(t *testing.T) {
			if got := engine.TypingHold("devin", devinFrame(t, tc.frame)); got != tc.want {
				t.Fatalf("TypingHold(%s) = %q want %q", tc.frame, got, tc.want)
			}
		})
	}
}

// A turn that ended on a plain-text question waits on an answer; the
// quiet-pane path reads it from the last content line.
func TestDevinQuestionTurnEnd(t *testing.T) {
	engine := defaultEngine(t)
	region, ok := engine.ActivityRegion("devin", devinFrame(t, "question"))
	if !ok {
		t.Fatal("no activity region in question frame")
	}
	if got := engine.TurnEndedState("devin", region); got != Waiting {
		t.Fatalf("TurnEndedState = %q want %q", got, Waiting)
	}
}
