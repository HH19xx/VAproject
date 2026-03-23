package domain

import (
	"encoding/json"
	"time"
)

type WorldSignal struct {
	ID              int             `json:"id"`
	Source          string          `json:"source"`
	LocationKey     string          `json:"location_key"`
	SignalType      string          `json:"signal_type"`
	SignalLabel     *string         `json:"signal_label,omitempty"`
	SignalUnit      *string         `json:"signal_unit,omitempty"`
	SignalValue     *float64        `json:"signal_value,omitempty"`
	Latitude        float64         `json:"latitude"`
	Longitude       float64         `json:"longitude"`
	ObservedAt      time.Time       `json:"observed_at"`
	TemperatureC    *float64        `json:"temperature_c,omitempty"`
	PrecipitationMM *float64        `json:"precipitation_mm,omitempty"`
	WindSpeedMS     *float64        `json:"wind_speed_ms,omitempty"`
	WeatherCode     *int            `json:"weather_code,omitempty"`
	RawJSON         json.RawMessage `json:"raw_json,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	CreateUser      string          `json:"create_user"`
	UpdateUser      string          `json:"update_user"`
	DeletedAt       *time.Time      `json:"deleted_at,omitempty"`
}
