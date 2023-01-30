package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	var showHelp, deleteFileOnError bool
	flag.BoolVar(&showHelp, "h", false, "show help")
	var rawEnvFilenames string
	flag.StringVar(&rawEnvFilenames, "f", "", "comma separated paths to .env files")
	flag.BoolVar(&deleteFileOnError, "s", false, "delete .env files if the command fails, copy exit status")

	flag.Parse()

	usage := `
Run a process with an env setup from a .env file

godotenv [-f ENV_FILE_PATHS] [-s] COMMAND_ARGS

ENV_FILE_PATHS: comma separated paths to .env files (.env is the default)
COMMAND_ARGS: command and args you want to run

  -s                         delete .env files if the command fails, copy exit status

example
  godotenv -f /path/to/something/.env,/another/path/.env fortune
`
	// if no args or -h flag
	// print usage and return
	args := flag.Args()
	if showHelp || len(args) == 0 {
		fmt.Println(usage)
		return
	}

	// load env
	var envFilenames []string
	if rawEnvFilenames != "" {
		envFilenames = strings.Split(rawEnvFilenames, ",")
	}

	// take rest of args and "exec" them
	cmd := args[0]
	cmdArgs := args[1:]

	err := godotenv.Exec(envFilenames, cmd, cmdArgs)
	if err != nil {
		if deleteFileOnError {
			if len(envFilenames) == 0 {
				envFilenames = []string{".env"}
			}
			for _, filename := range envFilenames {
				fileErr := os.Remove(filename)
				if fileErr != nil {
					log.Println(fileErr)
				}
			}
			if state, ok := err.(*exec.ExitError); ok {
				os.Exit(state.ExitCode())
			}
		}
		log.Fatal(err)
	}
}
