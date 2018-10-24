package repository

type Conditions map[string]interface{}

type Handler interface {
	getTable() db.Collection
	GetAll(filters Conditions, limit int, offset int) (error, []interface{})
	GetOne(id string) (error, interface{})
	Add(object *interface{}) (error, interface{})
	Update(id string, object interface{}) (error, *interface{})
	Delete(id string) error
}
