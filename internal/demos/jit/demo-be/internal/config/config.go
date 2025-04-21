package config

import "os"

type Config struct {
	AtlasPublicKey    string
	AtlasPrivateKey   string
	AtlasProjectID    string
	TemporalHost      string
	TemporalNamespace string
	Port              string
}

func LoadConfig() *Config {
	cfg := &Config{
		AtlasPublicKey:    os.Getenv("ATLAS_PUBLIC_KEY"),
		AtlasPrivateKey:   os.Getenv("ATLAS_PRIVATE_KEY"),
		AtlasProjectID:    os.Getenv("ATLAS_PROJECT_ID"),
		TemporalHost:      os.Getenv("TEMPORAL_HOST"),
		TemporalNamespace: os.Getenv("TEMPORAL_NAMESPACE"),
		Port:              os.Getenv("PORT"),
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.TemporalHost == "" {
		cfg.TemporalHost = "localhost:7233"
	}
	if cfg.TemporalNamespace == "" {
		cfg.TemporalNamespace = "default"
	}
	return cfg
}
