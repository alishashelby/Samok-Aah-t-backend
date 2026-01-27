package option

import "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"

func Any(key string, value any) logger.LogOption {
	return func(m map[string]any) {
		m[key] = value
	}
}

func Error(err error) logger.LogOption {
	return func(m map[string]any) {
		m["error"] = err.Error()
	}
}
