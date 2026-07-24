//go:build go1.16

package godotenv

import (
	"io/fs"
	"os"
	"testing"
	"testing/fstest"
)

func testMapFS() fstest.MapFS {
	return fstest.MapFS{
		".env": &fstest.MapFile{
			Data: []byte("FROM_DEFAULT=yes\n"),
		},
		"plain.env": &fstest.MapFile{
			Data: []byte("OPTION_A=1\nOPTION_B=2\nOPTION_C=3\nOPTION_D=4\nOPTION_E=5\nOPTION_F=\nOPTION_G=\nOPTION_H=1 2\n"),
		},
		"quoted.env": &fstest.MapFile{
			Data: []byte("OPTION_A='1'\nOPTION_B='2'\nOPTION_C=''\nOPTION_D='\\n'\nOPTION_E=\"1\"\nOPTION_F=\"2\"\nOPTION_G=\"\"\nOPTION_H=\"\\n\"\nOPTION_I=\"echo 'asd'\"\n"),
		},
		"equals.env": &fstest.MapFile{
			Data: []byte("OPTION_A=postgres://localhost:5432/database?sslmode=disable\n"),
		},
	}
}

func loadFSEnvAndCompareValues(t *testing.T, loader func(fsys fs.FS, files ...string) error, fsys fs.FS, envFileName string, expectedValues map[string]string, presets map[string]string) {
	t.Helper()

	os.Clearenv()

	for k, v := range presets {
		os.Setenv(k, v)
	}

	err := loader(fsys, envFileName)
	if err != nil {
		t.Fatalf("Error loading %v: %v", envFileName, err)
	}

	for k := range expectedValues {
		envValue := os.Getenv(k)
		v := expectedValues[k]
		if envValue != v {
			t.Errorf("Mismatch for key '%v': expected '%#v' got '%#v'", k, v, envValue)
		}
	}
}

func TestLoadFSWithNoArgsLoadsDotEnv(t *testing.T) {
	fsys := testMapFS()
	// remove default so open fails the same way as missing file
	delete(fsys, ".env")

	err := LoadFS(fsys)
	if err == nil {
		t.Fatal("expected error opening default .env")
	}
}

func TestOverloadFSWithNoArgsOverloadsDotEnv(t *testing.T) {
	fsys := testMapFS()
	delete(fsys, ".env")

	err := OverloadFS(fsys)
	if err == nil {
		t.Fatal("expected error opening default .env")
	}
}

func TestLoadFSFileNotFound(t *testing.T) {
	err := LoadFS(testMapFS(), "somefilethatwillneverexistever.env")
	if err == nil {
		t.Error("File wasn't found but LoadFS didn't return an error")
	}
}

func TestOverloadFSFileNotFound(t *testing.T) {
	err := OverloadFS(testMapFS(), "somefilethatwillneverexistever.env")
	if err == nil {
		t.Error("File wasn't found but OverloadFS didn't return an error")
	}
}

func TestReadFSPlainEnv(t *testing.T) {
	envMap, err := ReadFS(testMapFS(), "plain.env")
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	expectedValues := map[string]string{
		"OPTION_A": "1",
		"OPTION_B": "2",
		"OPTION_C": "3",
		"OPTION_D": "4",
		"OPTION_E": "5",
		"OPTION_F": "",
		"OPTION_G": "",
		"OPTION_H": "1 2",
	}

	if len(envMap) != len(expectedValues) {
		t.Errorf("Didn't get the right size map back: got %d want %d", len(envMap), len(expectedValues))
	}

	for key, value := range expectedValues {
		if envMap[key] != value {
			t.Errorf("ReadFS key %q: got %q want %q", key, envMap[key], value)
		}
	}
}

func TestLoadFSDoesNotOverride(t *testing.T) {
	presets := map[string]string{
		"OPTION_A": "do_not_override",
		"OPTION_B": "",
	}
	expectedValues := map[string]string{
		"OPTION_A": "do_not_override",
		"OPTION_B": "",
	}
	loadFSEnvAndCompareValues(t, LoadFS, testMapFS(), "plain.env", expectedValues, presets)
}

