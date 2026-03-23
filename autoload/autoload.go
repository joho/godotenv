package autoload

/*
	You can just read the .env file on import just by doing

		import _ "github.com/joho/godotenv/autoload"

	And bob's your mother's brother
*/

import "github.com/joho/godotenv"

var filenames = []string{".env", ".env.production", ".env.development", ".env.test", ".env.local"}

func init() {
	for index := range filenames {
		if _, err := os.Stat(filenames[index]); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			log.Println(err)
			return
		}

		if err := godotenv.Load(filenames[index]); err != nil {
			log.Println(err)
			return
		}

		return
	}
}
