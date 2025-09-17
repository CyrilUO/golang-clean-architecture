package entities

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
}
