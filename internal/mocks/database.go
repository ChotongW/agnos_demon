package mocks

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
)

// MockDB is a mock implementation of database.DB
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	ret := m.Called(ctx, sql, args)
	return ret.Get(0).(pgx.Rows), ret.Error(1)
}

func (m *MockDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	ret := m.Called(ctx, sql, args)
	return ret.Get(0).(pgx.Row)
}

func (m *MockDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	ret := m.Called(ctx, sql, args)
	// Return empty CommandTag if nil, or the actual tag
	if ret.Get(0) == nil {
		return pgconn.CommandTag{}, ret.Error(1)
	}
	return ret.Get(0).(pgconn.CommandTag), ret.Error(1)
}

func (m *MockDB) Close() {
	m.Called()
}

func (m *MockDB) Begin(ctx context.Context) (pgx.Tx, error) {
	ret := m.Called(ctx)
	// If we return nil for Tx, it might panic if used, but it satisfies the interface signature
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(pgx.Tx), ret.Error(1)
}

// MockRow is a mock implementation of pgx.Row
type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	// Combine arguments into a single slice to pass to Called
	// This allows us to match any number of arguments in the test expectation
	args := make([]interface{}, len(dest))
	for i, v := range dest {
		args[i] = v
	}

	// We pass the pointers themselves to Called so the mock action can modify them
	ret := m.Called(dest...)
	return ret.Error(0)
}

// MockRows is a mock implementation of pgx.Rows
type MockRows struct {
	mock.Mock
}

func (m *MockRows) Close() {
	m.Called()
}

func (m *MockRows) Err() error {
	ret := m.Called()
	return ret.Error(0)
}

func (m *MockRows) CommandTag() pgconn.CommandTag {
	ret := m.Called()
	return ret.Get(0).(pgconn.CommandTag)
}

func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	ret := m.Called()
	return ret.Get(0).([]pgconn.FieldDescription)
}

func (m *MockRows) Next() bool {
	ret := m.Called()
	return ret.Bool(0)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	ret := m.Called(dest...)
	return ret.Error(0)
}

func (m *MockRows) Values() ([]interface{}, error) {
	ret := m.Called()
	return ret.Get(0).([]interface{}), ret.Error(1)
}

func (m *MockRows) RawValues() [][]byte {
	ret := m.Called()
	return ret.Get(0).([][]byte)
}

func (m *MockRows) Conn() *pgx.Conn {
	ret := m.Called()
	return ret.Get(0).(*pgx.Conn)
}
