package cassandra

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/gocql/gocql"

	"rootrevolution-api/config"
)

func Connect(cfg *config.Config) (*gocql.Session, error) {
	cluster := gocql.NewCluster(cfg.Cassandra.Hosts...)
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second
	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}

	if cfg.Cassandra.Username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Cassandra.Username,
			Password: cfg.Cassandra.Password,
		}
	}

	if cfg.Cassandra.UseSSL {
		cluster.SslOpts = &gocql.SslOptions{
			Config: &tls.Config{
				InsecureSkipVerify: false,
			},
		}
	}

	// Connect without keyspace first to create it
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("connecting to cassandra: %w", err)
	}

	// Create keyspace
	createKS := fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '1'}
		AND durable_writes = true`, cfg.Cassandra.Keyspace)

	if err := session.Query(createKS).Exec(); err != nil {
		session.Close()
		return nil, fmt.Errorf("creating keyspace: %w", err)
	}
	session.Close()

	// Reconnect with keyspace
	cluster.Keyspace = cfg.Cassandra.Keyspace
	session, err = cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("connecting to keyspace: %w", err)
	}

	return session, nil
}
