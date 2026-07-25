package godotenv

import "testing"

func TestDoubleQuoteEscape_noBackslashNNReplaced(t *testing.T) {
	// The old implementation used sequential ReplaceAll which corrupted
	// values containing backslashes followed by 'n'. A literal "\n" in
	// the value should produce "\\n" not "\\\n".
	input := `hello\nworld`
	got := doubleQuoteEscape(input)
	expected := `hello\\nworld`
	if got != expected {
		t.Errorf("doubleQuoteEscape(%q) = %q, want %q", input, got, expected)
	}
}

func TestDoubleQuoteEscape_escapesAllSpecialChars(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{`\`, `\\`},
		{`"`, `\"`},
		{string('\r'), `\r`},
		{"$VAR", `\$VAR`},
		{`!bang`, `\!bang`},
		{`test\nline`, `test\\nline`}, // literal backslash-n, not newline char
	}
	for _, tt := range tests {
		got := doubleQuoteEscape(tt.input)
		if got != tt.want {
			t.Errorf("doubleQuoteEscape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
