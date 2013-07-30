package godotenv

import (
	"os"
	"testing"
)

func TestLoadPlainEnv(t *testing.T) {
	envFileName := "fixtures/plain.env"
	err := Load(envFileName)
	if err != nil {
		t.Fatalf("Error loading %v", envFileName)
	}

	plainValues := map[string]string{
		"OPTION_A": "1",
		"OPTION_B": "2",
		"OPTION_C": "3",
		"OPTION_D": "4",
		"OPTION_E": "5",
	}

	for k := range plainValues {
		envValue := os.Getenv(k)
		v := plainValues[k]
		if envValue != v {
			t.Errorf("Mismatch for key '%v': expected '%v' got '%v'", k, v, envValue)
		}
	}
}
