package dao

type Database interface {
	Open(Connection) error
	Close()
	Driver() interface{}
	Database() interface{}
}

type Connection interface {
	String() string
}
