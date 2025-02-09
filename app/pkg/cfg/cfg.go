package cfg

import (
	"fmt"
	"os"
)

func MustEnv(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	panic(fmt.Sprintf(`Env "%s" must be provided.`, key))
}
func EnvDefault(key string, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}
