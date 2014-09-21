package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	var showHelp bool
	flag.BoolVar(&showHelp, "h", false, "show help")
	var rawEnvFilenames string
	flag.StringVar(&rawEnvFilenames, "f", "", "comma separated paths to .env files")

	flag.Parse()

	usage := `
Run a process with a env setup from a .env file

godotenv [-f ENV_FILE_PATHS] COMMAND_ARGS

ENV_FILE_PATHS: comma separated paths to .env files
COMMAND_ARGS: command and args you want to run

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
	// TODO would be nicer for an empty rawEnvFilenames to give us an empty map
	// and then only call Load() once
	// other TODO error handling on Load()
	if rawEnvFilenames == "" {
		godotenv.Load()
	} else {
		envFilenames := strings.Split(rawEnvFilenames, ",")
		godotenv.Load(envFilenames...)
	}

	// take rest of args and "exec" them
	cmd := args[0]
	cmdArgs := args[1:len(args)]

	command := exec.Command(cmd, cmdArgs...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Start()
}
