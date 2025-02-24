package auth

// Method represents the sign-in method.
type Method string

const (
	// MethodEmail uses email & password.
	MethodEmail Method = "email"
	// MethodGoogle uses Google OAuth.
	MethodGoogle Method = "google"
)

func (m Method) IsEmail() bool {
	return m == MethodEmail
}

func (m Method) IsGoogle() bool {
	return m == MethodGoogle
}

func (m Method) String() string {
	return string(m)
}
