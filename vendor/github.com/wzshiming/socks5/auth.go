package socks5

// AuthenticationFunc Authentication interface is implemented
type AuthenticationFunc func(cmd Command, username, password string) bool

// Auth authentication processing
func (f AuthenticationFunc) Auth(cmd Command, username, password string) bool {
	return f(cmd, username, password)
}

// Authentication proxy authentication
type Authentication interface {
	Auth(cmd Command, username, password string) bool
}

// UserAuth basic authentication
func UserAuth(username, password string) Authentication {
	return AuthenticationFunc(func(c Command, u, p string) bool {
		return username == u && password == p
	})
}
