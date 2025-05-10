package cfgloader

import (
	"context"
	"fmt"
	"micro-mart/pkg/mmotel"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// LoadConfigFromEnv loads the configuration from the environment variables.
// The configuration should be a struct with the `env` tag.
//
// Note:
// Now only supports `string` and `int` types.
//
// Usage:
//
//	type Config struct {
//	    Host string `env:"HOST"`  //  get the value of the environment variable `HOST`
//	    Port int `env:"PORT"` //  get the value of the environment variable `PORT`
//	}
func LoadConfigFromEnv[T any](serviceName string) (*T, error) {
	var result T

	// Try to load from different possible locations
	envFile := ".env"
	env := os.Getenv("ENV")
	if env != "" {
		envFile = fmt.Sprintf(".env.%s", env)
	}

	// Try current directory first
	err := godotenv.Load(envFile)
	if err != nil {
		// Try service directory
		serviceEnvPath := fmt.Sprintf("services/%s/%s", serviceName, envFile)
		err = godotenv.Load(serviceEnvPath)
		if err != nil {
			// Try absolute path if provided
			if envPath := os.Getenv("ENV_FILE_PATH"); envPath != "" {
				err = godotenv.Load(envPath)
				if err != nil {
					mmotel.Error(context.Background(), fmt.Sprintf("Failed to load .env file from all locations: %s", err.Error()))
					return nil, fmt.Errorf("failed to load environment variables: %w", err)
				}
			} else {
				mmotel.Error(context.Background(), fmt.Sprintf("Failed to load .env file: %s", err.Error()))
				return nil, fmt.Errorf("failed to load environment variables: %w", err)
			}
		}
	}

	err = loadFromEnvironment(reflect.ValueOf(&result).Elem())
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// loadFromEnvironment use reflection to load the configuration from the environment variables.
func loadFromEnvironment(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.Kind() == reflect.Struct {
			err := loadFromEnvironment(field)
			if err != nil {
				return err
			}
			continue
		}

		envKey := fieldType.Tag.Get("env")
		if envKey == "" {
			continue
		}

		value := os.Getenv(envKey)
		if value == "" {
			return fmt.Errorf("missing environment variable: %s", envKey)
		}

		err := setFieldValue(field, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int:
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		field.SetInt(int64(intValue))
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	case reflect.Slice:
		arrayValue := strings.Split(value, ",")
		slice := reflect.MakeSlice(field.Type(), len(arrayValue), len(arrayValue))
		for i, str := range arrayValue {
			elem := slice.Index(i)
			if err := setFieldValue(elem, str); err != nil {
				return err
			}
		}
		field.Set(slice)
	default:
		return fmt.Errorf("unsupported type: %s", field.Kind())
	}

	return nil
}
