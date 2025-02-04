package oauth

import (
	"context"

	"github.com/roadmap-thesis/backend/pkg/config"
	"golang.org/x/oauth2"
	googleOauth "golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

type google struct {
	cfg *oauth2.Config
}

func NewGoogleProvider() Client {
	oauthConfig := &oauth2.Config{
		ClientID:     config.GoogleClientID(),
		ClientSecret: config.GoogleClientSecret(),
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

	user := User{
		Sub:    claims["sub"].(string),
		Email:  claims["email"].(string),
		Name:   claims["name"].(string),
		Avatar: claims["picture"].(string),
	}

	return user, nil
}
