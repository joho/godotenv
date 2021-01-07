// Package godotenv is a go port of the ruby dotenv library (https://github.com/bkeepers/dotenv)
//
// Examples/readme can be found on the github page at https://github.com/joho/godotenv
//
// The TL;DR is that you make a .env file that looks something like
//
// 		SOME_ENV_VAR=somevalue
//
// and then in your go code you can call
//
// 		godotenv.Load()
//
// and all the env vars declared in .env will be available through os.Getenv("SOME_ENV_VAR")
package godotenv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const doubleQuoteSpecialChars = "\\\n\r\"!$`"

// Load will read your env file(s) and load them into ENV for this process.
//
// Call this function as close as possible to the start of your program (ideally in main)
//
// If you call Load without any args it will default to loading .env in the current path
//
// You can otherwise tell it which files to load (there can be more than one) like
//
//		godotenv.Load("fileone", "filetwo")
//
// It's important to note that it WILL NOT OVERRIDE an env variable that already exists - consider the .env file to set dev vars or sensible defaults
func Load(filenames ...string) error {
	envMap, err := Read(filenames...)
	if err != nil {
		return err
	}

	for _, e := range MergeEnvSlices(EnvMapToSlice(envMap), os.Environ()) {
		// We assume env slice always has key and some value (even empty).
		if err := os.Setenv(strings.Split(e, "=")[0], strings.SplitN(e, "=", 2)[1]); err != nil {
			return err
		}
	}
	return nil
}

// Overload will read your env file(s) and load them into ENV for this process.
//
// Call this function as close as possible to the start of your program (ideally in main)
//
// If you call Overload without any args it will default to loading .env in the current path
//
// You can otherwise tell it which files to load (there can be more than one) like
//
//		godotenv.Overload("fileone", "filetwo")
//
// It's important to note this WILL OVERRIDE an env variable that already exists - consider the .env file to forcefilly set all vars.
func Overload(filenames ...string) (err error) {
	envMap, err := Read(filenames...)
	if err != nil {
		return err
	}

	for _, e := range MergeEnvSlices(os.Environ(), EnvMapToSlice(envMap)) {
		// We assume env slice always has key and some value (even empty).
		if err := os.Setenv(strings.Split(e, "=")[0], strings.SplitN(e, "=", 2)[1]); err != nil {
			return err
		}
	}
	return nil
}

// Read all env (with same file loading semantics as Load) but return values as
// a map rather than automatically writing values into env
func Read(filenames ...string) (envMap map[string]string, err error) {
	filenames = filenamesOrDefault(filenames)
	envMap = make(map[string]string)

	for _, filename := range filenames {
		individualEnvMap, individualErr := readFile(filename)
		if individualErr != nil {
			return nil, individualErr // Return early on a spazout.
		}

		for key, value := range individualEnvMap {
			envMap[key] = value
		}
	}
	return envMap, nil
}

// Parse reads an env file from io.Reader, returning a map of keys and values.
func Parse(r io.Reader) (envMap map[string]string, err error) {
	envMap = make(map[string]string)

	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	for _, fullLine := range lines {
		if !isIgnoredLine(fullLine) {
			var key, value string
			key, value, err = parseLine(fullLine, envMap)
			if err != nil {
				return nil, err
			}
			envMap[key] = value
		}
	}
	return
}

// Unmarshal reads an env file from a string, returning a map of keys and values.
func Unmarshal(str string) (envMap map[string]string, err error) {
	return Parse(strings.NewReader(str))
}

