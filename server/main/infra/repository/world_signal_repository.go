package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"server/main/domain"
)

type worldSignalRepository struct {
	db *sql.DB
}

func NewWorldSignalRepository(db *sql.DB) *worldSignalRepository {
	return &worldSignalRepository{db: db}
}

func (r *worldSignalRepository) UpsertMany(ctx context.Context, signals []*domain.WorldSignal) (int, error) {
	if len(signals) == 0 {
		return 0, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `
		INSERT INTO world_signals (
			source, location_key, latitude, longitude, observed_at,
			temperature_c, precipitation_mm, wind_speed_ms, weather_code,
			raw_json, create_user, update_user
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12
		)
		ON CONFLICT (source, location_key, observed_at)
		WHERE deleted_at IS NULL
		DO UPDATE SET
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			temperature_c = EXCLUDED.temperature_c,
			precipitation_mm = EXCLUDED.precipitation_mm,
			wind_speed_ms = EXCLUDED.wind_speed_ms,
			weather_code = EXCLUDED.weather_code,
			raw_json = EXCLUDED.raw_json,
			update_user = EXCLUDED.update_user,
			updated_at = CURRENT_TIMESTAMP
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	applied := 0
	for _, signal := range signals {
		if signal == nil {
			continue
		}
		_, err = stmt.ExecContext(
			ctx,
			signal.Source,
			signal.LocationKey,
			signal.Latitude,
			signal.Longitude,
			signal.ObservedAt,
			signal.TemperatureC,
			signal.PrecipitationMM,
			signal.WindSpeedMS,
			signal.WeatherCode,
			sqlNullRawJSON(signal.RawJSON),
			signal.CreateUser,
			signal.UpdateUser,
		)
		if err != nil {
			return 0, err
		}
		applied++
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return applied, nil
}

func (r *worldSignalRepository) FindByRange(
	ctx context.Context,
	source, locationKey string,
	from, to time.Time,
	limit int,
) ([]*domain.WorldSignal, error) {
	if limit < 1 || limit > 1000 {
		limit = 200
	}

	var args []interface{}
	args = append(args, from, to)

	conditions := []string{
		"deleted_at IS NULL",
		"observed_at >= $1",
		"observed_at <= $2",
	}

	if strings.TrimSpace(source) != "" {
		args = append(args, source)
		conditions = append(conditions, fmt.Sprintf("source = $%d", len(args)))
	}
	if strings.TrimSpace(locationKey) != "" {
		args = append(args, locationKey)
		conditions = append(conditions, fmt.Sprintf("location_key = $%d", len(args)))
	}

	args = append(args, limit)
	limitArgPos := len(args)

	query := fmt.Sprintf(`
		SELECT
			id,
			source,
			location_key,
			latitude,
			longitude,
			observed_at,
			temperature_c,
			precipitation_mm,
			wind_speed_ms,
			weather_code,
			raw_json,
			created_at,
			updated_at,
			create_user,
			update_user,
			deleted_at
		FROM world_signals
		WHERE %s
		ORDER BY observed_at ASC, id ASC
		LIMIT $%d
	`, strings.Join(conditions, " AND "), limitArgPos)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	signals := make([]*domain.WorldSignal, 0)
	for rows.Next() {
		var signal domain.WorldSignal
		var raw json.RawMessage
		if err := rows.Scan(
			&signal.ID,
			&signal.Source,
			&signal.LocationKey,
			&signal.Latitude,
			&signal.Longitude,
			&signal.ObservedAt,
			&signal.TemperatureC,
			&signal.PrecipitationMM,
			&signal.WindSpeedMS,
			&signal.WeatherCode,
			&raw,
			&signal.CreatedAt,
			&signal.UpdatedAt,
			&signal.CreateUser,
			&signal.UpdateUser,
			&signal.DeletedAt,
		); err != nil {
			return nil, err
		}
		if len(raw) > 0 {
			signal.RawJSON = raw
		}
		signals = append(signals, &signal)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return signals, nil
}

func sqlNullRawJSON(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}
	return []byte(raw)
}
