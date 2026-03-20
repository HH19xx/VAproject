package usecases

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/lib/pq"

	"server/main/domain"
)

type TagRepository interface {
	FindAll(ctx context.Context, userID int) ([]*domain.Tag, error)
	FindByID(ctx context.Context, id, userID int) (*domain.Tag, error)
	Create(ctx context.Context, tag *domain.Tag) (int, error)
	Update(ctx context.Context, tag *domain.Tag) error
	Delete(ctx context.Context, id, userID int) error
}

type TagUsecase struct {
	Repo TagRepository
}

var (
	ErrTagNameRequired  = errors.New("name は必須です")
	ErrTagNameTooLong   = errors.New("name は128文字以内で入力してください")
	ErrTagNameDuplicate = errors.New("同名タグが既に存在します")
)

func (u *TagUsecase) GetAll(ctx context.Context, userID int) ([]*domain.Tag, error) {
	return u.Repo.FindAll(ctx, userID)
}

func (u *TagUsecase) GetByID(ctx context.Context, id, userID int) (*domain.Tag, error) {
	return u.Repo.FindByID(ctx, id, userID)
}

func (u *TagUsecase) Create(ctx context.Context, userID int, name, groupName, description, user string) (*domain.Tag, error) {
	normalizedName := normalizeTagName(name)
	if normalizedName == "" {
		return nil, ErrTagNameRequired
	}
	if utf8.RuneCountInString(normalizedName) > 128 {
		return nil, ErrTagNameTooLong
	}
	if err := u.ensureNormalizedNameUnique(ctx, userID, normalizedName, nil); err != nil {
		return nil, err
	}

	tag := &domain.Tag{
		UserID:      userID,
		Name:        normalizedName,
		GroupName:   strings.TrimSpace(groupName),
		Description: strings.TrimSpace(description),
		CreateUser:  user,
		UpdateUser:  user,
	}
	id, err := u.Repo.Create(ctx, tag)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrTagNameDuplicate
		}
		return nil, err
	}
	return u.Repo.FindByID(ctx, id, userID)
}

func (u *TagUsecase) Update(ctx context.Context, userID, id int, name, groupName, description, user string) (*domain.Tag, error) {
	normalizedName := normalizeTagName(name)
	if normalizedName == "" {
		return nil, ErrTagNameRequired
	}
	if utf8.RuneCountInString(normalizedName) > 128 {
		return nil, ErrTagNameTooLong
	}
	if err := u.ensureNormalizedNameUnique(ctx, userID, normalizedName, &id); err != nil {
		return nil, err
	}

	tag, err := u.Repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	tag.Name = normalizedName
	tag.GroupName = strings.TrimSpace(groupName)
	tag.Description = strings.TrimSpace(description)
	tag.UpdateUser = user

	if err := u.Repo.Update(ctx, tag); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrTagNameDuplicate
		}
		return nil, err
	}
	return u.Repo.FindByID(ctx, id, userID)
}

func (u *TagUsecase) Delete(ctx context.Context, userID, id int) error {
	return u.Repo.Delete(ctx, id, userID)
}

var ErrTagNotFound = sql.ErrNoRows

func normalizeTagName(name string) string {
	fields := strings.Fields(strings.TrimSpace(name))
	if len(fields) == 0 {
		return ""
	}
	return strings.ToLower(strings.Join(fields, " "))
}

func (u *TagUsecase) ensureNormalizedNameUnique(ctx context.Context, userID int, normalizedName string, excludeID *int) error {
	tags, err := u.Repo.FindAll(ctx, userID)
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
