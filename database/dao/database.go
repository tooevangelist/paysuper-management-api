package dao

type Database interface {
	Open(Connection) error
	Close()
	Repository(string) Repository
	Driver() interface{}
	Database() interface{}
}

type Connection interface {
	String() string
}
