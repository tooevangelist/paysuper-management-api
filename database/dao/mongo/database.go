package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/paysuper/paysuper-management-api/database/dao"
	"sync"
)

type Source struct {
	name           string
	connection     dao.Connection
	session        *mgo.Session
	database       *mgo.Database
	repositoriesMu sync.Mutex
}

func Open(settings dao.Connection) (dao.Database, error) {
	d := &Source{}

	if err := d.Open(settings); err != nil {
		return nil, err
	}

	return d, nil
}

// Open attempts to connect to the database.
func (s *Source) Open(conn dao.Connection) error {
	s.connection = conn
	return s.open()
}

func (s *Source) open() error {
	var err error

	s.session, err = mgo.Dial(s.connection.String())

	if err != nil {
		return err
	}

	s.session.SetMode(mgo.Monotonic, true)

	s.database = s.session.DB("")

	return nil
}

// Close terminates the current database session.
func (s *Source) Close() {
	if s.session != nil {
		s.session.Close()
	}
}

// Driver returns the underlying *mgo.Session instance.
func (s *Source) Driver() interface{} {
	return s.session
}

// Driver returns the underlying *mgo.Database instance.
func (s *Source) Database() interface{} {
	return s.database
}
