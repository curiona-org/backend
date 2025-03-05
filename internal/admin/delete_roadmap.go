package admin

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
)

func (app *adminApplication) DeleteRoadmap(ctx context.Context, roadmapID int) error {
	err := app.repository.Roadmap.UpdateByID(ctx, roadmapID, func(roadmap *domain.Roadmap) (bool, error) {
		roadmap.Delete()
		return true, nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrRoadmapNotFound) {
			return cerrors.ErrNotFound.Msg("roadmap")
		}
		return err
	}

	return nil
}
