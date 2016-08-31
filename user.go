package main

// User user type.
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserCredentials current user credentials.
type UserCredentials struct {
	Username string
	Password string
}
