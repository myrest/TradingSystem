package models

type Customer struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	APIKey    string `json:"apikey"`
	SecretKey string `json:"secretkey"`
	IsAdmin   bool   `json:"isadmin"`
}
