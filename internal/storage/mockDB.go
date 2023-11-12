package storage

type Database interface {
	PingDB() error
}

var (
	_ Database = (*MockDB)(nil)
)

type MockDB struct{}

func (*MockDB) PingDB() error {
	return nil
}

func NewMockDB() *MockDB {
	return &MockDB{}
}
