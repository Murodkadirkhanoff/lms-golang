// Package env konfiguratsiyani environment o'zgaruvchilardan o'qish uchun
// default qiymatli helperlar.
package env

import (
	"os"
	"strconv"
	"time"
)

func String(key, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return defaultValue
}

func Int(key string, defaultValue int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultValue
}

func Duration(key string, defaultValue time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultValue
}
