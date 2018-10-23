package database

type Database interface {
	Open(*Connection) error
	Close() error
}

type Connection interface {
	String() string
}
