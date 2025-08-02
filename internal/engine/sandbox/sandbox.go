package sandbox

import (
	"context"
	"fmt"
	"github.com/apache/arrow-adbc/go/adbc"
	"github.com/apache/arrow-adbc/go/adbc/drivermgr"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/turbolytics/sqlsec/internal/db/queries/events"
	"github.com/turbolytics/sqlsec/internal/db/queries/rules"
	"go.uber.org/zap"
)

func arrowRecordFromEvent(event *events.Event) arrow.Record {
	pool := memory.NewGoAllocator()

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.BinaryTypes.String},
		{Name: "webhook_id", Type: arrow.BinaryTypes.String},
		{Name: "source", Type: arrow.BinaryTypes.String},
		{Name: "event_type", Type: arrow.BinaryTypes.String},
		{Name: "action", Type: arrow.BinaryTypes.String},
		{Name: "raw_payload", Type: arrow.BinaryTypes.String},
		{Name: "dedup_hash", Type: arrow.BinaryTypes.String},
		{Name: "received_at", Type: arrow.FixedWidthTypes.Timestamp_ms},
	}, nil)

	b := array.NewRecordBuilder(pool, schema)
	defer b.Release()

	payload, _ := event.RawPayload.MarshalJSON()

	b.Field(0).(*array.StringBuilder).Append(event.ID.String())
	b.Field(1).(*array.StringBuilder).Append(event.WebhookID.String())
	b.Field(2).(*array.StringBuilder).Append(event.Source)
	b.Field(3).(*array.StringBuilder).Append(event.EventType)
	b.Field(4).(*array.StringBuilder).Append(event.Action.String)
	b.Field(5).(*array.StringBuilder).Append(string(payload))
	b.Field(6).(*array.StringBuilder).Append(event.DedupHash.String)
	b.Field(7).(*array.TimestampBuilder).Append(arrow.Timestamp(event.ReceivedAt.Time.UnixMilli()))

	return b.NewRecord()
}

// Sandbox provides a controlled environment for executing sql queries and processing events.
type Sandbox struct {
	conn adbc.Connection

	logger *zap.Logger
}

type SandboxOption func(*Sandbox)

func WithLogger(logger *zap.Logger) SandboxOption {
	return func(s *Sandbox) {
		s.logger = logger
	}
}

func WithDuckDBMemoryConnection() SandboxOption {
	return func(s *Sandbox) {
		var drv drivermgr.Driver
		db, err := drv.NewDatabase(map[string]string{
			"driver":     "duckdb",
			"entrypoint": "duckdb_adbc_init",
			"path":       ":memory:",
		})
		if err != nil {
			panic(err) // Handle error appropriately in production code
		}

		conn, err := db.Open(nil)
		if err != nil {
			panic(err) // Handle error appropriately in production code
		}

		s.conn = conn
	}
}

func New(ctx context.Context, opts ...SandboxOption) (*Sandbox, error) {
	s := &Sandbox{
		logger: zap.NewNop(),
	}

	for _, opt := range opts {
		opt(s)
	}

	if err := s.InitTables(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Sandbox) InitTables(ctx context.Context) error {
	createTableSQL := `
	CREATE TABLE events (
		id UUID UNIQUE NOT NULL PRIMARY KEY,
		webhook_id UUID,
		source TEXT,
		event_type TEXT,
		action TEXT,
		raw_payload JSON,
		dedup_hash TEXT,
		received_at TIMESTAMP
	);`

	createTableStmt, err := s.conn.NewStatement()
	if err != nil {
		return fmt.Errorf("create statement: %w", err)
	}
	defer func() {
		if err := createTableStmt.Close(); err != nil {
			fmt.Printf("failed to close create table statement: %v\n", err)
		}
	}()
	if err := createTableStmt.SetSqlQuery(createTableSQL); err != nil {
		return fmt.Errorf("set query error: %v", err)
	}
	_, err = createTableStmt.ExecuteUpdate(ctx)
	if err != nil {
		return fmt.Errorf("query execution error: %v", err)
	}

	return nil
}

func (s *Sandbox) AddEvent(ctx context.Context, event *events.Event) error {
	insertSQL := `INSERT INTO events (
		id, 
		webhook_id, 
		source, event_type, action, raw_payload, dedup_hash, received_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	insertStmt, err := s.conn.NewStatement()
	if err != nil {
		return fmt.Errorf("new insert stmt: %w", err)
	}
	defer insertStmt.Close()

	if err := insertStmt.SetSqlQuery(insertSQL); err != nil {
		return fmt.Errorf("set insert sql: %w", err)
	}

	r := arrowRecordFromEvent(event)
	if err := insertStmt.Bind(ctx, r); err != nil {
		return fmt.Errorf("bind insert: %w", err)
	}

	if _, err := insertStmt.ExecuteUpdate(ctx); err != nil {
		return fmt.Errorf("execute insert: %w", err)
	}
	return nil
}

func (s *Sandbox) ExecuteRule(ctx context.Context, rule rules.Rule) (int, error) {
	if rule.Sql == "" {
		return -1, fmt.Errorf("rule SQL is empty for rule ID %s", rule.ID)
	}

	ruleStmt, err := s.conn.NewStatement()
	if err != nil {
		return -1, fmt.Errorf("new rule stmt: %w", err)
	}
	defer ruleStmt.Close()

	s.logger.Debug("Executing rule SQL",
		zap.String("rule_id", rule.ID.String()),
	)

	if err := ruleStmt.SetSqlQuery(rule.Sql); err != nil {
		return -1, fmt.Errorf("sandbox: set sql query: %w", err)
	}
	reader, _, err := ruleStmt.ExecuteQuery(ctx)
	if err != nil {
		return -1, fmt.Errorf("execute rule sql: %w", err)
	}
	defer reader.Release()

	// 4. Count rows
	count := 0
	for reader.Next() {
		reader.Release()
		count++
	}
	if err := reader.Err(); err != nil {
		return -1, fmt.Errorf("read rule results: %w", err)
	}

	return count, nil
}

func (s *Sandbox) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
