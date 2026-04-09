package setup

import "sync"

var (
	cfg  Configuration
	once sync.Once
)

// SetConfig sets the global configuration (thread-safe, only once).
func SetConfig(value Configuration) {
	once.Do(func() {
		cfg = value
	})
}

// GetConfig returns the global configuration.
func GetConfig() Configuration {
	return cfg
}

// UpdateConfigForTesting allows updating the config for testing purposes only.
// This should only be used in test files.
func UpdateConfigForTesting(value Configuration) {
	cfg = value
}
