package usecases

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"server/main/domain"
)

type ActionLogRepository interface {
	FindAll(ctx context.Context, userID int, tagIDs []int, limit, offset int, options ActionLogListOptions) ([]*domain.ActionLog, error)
	Count(ctx context.Context, userID int, tagIDs []int, options ActionLogListOptions) (int, error)
	FindByID(ctx context.Context, id, userID int) (*domain.ActionLog, error)
	ListAttributeKeys(ctx context.Context, userID int) ([]string, error)
	Create(ctx context.Context, log *domain.ActionLog) (int, error)
	Update(ctx context.Context, log *domain.ActionLog) error
	Delete(ctx context.Context, id, userID int) error
}

type ActionLogTagRepository interface {
	FindByID(ctx context.Context, id, userID int) (*domain.Tag, error)
}

type ActionLogPrototypeRepository interface {
	FindByID(ctx context.Context, id, userID int) (*domain.Prototype, error)
	Create(ctx context.Context, prototype *domain.Prototype) (int, error)
	Update(ctx context.Context, prototype *domain.Prototype) error
}

type ActionLogUsecase struct {
	Repo          ActionLogRepository
	TagRepo       ActionLogTagRepository
	PrototypeRepo ActionLogPrototypeRepository
}

type ActionLogListOptions struct {
	From          *time.Time
	To            *time.Time
	Sort          string
	Order         string
	AnyTagIDs     []int
	AnyTagGroups  [][]int
	ExcludeTagIDs []int
}

type ActionLogListMeta struct {
	From              *time.Time `json:"from,omitempty"`
	To                *time.Time `json:"to,omitempty"`
	Sort              string     `json:"sort"`
	Order             string     `json:"order"`
	UsedTagIDs        []int      `json:"used_tag_ids"`
	UsedAnyTagIDs     []int      `json:"used_any_tag_ids"`
	UsedAnyTagGroups  [][]int    `json:"used_any_tag_groups"`
	UsedExcludeTagIDs []int      `json:"used_exclude_tag_ids"`
	AxisCandidates    []string   `json:"axis_candidates"`
}

var (
	ErrActionLogTitleRequired            = errors.New("title is required")
	ErrActionLogTitleTooLong             = errors.New("title must be <= 128 chars")
	ErrActionLogOccurredAtRequired       = errors.New("occurred_at is required")
	ErrActionLogTagIDsRequired           = errors.New("tag_ids requires at least one value")
	ErrActionLogTagIDInvalid             = errors.New("tag_ids must contain only positive ids")
	ErrActionLogTagNotFound              = errors.New("specified tag_id was not found")
	ErrActionLogParentPrototypeNotFound  = errors.New("specified parent_prototype_id was not found")
	ErrActionLogParentPrototypeIDInvalid = errors.New("parent_prototype_id must be positive")
	ErrActionLogParentPrototypeSelf      = errors.New("self parent prototype is not allowed")
	ErrActionLogSortInvalid              = errors.New("sort is invalid")
	ErrActionLogOrderInvalid             = errors.New("order must be asc or desc")
	ErrActionLogAttributeKeyRequired     = errors.New("attribute key is required")
	ErrActionLogAttributeKeyTooLong      = errors.New("attribute key must be <= 64 chars")
	ErrActionLogAttributeValueRequired   = errors.New("attribute value_number is required")
	ErrActionLogAttributeValueInvalid    = errors.New("attribute value_number must be finite")
)

func (u *ActionLogUsecase) GetActionLogs(
	ctx context.Context,
	actorUserID int,
	tagIDs []int,
	page, limit int,
	options ActionLogListOptions,
) ([]*domain.ActionLog, int, ActionLogListMeta, error) {
	meta := ActionLogListMeta{
		AxisCandidates: []string{"occurred_at", "created_at", "updated_at", "tag_count", "prototype_id"},
	}

	normalizedOptions, err := normalizeListOptions(options)
	if err != nil {
		return nil, 0, meta, err
	}
	meta.From = normalizedOptions.From
	meta.To = normalizedOptions.To
	meta.Sort = normalizedOptions.Sort
	meta.Order = normalizedOptions.Order

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	normalizedTagIDs := normalizeTagIDs(tagIDs)
	meta.UsedTagIDs = normalizedTagIDs
	meta.UsedAnyTagIDs = normalizedOptions.AnyTagIDs
	meta.UsedAnyTagGroups = normalizedOptions.AnyTagGroups
	meta.UsedExcludeTagIDs = normalizedOptions.ExcludeTagIDs

	logs, err := u.Repo.FindAll(ctx, actorUserID, normalizedTagIDs, limit, offset, normalizedOptions)
	if err != nil {
		return nil, 0, meta, err
	}
	total, err := u.Repo.Count(ctx, actorUserID, normalizedTagIDs, normalizedOptions)
	if err != nil {
		return nil, 0, meta, err
	}
	attributeKeys, err := u.Repo.ListAttributeKeys(ctx, actorUserID)
	if err != nil {
		return nil, 0, meta, err
	}
	meta.AxisCandidates = append(meta.AxisCandidates, attributeKeys...)
	return logs, total, meta, nil
}

