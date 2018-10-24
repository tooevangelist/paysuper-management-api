package dao

type Database interface {
	Open(Connection) error
	Close()
}

type Connection interface {
	String() string
}