// EnvMapToSlice converts env map that you can get from `Read` or `Parse` into env sorted string
// slice commonly used by `exec.Command`. This allows to execute new process with relevant
// variables without changing current process environment.
// See https://golang.org/pkg/os/exec/#Cmd `Env` field to read more about slice format.
func EnvMapToSlice(m map[string]string) []string {
	s := make([]string, 0, len(m))
	for k, v := range m {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(s)
	return s
}

// Exec executes given command with args in separate process in environment that consists
// of env vars from the specified filenames (empty map falls back to default) and current process env vars on top.
//
// Simply hooks up os.Stdin/err/out to the command and calls Run()
//
// If you want more fine grained control over your command it's recommended
// that you use `Load()` or `Read()` and the `os/exec` package yourself.
func Exec(filenames []string, cmd string, cmdArgs []string) error {
	envMap, err := Read(filenames...)
	if err != nil {
		return err
	}

	command := exec.Command(cmd, cmdArgs...)
	command.Env = MergeEnvSlices(EnvMapToSlice(envMap), os.Environ())
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

// Write serializes the given environment and writes it to a file
func Write(envMap map[string]string, filename string) error {
	content, err := Marshal(envMap)
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	file.Sync()
	return err
}

// Marshal outputs the given environment as a dotenv-formatted environment file.
// Each line is in the format: KEY="VALUE" where VALUE is backslash-escaped.
func Marshal(envMap map[string]string) (string, error) {
	lines := make([]string, 0, len(envMap))
	for k, v := range envMap {
		if d, err := strconv.Atoi(v); err == nil {
			lines = append(lines, fmt.Sprintf(`%s=%d`, k, d))
		} else {
			lines = append(lines, fmt.Sprintf(`%s="%s"`, k, doubleQuoteEscape(v)))
		}
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n"), nil
}

func filenamesOrDefault(filenames []string) []string {
	if len(filenames) == 0 {
		return []string{".env"}
	}
	return filenames
}

// MergeEnvSlices merges two slices into single sorted one by applying `over` slice into `base`.
// The `over` slice will be used if the key overlaps.
func MergeEnvSlices(base, over []string) (merged []string) {
	sort.Strings(base)
	sort.Strings(over)

	var b, o int
	for b < len(base) || o < len(over) {

		if b >= len(base) {
			merged = append(merged, over[o])
			o++
			continue
		}

		if o >= len(over) {
			merged = append(merged, base[b])
			b++
			continue
		}

		switch strings.Compare(strings.Split(base[b], "=")[0], strings.Split(over[o], "=")[0]) {
		case 0:
			// Same keys. Instead of picking over element, ignore base one. This ensure correct behaviour if base
			// has duplicate elements.
			b++
		case 1:
			merged = append(merged, over[o])
			o++
		case -1:
			merged = append(merged, base[b])
			b++
		}
	}
	return merged
}

func readFile(filename string) (envMap map[string]string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	return Parse(file)
}

var exportRegex = regexp.MustCompile(`^\s*(?:export\s+)?(.*?)\s*$`)

func parseLine(line string, envMap map[string]string) (key string, value string, err error) {
	if len(line) == 0 {
		err = errors.New("zero length string")
		return
	}

	// ditch the comments (but keep quoted hashes)
	if strings.Contains(line, "#") {
		segmentsBetweenHashes := strings.Split(line, "#")
		quotesAreOpen := false
		var segmentsToKeep []string
		for _, segment := range segmentsBetweenHashes {
			if strings.Count(segment, "\"") == 1 || strings.Count(segment, "'") == 1 {
				if quotesAreOpen {
					quotesAreOpen = false
					segmentsToKeep = append(segmentsToKeep, segment)
				} else {
					quotesAreOpen = true
				}
			}

			if len(segmentsToKeep) == 0 || quotesAreOpen {
				segmentsToKeep = append(segmentsToKeep, segment)
			}
		}

		line = strings.Join(segmentsToKeep, "#")
	}

	firstEquals := strings.Index(line, "=")
	firstColon := strings.Index(line, ":")
	splitString := strings.SplitN(line, "=", 2)
	if firstColon != -1 && (firstColon < firstEquals || firstEquals == -1) {
		//this is a yaml-style line
		splitString = strings.SplitN(line, ":", 2)
	}

	if len(splitString) != 2 {
		err = errors.New("Can't separate key from value")
		return
	}

	// Parse the key
	key = splitString[0]
	if strings.HasPrefix(key, "export") {
		key = strings.TrimPrefix(key, "export")
	}
	key = strings.TrimSpace(key)

	key = exportRegex.ReplaceAllString(splitString[0], "$1")

	// Parse the value
	value = parseValue(splitString[1], envMap)
	return
}

var (
	singleQuotesRegex  = regexp.MustCompile(`\A'(.*)'\z`)
	doubleQuotesRegex  = regexp.MustCompile(`\A"(.*)"\z`)
	escapeRegex        = regexp.MustCompile(`\\.`)
	unescapeCharsRegex = regexp.MustCompile(`\\([^$])`)
)

func parseValue(value string, envMap map[string]string) string {

	// trim
	value = strings.Trim(value, " ")

	// check if we've got quoted values or possible escapes
	if len(value) > 1 {
		singleQuotes := singleQuotesRegex.FindStringSubmatch(value)

		doubleQuotes := doubleQuotesRegex.FindStringSubmatch(value)

		if singleQuotes != nil || doubleQuotes != nil {
			// pull the quotes off the edges
			value = value[1 : len(value)-1]
		}

		if doubleQuotes != nil {
			// expand newlines
			value = escapeRegex.ReplaceAllStringFunc(value, func(match string) string {
				c := strings.TrimPrefix(match, `\`)
				switch c {
				case "n":
					return "\n"
				case "r":
					return "\r"
				default:
					return match
				}
			})
			// unescape characters
			value = unescapeCharsRegex.ReplaceAllString(value, "$1")
		}

		if singleQuotes == nil {
			value = expandVariables(value, envMap)
		}
	}

	return value
}

var expandVarRegex = regexp.MustCompile(`(\\)?(\$)(\()?\{?([A-Z0-9_]+)?\}?`)

func expandVariables(v string, m map[string]string) string {
	return expandVarRegex.ReplaceAllStringFunc(v, func(s string) string {
		submatch := expandVarRegex.FindStringSubmatch(s)

		if submatch == nil {
			return s
		}
		if submatch[1] == "\\" || submatch[2] == "(" {
			return submatch[0][1:]
		} else if submatch[4] != "" {
			return m[submatch[4]]
		}
		return s
	})
}

func isIgnoredLine(line string) bool {
	trimmedLine := strings.TrimSpace(line)
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}

func doubleQuoteEscape(line string) string {
	for _, c := range doubleQuoteSpecialChars {
		toReplace := "\\" + string(c)
		if c == '\n' {
			toReplace = `\n`
		}
		if c == '\r' {
			toReplace = `\r`
		}
		line = strings.Replace(line, string(c), toReplace, -1)
	}
	return line
}
