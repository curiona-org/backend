package domain

import (
	"errors"
	"html"
	"time"
)

const (
	ProfileTable = "profiles"

	DefaultMaxGeneratedRoadmaps = 5 // Default maximum number of generated roadmaps per profile
)

var (
	ErrProfileNotFound = errors.New("profile not found")
)

type Profile struct {
	ID     int
	Name   string
	Avatar string

	MaxGeneratedRoadmaps int

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewProfile(name, avatar string) *Profile {
	if avatar == "" {
		avatar = getDefaultAvatar(name)
	}

	return &Profile{
		Name:                 name,
		Avatar:               avatar,
		MaxGeneratedRoadmaps: DefaultMaxGeneratedRoadmaps,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
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
