package display

import (
	"testing"
)

func assertUnchanged(t *testing.T, text string) {
	t.Helper()
	if got := dedent(text); got != text {
		t.Errorf("dedent() changed text.\nExpected: %q\nGot:      %q", text, got)
	}
}

func TestDedent_NoMargin(t *testing.T) {
	text := "Hello there.\nHow are you?\nOh good, I'm glad."
	assertUnchanged(t, text)

	text = "Hello there.\n\nBoo!"
	assertUnchanged(t, text)

	text = "Hello there.\n  This is indented."
	assertUnchanged(t, text)

	text = "Hello there.\n\n  Boo!\n"
	assertUnchanged(t, text)
}

func TestDedent_Even(t *testing.T) {
	text := "  Hello there.\n  How are ya?\n  Oh good."
	expect := "Hello there.\nHow are ya?\nOh good."
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}

	text = "  Hello there.\n\n  How are ya?\n  Oh good.\n"
	expect = "Hello there.\n\nHow are ya?\nOh good.\n"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}

	text = "  Hello there.\n  \n  How are ya?\n  Oh good.\n"
	expect = "Hello there.\n\nHow are ya?\nOh good.\n"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}
}

func TestDedent_Uneven(t *testing.T) {
	text := "    def foo():\n        while 1:\n            return foo\n    "
	expect := "def foo():\n    while 1:\n        return foo\n"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}

	text = "  Foo\n    Bar\n\n   Baz\n"
	expect = "Foo\n  Bar\n\n Baz\n"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}

	text = "  Foo\n    Bar\n \n   Baz\n"
	expect = "Foo\n  Bar\n\n Baz\n"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}
}

func TestDedent_Declining(t *testing.T) {
	text := "     Foo\n    Bar\n"
	expect := " Foo\nBar\n"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}

	text = "     Foo\n\n    Bar\n"
	expect = " Foo\n\nBar\n"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}

	text = "     Foo\n    \n    Bar\n"
	expect = " Foo\n\nBar\n"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}
}

func TestDedent_PreserveInternalTabs(t *testing.T) {
	text := "  hello\tthere\n  how are\tyou?"
	expect := "hello\tthere\nhow are\tyou?"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}

	if got := dedent(expect); got != expect {
		t.Errorf("dedent() changed text with no margin.\nExpected: %q\nGot:      %q", expect, got)
	}
}

func TestDedent_PreserveMarginTabs(t *testing.T) {
	text := "  hello there\n\thow are you?"
	assertUnchanged(t, text)

	text = "        hello there\n\thow are you?"
	assertUnchanged(t, text)

	text = "\thello there\n\thow are you?"
	expect := "hello there\nhow are you?"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}

	text = "  \thello there\n  \thow are you?"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}

	text = "  \t  hello there\n  \t  how are you?"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}

	text = "  \thello there\n  \t  how are you?"
	expect = "hello there\n  how are you?"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}

	text = "  \thello there\n   \thow are you?\n \tI'm fine, thanks"
	expect = " \thello there\n  \thow are you?\n\tI'm fine, thanks"
	if got := dedent(text); got != expect {
		t.Errorf("dedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}
}

// create a test for trimdedent which is much simpler
func TestTrimDedent(t *testing.T) {
	text := "\n  Hello there.\n  How are you?\n  Oh good.\n"
	expect := "Hello there.\nHow are you?\nOh good."
	if got := trimdedent(text); got != expect {
		t.Errorf("trimdedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}

	text = "\n  Hello there.\n\n  How are you?\n  Oh good.\n"
	expect = "Hello there.\n\nHow are you?\nOh good."
	if got := trimdedent(text); got != expect {
		t.Errorf("trimdedent() failed.\nExpected: %q\nGot:      %q", expect, got)
	}
}

// trim should remove whitespace up to and including the first newline at the start,
// and remove the last newline at the end, then dedent the remaining text.
func TestTrimDedentTrimming(t *testing.T) {
	result := "┌─────────────────────────┐\n│Hello                    │\n│World *                  │\n│Test                     │\n└─────────────────────────┘\n"
	expected := trimdedent(`
	┌─────────────────────────┐
	│Hello                    │
	│World *                  │
	│Test                     │
	└─────────────────────────┘
	`)
	if result != expected {
		t.Errorf("Expected highlighted display output to be:\n%s\nGot:\n%s", expected, result)
	}

	result = "A\nB\n"
	expected = trimdedent(`
	A
	B
	`)
	if result != expected {
		t.Errorf("Expected highlighted display output to be:\n%s\nGot:\n%s", expected, result)
	}

	// Remove the extra leading whitespace in `expected` var, before the first newline
	result = "A\nB\n"
	expected = trimdedent(`    
	A
	B
	`)
	if result != expected {
		t.Errorf("Expected highlighted display output to be:\n%s\nGot:\n%s", expected, result)
	}
}