func TestOverloadFSDoesOverride(t *testing.T) {
	presets := map[string]string{
		"OPTION_A": "do_not_override",
	}
	expectedValues := map[string]string{
		"OPTION_A": "1",
	}
	loadFSEnvAndCompareValues(t, OverloadFS, testMapFS(), "plain.env", expectedValues, presets)
}

func TestLoadFSPlainEnv(t *testing.T) {
	expectedValues := map[string]string{
		"OPTION_A": "1",
		"OPTION_B": "2",
		"OPTION_C": "3",
		"OPTION_D": "4",
		"OPTION_E": "5",
		"OPTION_H": "1 2",
	}
	loadFSEnvAndCompareValues(t, LoadFS, testMapFS(), "plain.env", expectedValues, noopPresets)
}

func TestLoadFSEqualsEnv(t *testing.T) {
	expectedValues := map[string]string{
		"OPTION_A": "postgres://localhost:5432/database?sslmode=disable",
	}
	loadFSEnvAndCompareValues(t, LoadFS, testMapFS(), "equals.env", expectedValues, noopPresets)
}

func TestLoadFSDefaultFilename(t *testing.T) {
	os.Clearenv()
	if err := LoadFS(testMapFS()); err != nil {
		t.Fatalf("LoadFS default: %v", err)
	}
	if got := os.Getenv("FROM_DEFAULT"); got != "yes" {
		t.Errorf("FROM_DEFAULT: got %q want %q", got, "yes")
	}
}

func TestLoadFSMultipleFiles(t *testing.T) {
	os.Clearenv()
	fsys := fstest.MapFS{
		"a.env": &fstest.MapFile{Data: []byte("ONE=1\nSHARED=from-a\n")},
		"b.env": &fstest.MapFile{Data: []byte("TWO=2\nSHARED=from-b\n")},
	}
	if err := LoadFS(fsys, "a.env", "b.env"); err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	if got := os.Getenv("ONE"); got != "1" {
		t.Errorf("ONE: got %q want 1", got)
	}
	if got := os.Getenv("TWO"); got != "2" {
		t.Errorf("TWO: got %q want 2", got)
	}
	// Load does not override: first file wins for SHARED
	if got := os.Getenv("SHARED"); got != "from-a" {
		t.Errorf("SHARED: got %q want from-a", got)
	}
}

func TestOverloadFSMultipleFiles(t *testing.T) {
	os.Clearenv()
	fsys := fstest.MapFS{
		"a.env": &fstest.MapFile{Data: []byte("ONE=1\nSHARED=from-a\n")},
		"b.env": &fstest.MapFile{Data: []byte("TWO=2\nSHARED=from-b\n")},
	}
	if err := OverloadFS(fsys, "a.env", "b.env"); err != nil {
		t.Fatalf("OverloadFS: %v", err)
	}
	if got := os.Getenv("ONE"); got != "1" {
		t.Errorf("ONE: got %q want 1", got)
	}
	if got := os.Getenv("TWO"); got != "2" {
		t.Errorf("TWO: got %q want 2", got)
	}
	// Overload: later file wins for SHARED
	if got := os.Getenv("SHARED"); got != "from-b" {
		t.Errorf("SHARED: got %q want from-b", got)
	}
}

func TestFSActualEnvVarsAreLeftAlone(t *testing.T) {
	os.Clearenv()
	os.Setenv("OPTION_A", "actualenv")
	if err := LoadFS(testMapFS(), "plain.env"); err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	if os.Getenv("OPTION_A") != "actualenv" {
		t.Error("An ENV var set earlier was overwritten")
	}
}

func TestReadFSMultipleFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"a.env": &fstest.MapFile{Data: []byte("ONE=1\nSHARED=from-a\n")},
		"b.env": &fstest.MapFile{Data: []byte("TWO=2\nSHARED=from-b\n")},
	}
	envMap, err := ReadFS(fsys, "a.env", "b.env")
	if err != nil {
		t.Fatalf("ReadFS: %v", err)
	}
	if envMap["ONE"] != "1" || envMap["TWO"] != "2" {
		t.Errorf("unexpected map: %#v", envMap)
	}
	if envMap["SHARED"] != "from-b" {
		t.Errorf("SHARED: got %q want from-b", envMap["SHARED"])
	}
}
