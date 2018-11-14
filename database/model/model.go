package model

const (
	DefaultLimit  = 100
	DefaultOffset = 0
)

type Error struct {
	// text error description
	Message string `json:"message"`
}
