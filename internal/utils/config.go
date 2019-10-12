package utils

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type httpEntry struct {
	Ip       string
	Port     int
	Timeout  int
	UserName string
	Password string
}

type dbEntry struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type otherEntry struct {
	CascadePath string
}

type Config struct {
	Http     httpEntry
	Database dbEntry
	Other    otherEntry
}

var Configuration Config

func (config *Config) Load(filename string) error {
	_, err := toml.DecodeFile(filename, config)
	return err
}

// GetConnectionString - get connection string for PostgreSQL
func (config *Config) GetConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Database.Host, config.Database.Port, config.Database.User, config.Database.Password, config.Database.Database)
}
