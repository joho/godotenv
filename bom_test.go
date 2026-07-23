package godotenv

import (
	"strings"
	"testing"
)

func TestParseUTF8BOM(t *testing.T) {
	// Windows editors often save .env with a UTF-8 BOM.
	input := "\ufeffFOO=bar\nBAZ=qux\n"
	env, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if env["FOO"] != "bar" {
		t.Fatalf("FOO: got %q want bar (keys=%v)", env["FOO"], env)
	}
	if env["BAZ"] != "qux" {
		t.Fatalf("BAZ: got %q want qux", env["BAZ"])
	}
}

func TestUnmarshalBytesUTF8BOM(t *testing.T) {
	env, err := UnmarshalBytes([]byte("\xef\xbb\xbfHELLO=world"))
	if err != nil {
		t.Fatalf("UnmarshalBytes: %v", err)
	}
	if env["HELLO"] != "world" {
		t.Fatalf("HELLO: got %q want world", env["HELLO"])
	}
}