func (u *ActionLogUsecase) GetActionLogByID(ctx context.Context, actorUserID, id int) (*domain.ActionLog, error) {
	return u.Repo.FindByID(ctx, id, actorUserID)
}

func (u *ActionLogUsecase) CreateActionLog(
	ctx context.Context,
	actorUserID int,
	title string,
	parentPrototypeID *int,
	occurredAt time.Time,
	notes,
	user string,
	tagIDs []int,
	attributes []domain.ActionLogAttribute,
) (*domain.ActionLog, error) {
	tagIDs, attributes, err := u.validateInput(ctx, actorUserID, title, parentPrototypeID, occurredAt, tagIDs, attributes)
	if err != nil {
		return nil, err
	}

	prototypeID, err := u.materializePrototype(ctx, actorUserID, nil, title, notes, user, parentPrototypeID, tagIDs)
	if err != nil {
		return nil, err
	}

	log := &domain.ActionLog{
		UserID:      actorUserID,
		Title:       title,
		PrototypeID: prototypeID,
		OccurredAt:  occurredAt,
		Notes:       notes,
		TagIDs:      tagIDs,
		Attributes:  attributes,
		CreateUser:  user,
		UpdateUser:  user,
	}
	id, err := u.Repo.Create(ctx, log)
	if err != nil {
		return nil, err
	}
	return u.Repo.FindByID(ctx, id, actorUserID)
}

func (u *ActionLogUsecase) UpdateActionLog(
	ctx context.Context,
	actorUserID, id int,
	title string,
	parentPrototypeID *int,
	occurredAt time.Time,
	notes,
	user string,
	tagIDs []int,
	attributes []domain.ActionLogAttribute,
) (*domain.ActionLog, error) {
	tagIDs, attributes, err := u.validateInput(ctx, actorUserID, title, parentPrototypeID, occurredAt, tagIDs, attributes)
	if err != nil {
		return nil, err
	}

	log, err := u.Repo.FindByID(ctx, id, actorUserID)
	if err != nil {
		return nil, err
	}

	prototypeID, err := u.materializePrototype(ctx, actorUserID, log.PrototypeID, title, notes, user, parentPrototypeID, tagIDs)
	if err != nil {
		return nil, err
	}

	log.Title = title
	log.PrototypeID = prototypeID
	log.OccurredAt = occurredAt
	log.Notes = notes
	log.TagIDs = tagIDs
	log.Attributes = attributes
	log.UpdateUser = user

	if err := u.Repo.Update(ctx, log); err != nil {
		return nil, err
	}
	return u.Repo.FindByID(ctx, id, actorUserID)
}

func (u *ActionLogUsecase) DeleteActionLog(ctx context.Context, actorUserID, id int) error {
	return u.Repo.Delete(ctx, id, actorUserID)
}

func (u *ActionLogUsecase) validateInput(
	ctx context.Context,
	actorUserID int,
	title string,
	parentPrototypeID *int,
	occurredAt time.Time,
	tagIDs []int,
	attributes []domain.ActionLogAttribute,
) ([]int, []domain.ActionLogAttribute, error) {
	if title == "" {
		return nil, nil, ErrActionLogTitleRequired
	}
	if len(title) > 128 {
		return nil, nil, ErrActionLogTitleTooLong
	}
	if occurredAt.IsZero() {
		return nil, nil, ErrActionLogOccurredAtRequired
	}

	tagIDs = normalizeTagIDs(tagIDs)
	if len(tagIDs) == 0 {
		return nil, nil, ErrActionLogTagIDsRequired
	}
	for _, tagID := range tagIDs {
		if tagID <= 0 {
			return nil, nil, ErrActionLogTagIDInvalid
		}
	}
	if err := u.validateTagsOwnership(ctx, actorUserID, tagIDs); err != nil {
		return nil, nil, err
	}

	if parentPrototypeID != nil {
		if *parentPrototypeID <= 0 {
			return nil, nil, ErrActionLogParentPrototypeIDInvalid
		}
		if err := u.validateParentPrototypeOwnership(ctx, actorUserID, *parentPrototypeID); err != nil {
			return nil, nil, err
		}
	}

	normalizedAttributes, err := normalizeAttributes(attributes)
	if err != nil {
		return nil, nil, err
	}

	return tagIDs, normalizedAttributes, nil
}

