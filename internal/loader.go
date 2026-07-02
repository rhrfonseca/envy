package envy

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func LoadEnvMap(envName string, dotenvFiles []string) (map[string]string, error) {
	result := make(map[string]string)

	if len(dotenvFiles) > 0 {
		for i := len(dotenvFiles) - 1; i >= 0; i-- {
			m, err := godotenv.Read(dotenvFiles[i])
			if err != nil {
				return nil, fmt.Errorf("reading dotenv file %q: %w", dotenvFiles[i], err)
			}
			for k, v := range m {
				result[k] = v
			}
		}
	} else {
		candidates := []string{".env"}
		if envName != "" {
			candidates = append(candidates, ".env."+envName)
		}
		candidates = append(candidates, ".env.local")
		if envName != "" {
			candidates = append(candidates, ".env."+envName+".local")
		}

		for _, f := range candidates {
			m, err := godotenv.Read(f)
			if err != nil {
				continue
			}
			for k, v := range m {
				result[k] = v
			}
		}
	}

	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result, nil
}
