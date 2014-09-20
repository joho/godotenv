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
	showHelp := *flag.Bool("h", false, "show help")
	rawEnvFilenames := *flag.String("f", "", "comma separated paths to .env files")

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
	if showHelp || len(os.Args) < 2 {
		fmt.Println(usage)
		return
	}

	// load env
	// TODO something in flag passing or whatever isn't quite right
	envFilenames := strings.Split(rawEnvFilenames, ",")
	fmt.Printf("env filenames %v\n", envFilenames)
	godotenv.Load(envFilenames...)
	fmt.Printf("FOO=%v\n", os.Getenv("FOO"))

	// take rest of args and "exec" them
	args := flag.Args()
	cmd := args[0]
	cmdArgs := args[1:len(args)]
	fmt.Printf("cmd %v with args: %v\n", cmd, cmdArgs)

	command := exec.Command(cmd, cmdArgs...)
	// command.Env = os.Environ()
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Start()
}