func (u *ActionLogUsecase) materializePrototype(
	ctx context.Context,
	actorUserID int,
	currentPrototypeID *int,
	title, notes, user string,
	parentPrototypeID *int,
	tagIDs []int,
) (*int, error) {
	if u.PrototypeRepo == nil {
		return currentPrototypeID, nil
	}

	if currentPrototypeID != nil {
		prototype, err := u.PrototypeRepo.FindByID(ctx, *currentPrototypeID, actorUserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				currentPrototypeID = nil
			} else {
				return nil, err
			}
		} else {
			if parentPrototypeID != nil && *parentPrototypeID == prototype.ID {
				return nil, ErrActionLogParentPrototypeSelf
			}
			prototype.Name = title
			prototype.Description = notes
			prototype.ParentPrototypeID = parentPrototypeID
			prototype.TagIDs = tagIDs
			prototype.UpdateUser = user
			if err := u.PrototypeRepo.Update(ctx, prototype); err != nil {
				return nil, err
			}
			return currentPrototypeID, nil
		}
	}

	prototype := &domain.Prototype{
		UserID:            actorUserID,
		Name:              title,
		Description:       notes,
		ParentPrototypeID: parentPrototypeID,
		TagIDs:            tagIDs,
		CreateUser:        user,
		UpdateUser:        user,
	}
	id, err := u.PrototypeRepo.Create(ctx, prototype)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (u *ActionLogUsecase) validateTagsOwnership(ctx context.Context, actorUserID int, tagIDs []int) error {
	if u.TagRepo == nil {
		return nil
	}
	for _, tagID := range tagIDs {
		if _, err := u.TagRepo.FindByID(ctx, tagID, actorUserID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrActionLogTagNotFound
			}
			return err
		}
	}
	return nil
}

func (u *ActionLogUsecase) validateParentPrototypeOwnership(ctx context.Context, actorUserID, parentPrototypeID int) error {
	if u.PrototypeRepo == nil {
		return nil
	}
	if _, err := u.PrototypeRepo.FindByID(ctx, parentPrototypeID, actorUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrActionLogParentPrototypeNotFound
		}
		return err
	}
	return nil
}

func normalizeTagIDs(tagIDs []int) []int {
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

func normalizeListOptions(options ActionLogListOptions) (ActionLogListOptions, error) {
	if options.Sort == "" {
		options.Sort = "occurred_at"
	}
	if options.Order == "" {
		options.Order = "desc"
	}

	switch options.Sort {
	case "occurred_at", "created_at", "updated_at", "title", "tag_count":
	default:
		return options, ErrActionLogSortInvalid
	}

	switch options.Order {
	case "asc", "desc":
	default:
		return options, ErrActionLogOrderInvalid
	}

	if options.From != nil && options.To != nil && options.From.After(*options.To) {
		from := options.To
		to := options.From
		options.From = from
		options.To = to
	}

	options.AnyTagIDs = normalizeTagIDs(options.AnyTagIDs)
	options.AnyTagGroups = normalizeTagIDGroups(options.AnyTagGroups)
	options.ExcludeTagIDs = normalizeTagIDs(options.ExcludeTagIDs)
	for _, tagID := range options.AnyTagIDs {
		if tagID <= 0 {
			return options, ErrActionLogTagIDInvalid
		}
	}
	for _, tagID := range options.ExcludeTagIDs {
		if tagID <= 0 {
			return options, ErrActionLogTagIDInvalid
		}
	}
	for _, group := range options.AnyTagGroups {
		for _, tagID := range group {
			if tagID <= 0 {
				return options, ErrActionLogTagIDInvalid
			}
		}
	}
	return options, nil
}

func normalizeTagIDGroups(groups [][]int) [][]int {
	if len(groups) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([][]int, 0, len(groups))
	for _, group := range groups {
		normalized := normalizeTagIDs(group)
		if len(normalized) == 0 {
			continue
		}
		key := ""
		var b strings.Builder
		for i, id := range normalized {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(strconv.Itoa(id))
		}
		key = b.String()
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, normalized)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

var ErrActionLogNotFound = sql.ErrNoRows

func normalizeAttributes(attributes []domain.ActionLogAttribute) ([]domain.ActionLogAttribute, error) {
	if len(attributes) == 0 {
		return nil, nil
	}

	seen := map[string]struct{}{}
	normalized := make([]domain.ActionLogAttribute, 0, len(attributes))

	for _, attr := range attributes {
		key := strings.ToLower(strings.TrimSpace(attr.Key))
		key = strings.ReplaceAll(key, " ", "_")
		if key == "" {
			return nil, ErrActionLogAttributeKeyRequired
		}
		if len(key) > 64 {
			return nil, ErrActionLogAttributeKeyTooLong
		}
		if attr.ValueNumber == nil {
			return nil, ErrActionLogAttributeValueRequired
		}
		if math.IsNaN(*attr.ValueNumber) || math.IsInf(*attr.ValueNumber, 0) {
			return nil, ErrActionLogAttributeValueInvalid
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		val := *attr.ValueNumber
		normalized = append(normalized, domain.ActionLogAttribute{
			Key:         key,
			ValueNumber: &val,
		})
	}

	if len(normalized) == 0 {
		return nil, nil
	}
	return normalized, nil
}
