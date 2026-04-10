package config

import (
	"os"
	"strings"
)

type Config struct {
	Server    ServerConfig
	Cassandra CassandraConfig
	JWT       JWTConfig
	Owner     OwnerConfig
	App       AppConfig
	Dropbox   DropboxConfig
	Email     EmailConfig
}

type ServerConfig struct {
	Port string
}

type CassandraConfig struct {
	Hosts    []string
	Keyspace string
	Username string
	Password string
	UseSSL   bool
}

type JWTConfig struct {
	Secret string
}

type OwnerConfig struct {
	Email string
}

type AppConfig struct {
	Name    string
	BaseURL string
	Org     string
}

type DropboxConfig struct {
	BaseURL string
}

type EmailConfig struct {
	BaseURL string
	Org     string
}

func Load() *Config {
	hosts := os.Getenv("CASSANDRA_HOSTS")
	if hosts == "" {
		hosts = "safer.easipath.com"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "rootrevolution-secret-key-2025-HYIEHE736362373785LK444dd4"
	}

	ownerEmail := os.Getenv("OWNER_EMAIL")
	if ownerEmail == "" {
		ownerEmail = "denise.cochrane78@gmail.com"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://safer.easipath.com"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	useSSL := os.Getenv("CASSANDRA_USE_SSL") == "true"

	return &Config{
		Server: ServerConfig{
			Port: port,
		},
		Cassandra: CassandraConfig{
			Hosts:    strings.Split(hosts, ","),
			Keyspace: "rootrevolution",
			Username: os.Getenv("CASSANDRA_USERNAME"),
			Password: os.Getenv("CASSANDRA_PASSWORD"),
			UseSSL:   useSSL,
		},
		JWT: JWTConfig{
			Secret: jwtSecret,
		},
		Owner: OwnerConfig{
			Email: ownerEmail,
		},
		App: AppConfig{
			Name:    "rootrevolutionapi",
			BaseURL: baseURL,
			Org:     "C10201",
		},
		Dropbox: DropboxConfig{
			BaseURL: "https://cloudcalls.easipath.com/backend-biatechdropbox/api",
		},
		Email: EmailConfig{
			BaseURL: "https://cloudcalls.easipath.com/backend-biatechmailling/api/send-mail/post/within-org",
			Org:     "CM001",
		},
	}
}
