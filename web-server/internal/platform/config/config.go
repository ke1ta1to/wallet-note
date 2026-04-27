package config

import (
	"cmp"
	"os"
)

type Config struct {
	Port      string
	TableName string
}

func Load() Config {
	return Config{
		Port:      cmp.Or(os.Getenv("PORT"), "8080"),
		TableName: os.Getenv("WALLET_NOTE_TABLE"),
	}
}
