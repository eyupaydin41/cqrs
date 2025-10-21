package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func InitClickHouse() driver.Conn {
	LoadEnv()

	host := GetEnv("CLICKHOUSE_HOST")
	user := GetEnv("CLICKHOUSE_USER")
	fmt.Print("user", user)
	password := GetEnv("CLICKHOUSE_PASSWORD")
	database := GetEnv("CLICKHOUSE_DB")

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{host},
		Auth: clickhouse.Auth{
			Database: database,
			Username: user,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout:      time.Second * 30,
		MaxOpenConns:     5,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Hour,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	})

	if err != nil {
		log.Fatalf("failed to connect to ClickHouse: %v", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping ClickHouse: %v", err)
	}

	log.Println("connected to ClickHouse successfully")

	if err := createEventTable(conn); err != nil {
		log.Fatalf("failed to create event table: %v", err)
	}

	return conn
}

func createEventTable(conn driver.Conn) error {
	ctx := context.Background()
	query := `
		CREATE TABLE IF NOT EXISTS events (
			id String,
			event_type String,
			aggregate_id String,
			payload String,
			timestamp DateTime64(3),
			version UInt32,
			INDEX idx_event_type event_type TYPE minmax GRANULARITY 4,
			INDEX idx_aggregate_id aggregate_id TYPE minmax GRANULARITY 4,
			INDEX idx_timestamp timestamp TYPE minmax GRANULARITY 1
		) ENGINE = MergeTree()
		ORDER BY (timestamp, id)
		PARTITION BY toYYYYMM(timestamp)
		SETTINGS index_granularity = 8192
	`

	if err := conn.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	log.Println("event table created or already exists")
	return nil
}
