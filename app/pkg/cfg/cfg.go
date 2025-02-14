package cfg

import (
	"fmt"
	"os"
	"strings"
)

func MustEnv(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	panic(fmt.Sprintf(`Env "%s" must be provided.`, key))
}

func AssertEnv(keys ...string) map[string]string {
	var missing []string
	asserted := make(map[string]string, len(keys))

	for _, key := range keys {
		v := os.Getenv(key)
		if v == "" {
			missing = append(missing, fmt.Sprintf(`"%s"`, key))
			continue
		}
		asserted[key] = v
	}

	if len(missing) > 0 {
		panic(fmt.Sprintf("Environment variables [%s] must be provided", strings.Join(missing, ", ")))
	}
	return asserted
}

func EnvDefault(key string, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}
