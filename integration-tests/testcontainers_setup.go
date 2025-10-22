package integration_tests

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupClickHouse - ClickHouse containerını başlatır ve bağlantı döner
func SetupClickHouse(ctx context.Context) (testcontainers.Container, driver.Conn, error) {
	req := testcontainers.ContainerRequest{
		Image:        "clickhouse/clickhouse-server:latest",
		ExposedPorts: []string{"9000/tcp", "8123/tcp"},
		Env: map[string]string{
			"CLICKHOUSE_DB":       "eventstore",
			"CLICKHOUSE_USER":     "admin",
			"CLICKHOUSE_PASSWORD": "admin123",
		},
		WaitingFor: wait.ForListeningPort("9000/tcp").WithStartupTimeout(180 * time.Second),
	}

	// Container'ı başlat
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start clickhouse container: %w", err)
	}

	// Container'ın host ve portunu al
	host, err := container.Host(ctx)
	if err != nil {
		return nil, nil, err
	}

	port, err := container.MappedPort(ctx, "9000")
	if err != nil {
		return nil, nil, err
	}

	// ClickHouse bağlantısı oluştur (retry logic ile)
	var conn driver.Conn
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{fmt.Sprintf("%s:%s", host, port.Port())},
			Auth: clickhouse.Auth{
				Database: "eventstore",
				Username: "admin",
				Password: "admin123",
			},
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			if i == maxRetries-1 {
				return nil, nil, fmt.Errorf("failed to connect to clickhouse after %d retries: %w", maxRetries, err)
			}
			time.Sleep(2 * time.Second)
			continue
		}

		// Ping ile bağlantıyı test et
		if err := conn.Ping(ctx); err != nil {
			if i == maxRetries-1 {
				return nil, nil, fmt.Errorf("failed to ping clickhouse after %d retries: %w", maxRetries, err)
			}
			time.Sleep(2 * time.Second)
			continue
		}

		// Bağlantı başarılı
		break
	}

	// Events tablosunu oluştur
	if err := createEventsTable(ctx, conn); err != nil {
		return nil, nil, fmt.Errorf("failed to create events table: %w", err)
	}

	return container, conn, nil
}

// createEventsTable - events tablosunu oluşturur
func createEventsTable(ctx context.Context, conn driver.Conn) error {
	query := `
		CREATE TABLE IF NOT EXISTS events (
			id String,
			event_type String,
			aggregate_id String,
			payload String,
			timestamp DateTime64(3),
			version UInt32
		) ENGINE = MergeTree()
		ORDER BY (timestamp, id)
	`
	return conn.Exec(ctx, query)
}

// SetupKafka - Kafka containerını başlatır ve broker adresini döner
func SetupKafka(ctx context.Context) (testcontainers.Container, string, error) {
	kafkaContainer, err := kafka.RunContainer(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to start kafka container: %w", err)
	}

	// Broker adresini al
	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get kafka brokers: %w", err)
	}

	if len(brokers) == 0 {
		return nil, "", fmt.Errorf("no kafka brokers available")
	}

	broker := brokers[0]

	return kafkaContainer, broker, nil
}
