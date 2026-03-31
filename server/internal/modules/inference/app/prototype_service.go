package app

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	inferencedomain "server/internal/modules/inference/domain"
)

type PrototypeRepository interface {
	FindAll(ctx context.Context, userID int) ([]*inferencedomain.Prototype, error)
	FindByID(ctx context.Context, id, userID int) (*inferencedomain.Prototype, error)
	Create(ctx context.Context, prototype *inferencedomain.Prototype) (int, error)
	Update(ctx context.Context, prototype *inferencedomain.Prototype) error
	Delete(ctx context.Context, id, userID int) error
}

type PrototypeTagRepository interface {
	FindByID(ctx context.Context, id, userID int) (*inferencedomain.Tag, error)
}

type PrototypeService struct {
	Repo    PrototypeRepository
	TagRepo PrototypeTagRepository
}

var (
	ErrPrototypeNameRequired       = errors.New("name is required")
	ErrPrototypeNameTooLong        = errors.New("name must be <= 128 chars")
	ErrPrototypeNameDuplicate      = errors.New("prototype name already exists")
	ErrPrototypeTagIDsRequired     = errors.New("tag_ids requires at least one value")
	ErrPrototypeTagIDInvalid       = errors.New("tag ids must be positive")
	ErrPrototypeTagNotFound        = errors.New("specified tag_id was not found")
	ErrPrototypeParentNotFound     = errors.New("parent prototype was not found")
	ErrPrototypeParentNotOwned     = errors.New("parent prototype is not owned by this user")
	ErrPrototypeParentSelfNotAllow = errors.New("prototype cannot reference itself as parent")
)

func (s *PrototypeService) GetAll(ctx context.Context, userID int) ([]*inferencedomain.Prototype, error) {
	return s.Repo.FindAll(ctx, userID)
}

func (s *PrototypeService) GetByID(ctx context.Context, id, userID int) (*inferencedomain.Prototype, error) {
	return s.Repo.FindByID(ctx, id, userID)
}

func (s *PrototypeService) Create(
	ctx context.Context,
	userID int,
	name, description, user string,
	parentPrototypeID *int,
	tagIDs []int,
) (*inferencedomain.Prototype, error) {
	normalizedTagIDs, err := s.validatePrototypeInput(ctx, userID, 0, name, parentPrototypeID, tagIDs)
	if err != nil {
		return nil, err
	}

	prototype := &inferencedomain.Prototype{
		UserID:            userID,
		Name:              name,
		Description:       description,
		ParentPrototypeID: parentPrototypeID,
		TagIDs:            normalizedTagIDs,
		CreateUser:        user,
		UpdateUser:        user,
	}

	id, err := s.Repo.Create(ctx, prototype)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrPrototypeNameDuplicate
		}
		return nil, err
	}

	return s.Repo.FindByID(ctx, id, userID)
}

func (s *PrototypeService) Update(
	ctx context.Context,
	userID, id int,
	name, description, user string,
	parentPrototypeID *int,
	tagIDs []int,
) (*inferencedomain.Prototype, error) {
	normalizedTagIDs, err := s.validatePrototypeInput(ctx, userID, id, name, parentPrototypeID, tagIDs)
	if err != nil {
		return nil, err
	}

	prototype, err := s.Repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	prototype.Name = name
	prototype.Description = description
	prototype.ParentPrototypeID = parentPrototypeID
	prototype.TagIDs = normalizedTagIDs
	prototype.UpdateUser = user

	if err := s.Repo.Update(ctx, prototype); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrPrototypeNameDuplicate
		}
		return nil, err
	}

	return s.Repo.FindByID(ctx, id, userID)
}

func (s *PrototypeService) Delete(ctx context.Context, userID, id int) error {
	return s.Repo.Delete(ctx, id, userID)
}

func (s *PrototypeService) validatePrototypeInput(
	ctx context.Context,
	userID, currentID int,
	name string,
	parentPrototypeID *int,
	tagIDs []int,
) ([]int, error) {
	if name == "" {
		return nil, ErrPrototypeNameRequired
	}
	if len(name) > 128 {
		return nil, ErrPrototypeNameTooLong
	}

	tagIDs = normalizeTargetTagIDs(tagIDs)
	if len(tagIDs) == 0 {
		return nil, ErrPrototypeTagIDsRequired
	}
	for _, tagID := range tagIDs {
		if tagID <= 0 {
			return nil, ErrPrototypeTagIDInvalid
		}
	}
	if err := s.validateTagsOwnership(ctx, userID, tagIDs); err != nil {
		return nil, err
	}

	if parentPrototypeID != nil {
		if *parentPrototypeID <= 0 {
			return nil, ErrPrototypeParentNotFound
		}
		if currentID > 0 && *parentPrototypeID == currentID {
			return nil, ErrPrototypeParentSelfNotAllow
		}
		if _, err := s.Repo.FindByID(ctx, *parentPrototypeID, userID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrPrototypeParentNotOwned
			}
			return nil, err
		}
	}

	return tagIDs, nil
}

func (s *PrototypeService) validateTagsOwnership(ctx context.Context, userID int, tagIDs []int) error {
	if s.TagRepo == nil {
		return nil
	}

	for _, tagID := range tagIDs {
		if _, err := s.TagRepo.FindByID(ctx, tagID, userID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrPrototypeTagNotFound
			}
			return err
		}
	}

	return nil
}
