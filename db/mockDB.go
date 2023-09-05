package db

import (
	"github.com/rs/zerolog/log"

	"github.com/sch8ill/gscrawler/types"
)

// a mock version of db.DB
// implements db.DBI
type MockDB struct{}

func NewMockDB() DBI {
	return &MockDB{}
}

// pretends to connect to a database
func (m *MockDB) Connect() error {
	log.Debug().Msg("Connected to mock database")
	return nil
}

// pretends to close the database client
func (m *MockDB) Close() error {
	return nil
}

// pretends to insert a site into the database
func (m *MockDB) InsertSite(_ types.Site) error {
	return nil
}
