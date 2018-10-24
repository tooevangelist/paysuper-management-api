package dao

type Database interface {
	Open(Connection) error
	Close()
	Repository(string) Repository
}

type Connection interface {
	String() string
}
