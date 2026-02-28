package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Instance struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type Config struct {
	Instances  map[string]Instance
	names      []string
	UserHeader string
	Port       string
}

func (c *Config) InstanceNames() []string {
	return c.names
}

func Load() *Config {
	dbInstances := os.Getenv("DB_INSTANCES")
	if dbInstances == "" {
		log.Fatal("DB_INSTANCES is required")
	}

	names := strings.Split(dbInstances, ",")
	instances := make(map[string]Instance, len(names))

	for _, name := range names {
		upper := strings.ToUpper(name)

		host := requireEnv(upper + "_HOST")
		portStr := requireEnv(upper + "_PORT")
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Fatalf("%s_PORT must be a number", upper)
		}
		user := requireEnv(upper + "_USER")
		password := os.Getenv(upper + "_PASSWORD")
		database := os.Getenv(upper + "_DATABASE")

		instances[name] = Instance{
			Host:     host,
			Port:     port,
			User:     user,
			Password: password,
			Database: database,
		}
	}

	listenPort := os.Getenv("PORT")
	if listenPort == "" {
		listenPort = "80"
	}

	return &Config{
		Instances:  instances,
		names:      names,
		UserHeader: os.Getenv("USER_HEADER"),
		Port:       listenPort,
	}
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s is required", key)
	}
	return val
}
