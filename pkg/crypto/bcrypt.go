package crypto

import "golang.org/x/crypto/bcrypt"

// BcryptHash hashes the password.
func BcryptHash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), 10)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// BcryptCompare compares the password with the hash.
func BcryptCompare(hash, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}
