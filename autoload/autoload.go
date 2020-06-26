package autoload

/*
	You can just read the .env file on import just by doing

		import _ "github.com/mniak/godotenv/autoload"

	And bob's your mother's brother
*/

import "github.com/mniak/godotenv"

func init() {
	godotenv.Load()
}
