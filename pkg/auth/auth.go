package auth

type Auth interface {
	GetAuthData() (string, string)
}
