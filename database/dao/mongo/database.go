package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/paysuper/paysuper-management-api/database/dao"
	"github.com/paysuper/paysuper-management-api/database/dao/mongo/repository"
	"sync"
)

type Source struct {
	name           string
	connection     dao.Connection
	session        *mgo.Session
	repositories   map[string]*repository.Repository
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

	s.repositories = map[string]*repository.Repository{}
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
		name:         s.name,
		connection:   s.connection,
		session:      newSession,
		database:     newSession.DB(s.database.Name),
		repositories: map[string]*repository.Repository{},
	}

	return clone, nil
}

// Repository returns a repository by name.
func (s *Source) Repository(name string) dao.Repository {
	s.repositoriesMu.Lock()
	defer s.repositoriesMu.Unlock()

	var rep *repository.Repository
	var ok bool

	if rep, ok = s.repositories[name]; !ok {
		c, err := s.Clone()

		if err != nil {
			rep = &repository.Repository{Collection: s.database.C(name)}
		} else {
			rep = &repository.Repository{Collection: c.(*Source).database.C(name)}
		}

		s.repositories[name] = rep
	}

	return rep
}

// Driver returns the underlying *mgo.Session instance.
func (s *Source) Driver() interface{} {
	return s.session
}

// Driver returns the underlying *mgo.Database instance.
func (s *Source) Database() interface{} {
	return s.database
}
