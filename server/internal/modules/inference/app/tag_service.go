package app

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/lib/pq"

	inferencedomain "server/internal/modules/inference/domain"
)

type TagRepository interface {
	FindAll(ctx context.Context, userID int) ([]*inferencedomain.Tag, error)
	FindByID(ctx context.Context, id, userID int) (*inferencedomain.Tag, error)
	Create(ctx context.Context, tag *inferencedomain.Tag) (int, error)
	Update(ctx context.Context, tag *inferencedomain.Tag) error
	Delete(ctx context.Context, id, userID int) error
}

type TagService struct {
	Repo TagRepository
}

var (
	ErrTagNameRequired  = errors.New("name is required")
	ErrTagNameTooLong   = errors.New("name must be <= 128 chars")
	ErrTagNameDuplicate = errors.New("tag name already exists")
	ErrTagNotFound      = sql.ErrNoRows
)

func (s *TagService) GetAll(ctx context.Context, userID int) ([]*inferencedomain.Tag, error) {
	return s.Repo.FindAll(ctx, userID)
}

func (s *TagService) GetByID(ctx context.Context, id, userID int) (*inferencedomain.Tag, error) {
	return s.Repo.FindByID(ctx, id, userID)
}

func (s *TagService) Create(ctx context.Context, userID int, name, groupName, description, user string) (*inferencedomain.Tag, error) {
	normalizedName := normalizeTagName(name)
	if normalizedName == "" {
		return nil, ErrTagNameRequired
	}
	if utf8.RuneCountInString(normalizedName) > 128 {
		return nil, ErrTagNameTooLong
	}
	if err := s.ensureNormalizedNameUnique(ctx, userID, normalizedName, nil); err != nil {
		return nil, err
	}

	tag := &inferencedomain.Tag{
		UserID:      userID,
		Name:        normalizedName,
		GroupName:   strings.TrimSpace(groupName),
		Description: strings.TrimSpace(description),
		CreateUser:  user,
		UpdateUser:  user,
	}
	id, err := s.Repo.Create(ctx, tag)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrTagNameDuplicate
		}
		return nil, err
	}
	return s.Repo.FindByID(ctx, id, userID)
}

func (s *TagService) Update(ctx context.Context, userID, id int, name, groupName, description, user string) (*inferencedomain.Tag, error) {
	normalizedName := normalizeTagName(name)
	if normalizedName == "" {
		return nil, ErrTagNameRequired
	}
	if utf8.RuneCountInString(normalizedName) > 128 {
		return nil, ErrTagNameTooLong
	}
	if err := s.ensureNormalizedNameUnique(ctx, userID, normalizedName, &id); err != nil {
		return nil, err
	}

	tag, err := s.Repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	tag.Name = normalizedName
	tag.GroupName = strings.TrimSpace(groupName)
	tag.Description = strings.TrimSpace(description)
	tag.UpdateUser = user

	if err := s.Repo.Update(ctx, tag); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrTagNameDuplicate
		}
		return nil, err
	}
	return s.Repo.FindByID(ctx, id, userID)
}

func (s *TagService) Delete(ctx context.Context, userID, id int) error {
	return s.Repo.Delete(ctx, id, userID)
}

func normalizeTagName(name string) string {
	fields := strings.Fields(strings.TrimSpace(name))
	if len(fields) == 0 {
		return ""
	}
	return strings.ToLower(strings.Join(fields, " "))
}

func (s *TagService) ensureNormalizedNameUnique(ctx context.Context, userID int, normalizedName string, excludeID *int) error {
	tags, err := s.Repo.FindAll(ctx, userID)
	if err != nil {
		return err
	}
	for _, tag := range tags {
		if excludeID != nil && tag.ID == *excludeID {
			continue
		}
		if normalizeTagName(tag.Name) == normalizedName {
			return ErrTagNameDuplicate
		}
	}
	return nil
}
