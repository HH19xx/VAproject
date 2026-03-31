package app

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

	externalsignaldomain "server/internal/modules/externalsignal/domain"
)

type WorldSignalRepository interface {
	UpsertMany(ctx context.Context, signals []*externalsignaldomain.WorldSignal) (int, error)
	FindByRange(ctx context.Context, source, locationKey, signalType string, from, to time.Time, limit int) ([]*externalsignaldomain.WorldSignal, error)
}

type FetchResult struct {
	Source       string    `json:"source"`
	LocationKey  string    `json:"location_key"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	From         time.Time `json:"from"`
	To           time.Time `json:"to"`
	AppliedCount int       `json:"applied_count"`
}

type QueryService struct {
	Repo       WorldSignalRepository
	HTTPClient *http.Client
}

var (
	ErrWorldSignalLocationRequired  = errors.New("location_key is required")
	ErrWorldSignalSourceUnsupported = errors.New("source is unsupported")
	ErrWorldSignalLatitudeInvalid   = errors.New("latitude must be between -90 and 90")
	ErrWorldSignalLongitudeInvalid  = errors.New("longitude must be between -180 and 180")
	ErrWorldSignalDaysInvalid       = errors.New("past_days and forecast_days must be within 0..14")
	ErrWorldSignalWindowInvalid     = errors.New("from/to range is invalid")
)

const (
	worldSignalSourceOpenMeteo = "open_meteo"
)

func (s *QueryService) FetchAndStore(
	ctx context.Context,
	source string,
	locationKey string,
	latitude, longitude float64,
	pastDays, forecastDays int,
	actor string,
) (*FetchResult, error) {
	switch strings.TrimSpace(source) {
	case "", worldSignalSourceOpenMeteo:
		return s.FetchAndStoreOpenMeteo(ctx, locationKey, latitude, longitude, pastDays, forecastDays, actor)
	case worldSignalSourceEStatDashboard:
		return s.FetchAndStoreEStatDashboard(ctx, locationKey, actor)
	default:
		return nil, ErrWorldSignalSourceUnsupported
	}
}

func (s *QueryService) FetchAndStoreOpenMeteo(
	ctx context.Context,
	locationKey string,
	latitude, longitude float64,
	pastDays, forecastDays int,
	actor string,
) (*FetchResult, error) {
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

	client := s.HTTPClient
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
	appliedCount, err := s.Repo.UpsertMany(ctx, signals)
	if err != nil {
		return nil, err
	}

	return &FetchResult{
		Source:       worldSignalSourceOpenMeteo,
		LocationKey:  locationKey,
		Latitude:     latitude,
		Longitude:    longitude,
		From:         from,
		To:           to,
		AppliedCount: appliedCount,
	}, nil
}

func (s *QueryService) ListByRange(
	ctx context.Context,
	source, locationKey, signalType string,
	from, to time.Time,
	limit int,
) ([]*externalsignaldomain.WorldSignal, error) {
	if from.IsZero() || to.IsZero() || from.After(to) {
		return nil, ErrWorldSignalWindowInvalid
	}
	if strings.TrimSpace(source) == "" {
		source = worldSignalSourceOpenMeteo
	}
	if limit < 1 || limit > 1000 {
		limit = 200
	}
	return s.Repo.FindByRange(ctx, source, locationKey, signalType, from, to, limit)
}

func (s *QueryService) FetchAndStoreEStatDashboard(
	ctx context.Context,
	locationKey string,
	actor string,
) (*FetchResult, error) {
	locationKey = strings.TrimSpace(locationKey)
	if locationKey == "" {
		return nil, ErrWorldSignalLocationRequired
	}
	if actor == "" {
		actor = "system"
	}

	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	reqURL := buildEStatDashboardURL(locationKey)
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
		return nil, fmt.Errorf("e-stat dashboard returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	signals, from, to, err := convertEStatDashboardPayload(body, locationKey, actor)
	if err != nil {
		return nil, err
	}
	appliedCount, err := s.Repo.UpsertMany(ctx, signals)
	if err != nil {
		return nil, err
	}

	return &FetchResult{
		Source:       worldSignalSourceEStatDashboard,
		LocationKey:  locationKey,
		Latitude:     0,
		Longitude:    0,
		From:         from,
		To:           to,
		AppliedCount: appliedCount,
	}, nil
}

func buildEStatDashboardURL(locationKey string) string {
	params := url.Values{}
	params.Set("Lang", "JP")
	params.Set("IndicatorCode", strings.Join(eStatDashboardIndicatorCodes(), ","))
	params.Set("RegionCode", locationKey)
	params.Set("RegionalRank", "3")
	params.Set("Cycle", "3")
	params.Set("IsSeasonalAdjustment", "1")
	params.Set("MetaGetFlg", "Y")
	params.Set("SectionHeaderFlg", "1")
	return "https://dashboard.e-stat.go.jp/api/1.0/Json/getData?" + params.Encode()
}

type eStatDashboardResponse struct {
	GetStats struct {
		StatisticalData struct {
			ClassInf struct {
				ClassObj json.RawMessage `json:"CLASS_OBJ"`
			} `json:"CLASS_INF"`
			DataInf struct {
				DataObj json.RawMessage `json:"DATA_OBJ"`
			} `json:"DATA_INF"`
		} `json:"STATISTICAL_DATA"`
	} `json:"GET_STATS"`
}

type eStatDashboardDataObj struct {
	Value struct {
		Indicator  string `json:"@indicator"`
		RegionCode string `json:"@regionCode"`
		Time       string `json:"@time"`
		Value      string `json:"$"`
	} `json:"VALUE"`
}

type eStatDashboardClassObj struct {
	ID    string                    `json:"@id"`
	Class []eStatDashboardClassItem `json:"CLASS"`
}

type eStatDashboardClassItem struct {
	Code string `json:"@code"`
	Name string `json:"@name"`
}

func convertEStatDashboardPayload(body []byte, locationKey, actor string) ([]*externalsignaldomain.WorldSignal, time.Time, time.Time, error) {
	var payload eStatDashboardResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, time.Time{}, time.Time{}, err
	}

	dataObjs, err := decodeEStatDashboardDataObjects(payload.GetStats.StatisticalData.DataInf.DataObj)
	if err != nil {
		return nil, time.Time{}, time.Time{}, err
	}
	classObjs, err := decodeEStatDashboardClassObjects(payload.GetStats.StatisticalData.ClassInf.ClassObj)
	if err != nil {
		return nil, time.Time{}, time.Time{}, err
	}
	indicatorNames := buildEStatIndicatorNameMap(classObjs)

	signals := make([]*externalsignaldomain.WorldSignal, 0, len(dataObjs)*2)
	var from time.Time
	var to time.Time
	ageByYear := map[time.Time]map[string]float64{}

	for _, obj := range dataObjs {
		if strings.TrimSpace(obj.Value.RegionCode) != locationKey {
			continue
		}
		observedAt, err := parseEStatDashboardTime(obj.Value.Time)
		if err != nil {
			continue
		}
		value, err := strconv.ParseFloat(strings.ReplaceAll(obj.Value.Value, ",", ""), 64)
		if err != nil {
			continue
		}
		if from.IsZero() || observedAt.Before(from) {
			from = observedAt
		}
		if to.IsZero() || observedAt.After(to) {
			to = observedAt
		}

		definition, ok := findEStatDashboardSignalByIndicator(obj.Value.Indicator)
		if !ok {
			continue
		}
		if _, ok := ageByYear[observedAt]; !ok {
			ageByYear[observedAt] = map[string]float64{}
		}
		ageByYear[observedAt][definition.SignalType] = value

		raw, _ := json.Marshal(map[string]interface{}{
			"time":           obj.Value.Time,
			"indicator_code": obj.Value.Indicator,
			"region_code":    obj.Value.RegionCode,
			"value":          value,
			"provider":       worldSignalSourceEStatDashboard,
		})

		name := indicatorNames[obj.Value.Indicator]
		if name == "" {
			name = definition.Label
		}
		signals = append(signals, &externalsignaldomain.WorldSignal{
			Source:      worldSignalSourceEStatDashboard,
			LocationKey: locationKey,
			SignalType:  definition.SignalType,
			SignalLabel: stringPtr(name),
			SignalUnit:  stringPtr(definition.Unit),
			SignalValue: float64Ptr(roundFloat(value, 3)),
			ObservedAt:  observedAt,
			RawJSON:     raw,
			CreateUser:  actor,
			UpdateUser:  actor,
		})
	}

	for observedAt, values := range ageByYear {
		youth, okYouth := values["youth_ratio"]
		senior, okSenior := values["senior_ratio"]
		if !okYouth || !okSenior {
			continue
		}
		productive := roundFloat(100-youth-senior, 3)
		raw, _ := json.Marshal(map[string]interface{}{
			"time":     observedAt.Format("2006CY00"),
			"provider": worldSignalSourceEStatDashboard,
			"derived":  true,
			"value":    productive,
		})
		productiveDefinition, _ := findEStatDashboardSignalByType("productive_age_ratio")
		signals = append(signals, &externalsignaldomain.WorldSignal{
			Source:      worldSignalSourceEStatDashboard,
			LocationKey: locationKey,
			SignalType:  productiveDefinition.SignalType,
			SignalLabel: stringPtr(productiveDefinition.Label),
			SignalUnit:  stringPtr(productiveDefinition.Unit),
			SignalValue: float64Ptr(productive),
			ObservedAt:  observedAt,
			RawJSON:     raw,
			CreateUser:  actor,
			UpdateUser:  actor,
		})
	}

	return signals, from, to, nil
}

func decodeEStatDashboardDataObjects(raw json.RawMessage) ([]eStatDashboardDataObj, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var list []eStatDashboardDataObj
	if err := json.Unmarshal(raw, &list); err == nil {
		return list, nil
	}
	var one eStatDashboardDataObj
	if err := json.Unmarshal(raw, &one); err != nil {
		return nil, err
	}
	return []eStatDashboardDataObj{one}, nil
}

func decodeEStatDashboardClassObjects(raw json.RawMessage) ([]eStatDashboardClassObj, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var list []eStatDashboardClassObj
	if err := json.Unmarshal(raw, &list); err == nil {
		return list, nil
	}
	var one eStatDashboardClassObj
	if err := json.Unmarshal(raw, &one); err != nil {
		return nil, err
	}
	return []eStatDashboardClassObj{one}, nil
}

func buildEStatIndicatorNameMap(classObjs []eStatDashboardClassObj) map[string]string {
	result := map[string]string{}
	for _, classObj := range classObjs {
		for _, item := range classObj.Class {
			if item.Code != "" && item.Name != "" {
				result[item.Code] = item.Name
			}
		}
	}
	return result
}

func parseEStatDashboardTime(raw string) (time.Time, error) {
	if strings.HasSuffix(raw, "CY00") {
		year, err := strconv.Atoi(strings.TrimSuffix(raw, "CY00"))
		if err != nil {
			return time.Time{}, err
		}
		return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), nil
	}
	return time.Parse(time.RFC3339, raw)
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

func convertOpenMeteoPayload(payload openMeteoResponse, locationKey, actor string) ([]*externalsignaldomain.WorldSignal, time.Time, time.Time, error) {
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

	signals := make([]*externalsignaldomain.WorldSignal, 0, size)
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

		signals = append(signals, &externalsignaldomain.WorldSignal{
			Source:          worldSignalSourceOpenMeteo,
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

func float64Ptr(v float64) *float64 { return &v }
func intPtr(v int) *int             { return &v }
func stringPtr(v string) *string    { return &v }

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
