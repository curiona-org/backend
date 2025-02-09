package auth

type Authorizer interface {
	Generate(id int) (string, error)
	Parse(token string) (*Payload, error)
}
