package app

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	inferencedomain "server/internal/modules/inference/domain"
)

type TargetRepository interface {
	FindAll(ctx context.Context, userID, limit, offset int) ([]*inferencedomain.Target, error)
	FindByID(ctx context.Context, id, userID int) (*inferencedomain.Target, error)
	Create(ctx context.Context, target *inferencedomain.Target) (int, error)
	Update(ctx context.Context, target *inferencedomain.Target) error
	Delete(ctx context.Context, id, userID int) error
	Count(ctx context.Context, userID int) (int, error)
	FindActionLogsByTargetID(ctx context.Context, targetID, userID, limit, offset int) ([]*inferencedomain.ActionLog, error)
	CountActionLogsByTargetID(ctx context.Context, targetID, userID int) (int, error)
}

type TargetTagRepository interface {
	FindByID(ctx context.Context, id, userID int) (*inferencedomain.Tag, error)
}

type TargetService struct {
	Repo    TargetRepository
	TagRepo TargetTagRepository
}

var (
	ErrTargetNameRequired     = errors.New("target name is required")
	ErrTargetNameTooLong      = errors.New("target name must be <= 64 chars")
	ErrTargetNameDuplicate    = errors.New("target name already exists")
	ErrTargetTagIDsRequired   = errors.New("tag_ids requires at least one value")
	ErrTargetTagIDInvalid     = errors.New("tag ids must be positive")
	ErrTargetTagNotFound      = errors.New("specified tag_id was not found")
	ErrTargetMatchModeInvalid = errors.New("match_mode must be AND")
	DefaultTargetMatchModeAND = "AND"
)

func (s *TargetService) GetTargets(ctx context.Context, userID, page, limit int) ([]*inferencedomain.Target, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	targets, err := s.Repo.FindAll(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.Repo.Count(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return targets, total, nil
}

func (s *TargetService) GetTargetByID(ctx context.Context, id, userID int) (*inferencedomain.Target, error) {
	return s.Repo.FindByID(ctx, id, userID)
}

func (s *TargetService) CreateTarget(
	ctx context.Context,
	userID int,
	name, description, createUser, matchMode, queryText string,
	tagIDs, anyTagIDs, excludeTagIDs []int,
	anyTagGroups [][]int,
) (*inferencedomain.Target, error) {
	normalizedTagIDs, normalizedAnyTagIDs, normalizedExcludeTagIDs, normalizedAnyTagGroups, err :=
		s.validateTargetInput(ctx, userID, name, matchMode, tagIDs, anyTagIDs, excludeTagIDs, anyTagGroups)
	if err != nil {
		return nil, err
	}

	target := &inferencedomain.Target{
		UserID:        userID,
		Name:          name,
		Description:   description,
		MatchMode:     DefaultTargetMatchModeAND,
		QueryText:     queryText,
		TagIDs:        normalizedTagIDs,
		AnyTagIDs:     normalizedAnyTagIDs,
		AnyTagGroups:  normalizedAnyTagGroups,
		ExcludeTagIDs: normalizedExcludeTagIDs,
		CreateUser:    createUser,
		UpdateUser:    createUser,
	}

	id, err := s.Repo.Create(ctx, target)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrTargetNameDuplicate
		}
		return nil, err
	}

	return s.Repo.FindByID(ctx, id, userID)
}

func (s *TargetService) UpdateTarget(
	ctx context.Context,
	userID, id int,
	name, description, updateUser, matchMode, queryText string,
	tagIDs, anyTagIDs, excludeTagIDs []int,
	anyTagGroups [][]int,
) (*inferencedomain.Target, error) {
	normalizedTagIDs, normalizedAnyTagIDs, normalizedExcludeTagIDs, normalizedAnyTagGroups, err :=
		s.validateTargetInput(ctx, userID, name, matchMode, tagIDs, anyTagIDs, excludeTagIDs, anyTagGroups)
	if err != nil {
		return nil, err
	}

	target, err := s.Repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	target.Name = name
	target.Description = description
	target.MatchMode = DefaultTargetMatchModeAND
	target.QueryText = queryText
	target.TagIDs = normalizedTagIDs
	target.AnyTagIDs = normalizedAnyTagIDs
	target.AnyTagGroups = normalizedAnyTagGroups
	target.ExcludeTagIDs = normalizedExcludeTagIDs
	target.UpdateUser = updateUser

	if err := s.Repo.Update(ctx, target); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrTargetNameDuplicate
		}
		return nil, err
	}

	return s.Repo.FindByID(ctx, id, userID)
}

