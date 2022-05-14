package helper

import (
	"fmt"
	"os"
	"strconv"
)

// GetStringEnv retrieves the value of the environment variable named by the key.
// It returns the default value if the variable is not present.
func GetStringEnv(key, defaultValue string) string {
	if value, present := os.LookupEnv(key); present {
		return value
	}
	return defaultValue
}

func GetFloat64Env(key string, defaultValue float64) float64 {
	if value, present := os.LookupEnv(key); present {
		if float, err := strconv.ParseFloat(value, 64); err != nil {
			return defaultValue
		} else {
			return float
		}
	}
	return defaultValue
}

// GetBoolEnv retrieves the value of the environment variable named by the key.
// It returns the default value if the variable is not present.
// It accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
// Any other value returns an error.
func GetBoolEnv(key string, defaultValue bool) (bool, error) {
	return strconv.ParseBool(GetStringEnv(key, strconv.FormatBool(defaultValue)))
}

// GetRequiredStringEnv retrieves the value of the environment variable named by the key.
// it will raise panic it not exists
func GetRequiredStringEnv(key string) string {
	value := GetStringEnv(key, "")
	if value == "" {
		panic(fmt.Sprintf("Please set %s environment first", key))
	}
	return value
}

func IsProductionEnvironment() bool {
	return GetStringEnv("ENVIRONMENT", "production") == "production"
}
