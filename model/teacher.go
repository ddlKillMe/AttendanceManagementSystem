package model

type Teacher struct {
	Name     string `redis:"name"`
	TID      string `redis:"tid"`
	Password string `redis:"password"`
}
