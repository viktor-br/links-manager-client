package main

// User user type.
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserAuth user authentication info.
type UserAuth struct {
	User     *UserAuthUser `json:"user"`
	Username string        `json:"username"`
	Expires  int64         `json:"expires"`
}

// UserAuthUser user authentication specific details of a user.
// Separate User and UserAuthUser to avoid passing user ID
// during user creation to remote server.
type UserAuthUser struct {
	ID string `json:"id"`
}

// UserCredentials current user credentials.
type UserCredentials struct {
	Username string
	Password string
}
