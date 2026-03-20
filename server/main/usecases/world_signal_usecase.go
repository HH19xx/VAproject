package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"server/main/domain"
)

type WorldSignalRepository interface {
	UpsertMany(ctx context.Context, signals []*domain.WorldSignal) (int, error)
	FindByRange(ctx context.Context, source, locationKey string, from, to time.Time, limit int) ([]*domain.WorldSignal, error)
}

type WorldSignalFetchResult struct {
	Source       string    `json:"source"`
	LocationKey  string    `json:"location_key"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	From         time.Time `json:"from"`
	To           time.Time `json:"to"`
	AppliedCount int       `json:"applied_count"`
}

type WorldSignalUsecase struct {
	Repo       WorldSignalRepository
	HTTPClient *http.Client
}

var (
	ErrWorldSignalLocationRequired = errors.New("location_key is required")
	ErrWorldSignalLatitudeInvalid  = errors.New("latitude must be between -90 and 90")
	ErrWorldSignalLongitudeInvalid = errors.New("longitude must be between -180 and 180")
	ErrWorldSignalDaysInvalid      = errors.New("past_days and forecast_days must be within 0..14")
	ErrWorldSignalWindowInvalid    = errors.New("from/to range is invalid")
)

func (u *WorldSignalUsecase) FetchAndStoreOpenMeteo(
	ctx context.Context,
	locationKey string,
	latitude, longitude float64,
	pastDays, forecastDays int,
	actor string,
) (*WorldSignalFetchResult, error) {
	locationKey = strings.TrimSpace(locationKey)
	if locationKey == "" {
		return nil, ErrWorldSignalLocationRequired
	}
	if latitude < -90 || latitude > 90 {
		return nil, ErrWorldSignalLatitudeInvalid
	}
	if longitude < -180 || longitude > 180 {
		return nil, ErrWorldSignalLongitudeInvalid
	}
	if pastDays < 0 || pastDays > 14 || forecastDays < 0 || forecastDays > 14 {
		return nil, ErrWorldSignalDaysInvalid
	}
	if actor == "" {
		actor = "system"
	}

	client := u.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	reqURL := buildOpenMeteoURL(latitude, longitude, pastDays, forecastDays)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open-meteo returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var payload openMeteoResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	signals, from, to, err := convertOpenMeteoPayload(payload, locationKey, actor)
	if err != nil {
		return nil, err
	}
	appliedCount, err := u.Repo.UpsertMany(ctx, signals)
	if err != nil {
		return nil, err
	}

	return &WorldSignalFetchResult{
		Source:       "open_meteo",
		LocationKey:  locationKey,
		Latitude:     latitude,
		Longitude:    longitude,
		From:         from,
		To:           to,
		AppliedCount: appliedCount,
	}, nil
}

func (u *WorldSignalUsecase) ListByRange(
	ctx context.Context,
	source, locationKey string,
	from, to time.Time,
	limit int,
) ([]*domain.WorldSignal, error) {
	if from.IsZero() || to.IsZero() || from.After(to) {
		return nil, ErrWorldSignalWindowInvalid
	}
	if strings.TrimSpace(source) == "" {
		source = "open_meteo"
	}
	if limit < 1 || limit > 1000 {
		limit = 200
	}
	return u.Repo.FindByRange(ctx, source, locationKey, from, to, limit)
}

type openMeteoResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Hourly    struct {
		Time          []string  `json:"time"`
		Temperature2M []float64 `json:"temperature_2m"`
		Precipitation []float64 `json:"precipitation"`
		WeatherCode   []int     `json:"weathercode"`
		WindSpeed10M  []float64 `json:"windspeed_10m"`
	} `json:"hourly"`
}

func buildOpenMeteoURL(latitude, longitude float64, pastDays, forecastDays int) string {
	params := url.Values{}
	params.Set("latitude", strconv.FormatFloat(latitude, 'f', 6, 64))
	params.Set("longitude", strconv.FormatFloat(longitude, 'f', 6, 64))
	params.Set("hourly", "temperature_2m,precipitation,weathercode,windspeed_10m")
	params.Set("timezone", "Asia/Tokyo")
	params.Set("past_days", strconv.Itoa(pastDays))
	params.Set("forecast_days", strconv.Itoa(maxInt(forecastDays, 1)))
	return "https://api.open-meteo.com/v1/forecast?" + params.Encode()
}

func convertOpenMeteoPayload(
	payload openMeteoResponse,
	locationKey, actor string,
) ([]*domain.WorldSignal, time.Time, time.Time, error) {
	timeLen := len(payload.Hourly.Time)
	if timeLen == 0 {
		return nil, time.Time{}, time.Time{}, nil
	}

	size := minInt(
		timeLen,
		len(payload.Hourly.Temperature2M),
		len(payload.Hourly.Precipitation),
		len(payload.Hourly.WeatherCode),
		len(payload.Hourly.WindSpeed10M),
	)
	if size == 0 {
		return nil, time.Time{}, time.Time{}, nil
	}

	signals := make([]*domain.WorldSignal, 0, size)
	var from time.Time
	var to time.Time

	for i := 0; i < size; i++ {
		observedAt, err := parseOpenMeteoTime(payload.Hourly.Time[i])
		if err != nil {
			continue
		}
		if from.IsZero() || observedAt.Before(from) {
			from = observedAt
		}
		if to.IsZero() || observedAt.After(to) {
			to = observedAt
		}

		temp := roundFloat(payload.Hourly.Temperature2M[i], 3)
		precip := roundFloat(payload.Hourly.Precipitation[i], 3)
		wind := roundFloat(payload.Hourly.WindSpeed10M[i], 3)
		code := payload.Hourly.WeatherCode[i]

		raw, _ := json.Marshal(map[string]interface{}{
			"time":             payload.Hourly.Time[i],
			"temperature_2m":   temp,
			"precipitation":    precip,
			"weathercode":      code,
			"windspeed_10m":    wind,
			"origin_latitude":  payload.Latitude,
			"origin_longitude": payload.Longitude,
		})

		signals = append(signals, &domain.WorldSignal{
			Source:          "open_meteo",
			LocationKey:     locationKey,
			Latitude:        payload.Latitude,
			Longitude:       payload.Longitude,
			ObservedAt:      observedAt,
			TemperatureC:    float64Ptr(temp),
			PrecipitationMM: float64Ptr(precip),
			WindSpeedMS:     float64Ptr(wind),
			WeatherCode:     intPtr(code),
			RawJSON:         raw,
			CreateUser:      actor,
			UpdateUser:      actor,
		})
	}

	return signals, from, to, nil
}

func parseOpenMeteoTime(raw string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02T15:04", raw); err == nil {
		jst, locErr := time.LoadLocation("Asia/Tokyo")
		if locErr != nil {
			return t.UTC(), nil
		}
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, jst).UTC(), nil
	}
	return time.Parse(time.RFC3339, raw)
}

func roundFloat(value float64, places int) float64 {
	pow := math.Pow10(places)
	return math.Round(value*pow) / pow
}

func float64Ptr(v float64) *float64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func minInt(values ...int) int {
	if len(values) == 0 {
		return 0
	}
	minValue := values[0]
	for _, v := range values[1:] {
		if v < minValue {
			minValue = v
		}
	}
	return minValue
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
