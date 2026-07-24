//go:build go1.16

package godotenv

import (
	"io/fs"
)

// LoadFS is like Load but reads files from fsys (for example an embed.FS).
//
// If you call LoadFS without any filenames it will default to loading ".env".
// Existing environment variables are not overridden.
func LoadFS(fsys fs.FS, filenames ...string) (err error) {
	filenames = filenamesOrDefault(filenames)

	for _, filename := range filenames {
		err = loadFileFS(fsys, filename, false)
		if err != nil {
			return // return early on a spazout
		}
	}
	return
}

// OverloadFS is like Overload but reads files from fsys (for example an embed.FS).
//
// If you call OverloadFS without any filenames it will default to loading ".env".
// Existing environment variables ARE overridden.
func OverloadFS(fsys fs.FS, filenames ...string) (err error) {
	filenames = filenamesOrDefault(filenames)

	for _, filename := range filenames {
		err = loadFileFS(fsys, filename, true)
		if err != nil {
			return // return early on a spazout
		}
	}
	return
}

// ReadFS is like Read but reads files from fsys (for example an embed.FS).
//
// If you call ReadFS without any filenames it will default to loading ".env".
// Values are returned as a map and are not written into the process environment.
func ReadFS(fsys fs.FS, filenames ...string) (envMap map[string]string, err error) {
	filenames = filenamesOrDefault(filenames)
	envMap = make(map[string]string)

	for _, filename := range filenames {
		individualEnvMap, individualErr := readFileFS(fsys, filename)

		if individualErr != nil {
			err = individualErr
			return // return early on a spazout
		}

		for key, value := range individualEnvMap {
			envMap[key] = value
		}
	}

	return
}

func loadFileFS(fsys fs.FS, filename string, overload bool) error {
	envMap, err := readFileFS(fsys, filename)
	if err != nil {
		return err
	}

	return setEnvMap(envMap, overload)
}

func readFileFS(fsys fs.FS, filename string) (envMap map[string]string, err error) {
	file, err := fsys.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	return Parse(file)
}
