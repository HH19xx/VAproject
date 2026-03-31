package app

import (
	"context"
	"errors"
)

type DistributionService struct {
	Analyzer DistributionAnalyzerPort
}

var (
	ErrDistributionDatasetInvalid = errors.New("dataset must be action_logs or world_signals")
	ErrDistributionAxisRequired   = errors.New("axis is required")
	ErrDistributionAxisInvalid    = errors.New("axis is invalid for the selected dataset")
	ErrDistributionLocationNeeded = errors.New("location_key is required for world_signals")
	ErrDistributionWindowInvalid  = errors.New("from/to range is required and must be valid")
)

func (u *DistributionService) Analyze(
	ctx context.Context,
	userID int,
	input DistributionInput,
) (*DistributionResult, error) {
	return u.Analyzer.Analyze(ctx, userID, input)
}

func IsDistributionValidationError(err error) bool {
	return errors.Is(err, ErrDistributionDatasetInvalid) ||
		errors.Is(err, ErrDistributionAxisRequired) ||
		errors.Is(err, ErrDistributionAxisInvalid) ||
		errors.Is(err, ErrDistributionLocationNeeded) ||
		errors.Is(err, ErrDistributionWindowInvalid)
}
