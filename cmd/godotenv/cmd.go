package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"

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
Run a process with an env setup from a .env file

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
	var envFilenames []string
	if rawEnvFilenames != "" {
		envFilenames = strings.Split(rawEnvFilenames, ",")
	}

	if err := godotenv.Load(envFilenames...); err != nil {
		log.Fatal(err)
	}

	// take rest of args and "exec" them
	cmd := args[0]
	cmdArgs := args[1:]

	command := exec.Command(cmd, cmdArgs...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// Ignore interrupts so we don't exit before the sub-process does.
	// This signal will still get passed to the sub-process.
	signal.Ignore(os.Interrupt)

	if err := command.Run(); err != nil {
		log.Fatal(err)
	}
}
