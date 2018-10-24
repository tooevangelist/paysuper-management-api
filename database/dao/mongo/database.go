package mongo

import (
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"gopkg.in/mgo.v2"
)

type Source struct {
	name       string
	connection dao.Connection
	session    *mgo.Session
	database   *mgo.Database
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

// Clone returns a cloned db.Database session.
func (s *Source) Clone() (dao.Database, error) {
	newSession := s.session.Copy()

	clone := &Source{
		name:       s.name,
		connection: s.connection,
		session:    newSession,
		database:   newSession.DB(s.database.Name),
	}

	return clone, nil
}

// Source returns specified connection source struct.
func (s *Source) Source() interface{} {
	return s
}
