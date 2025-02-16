package oauth

import (
	"context"
	"errors"

	"golang.org/x/oauth2"
	googleOauth "golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

type google struct {
	cfg *oauth2.Config
}

func NewGoogleProvider(clientID, clientSecret string) Client {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     googleOauth.Endpoint,
	}

	return &google{
		cfg: oauthConfig,
	}
}

func (o *google) Verify(ctx context.Context, token string) (User, error) {
	payload, err := idtoken.Validate(ctx, token, o.cfg.ClientID)
	if err != nil {
		return User{}, err
	}

	claims := payload.Claims

	sub, ok := claims["sub"].(string)
	if !ok {
		return User{}, errors.New("missing sub claim")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return User{}, errors.New("missing email claim")
	}

	name, ok := claims["name"].(string)
	if !ok {
		return User{}, errors.New("missing name claim")
	}

	picture, ok := claims["picture"].(string)
	if !ok {
		return User{}, errors.New("missing picture claim")
	}

	user := User{
		Sub:    sub,
		Email:  email,
		Name:   name,
		Avatar: picture,
	}

	return user, nil
}