func (s *TargetService) DeleteTarget(ctx context.Context, userID, id int) error {
	return s.Repo.Delete(ctx, id, userID)
}

func (s *TargetService) GetActionLogsByTargetID(ctx context.Context, userID, targetID, page, limit int) ([]*inferencedomain.ActionLog, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	if _, err := s.Repo.FindByID(ctx, targetID, userID); err != nil {
		return nil, 0, err
	}

	logs, err := s.Repo.FindActionLogsByTargetID(ctx, targetID, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.Repo.CountActionLogsByTargetID(ctx, targetID, userID)
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (s *TargetService) validateTargetInput(
	ctx context.Context,
	userID int,
	name, matchMode string,
	tagIDs, anyTagIDs, excludeTagIDs []int,
	anyTagGroups [][]int,
) ([]int, []int, []int, [][]int, error) {
	if name == "" {
		return nil, nil, nil, nil, ErrTargetNameRequired
	}
	if len(name) > 64 {
		return nil, nil, nil, nil, ErrTargetNameTooLong
	}
	if matchMode != "" && matchMode != DefaultTargetMatchModeAND {
		return nil, nil, nil, nil, ErrTargetMatchModeInvalid
	}

	normalizedTagIDs := normalizeTargetTagIDs(tagIDs)
	if len(normalizedTagIDs) == 0 {
		return nil, nil, nil, nil, ErrTargetTagIDsRequired
	}

	normalizedAnyTagIDs := normalizeTargetTagIDs(anyTagIDs)
	normalizedExcludeTagIDs := normalizeTargetTagIDs(excludeTagIDs)
	normalizedAnyTagGroups := normalizeTargetTagGroups(anyTagGroups)

	allTagIDs := make([]int, 0, len(normalizedTagIDs)+len(normalizedAnyTagIDs)+len(normalizedExcludeTagIDs))
	allTagIDs = append(allTagIDs, normalizedTagIDs...)
	allTagIDs = append(allTagIDs, normalizedAnyTagIDs...)
	allTagIDs = append(allTagIDs, normalizedExcludeTagIDs...)
	for _, group := range normalizedAnyTagGroups {
		allTagIDs = append(allTagIDs, group...)
	}
	allTagIDs = normalizeTargetTagIDs(allTagIDs)

	for _, tagID := range allTagIDs {
		if tagID <= 0 {
			return nil, nil, nil, nil, ErrTargetTagIDInvalid
		}
	}

	if err := s.validateTagsOwnership(ctx, userID, allTagIDs); err != nil {
		return nil, nil, nil, nil, err
	}

	return normalizedTagIDs, normalizedAnyTagIDs, normalizedExcludeTagIDs, normalizedAnyTagGroups, nil
}

func (s *TargetService) validateTagsOwnership(ctx context.Context, userID int, tagIDs []int) error {
	if s.TagRepo == nil {
		return nil
	}

	for _, tagID := range tagIDs {
		if _, err := s.TagRepo.FindByID(ctx, tagID, userID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrTargetTagNotFound
			}
			return err
		}
	}

	return nil
}

func normalizeTargetTagIDs(tagIDs []int) []int {
	if len(tagIDs) == 0 {
		return nil
	}

	seen := make(map[int]struct{}, len(tagIDs))
	out := make([]int, 0, len(tagIDs))
	for _, id := range tagIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}

	return out
}

func normalizeTargetTagGroups(groups [][]int) [][]int {
	if len(groups) == 0 {
		return nil
	}

	out := make([][]int, 0, len(groups))
	for _, group := range groups {
		normalized := normalizeTargetTagIDs(group)
		if len(normalized) == 0 {
			continue
		}
		out = append(out, normalized)
	}

	if len(out) == 0 {
		return nil
	}

	return out
}
