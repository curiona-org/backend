package domain

import (
	"errors"
	"html"
	"time"
)

const (
	ProfileTable = "profiles"
)

var (
	ErrProfileNotFound = errors.New("profile not found")
)

type Profile struct {
	ID     int
	Name   string
	Avatar string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewProfile(name, avatar string) *Profile {
	if avatar == "" {
		avatar = getDefaultAvatar(name)
	}

	return &Profile{
		Name:      name,
		Avatar:    avatar,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (p *Profile) Update(name string) {
	p.Name = name
	p.UpdatedAt = time.Now()
}

func (p *Profile) IsZero() bool {
	return p.ID == 0 &&
		p.Name == "" &&
		p.Avatar == "" &&
		p.CreatedAt.IsZero() &&
		p.UpdatedAt.IsZero()
}

func getDefaultAvatar(name string) string {
	return "https://hostedboringavatars.vercel.app/api/beam?colors=1DA1F2,14171A,657786,F5F8FA&name=" + html.EscapeString(name)
}
