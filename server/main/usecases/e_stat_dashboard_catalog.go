package usecases

type EStatDashboardSignalDefinition struct {
	SignalType    string
	Label         string
	Unit          string
	IndicatorCode string
	Derived       bool
}

const worldSignalSourceEStatDashboard = "e_stat_dashboard"

var eStatDashboardSignalCatalog = []EStatDashboardSignalDefinition{
	{
		SignalType:    "population_total",
		Label:         "総人口",
		Unit:          "人",
		IndicatorCode: "0201010000000010000",
	},
	{
		SignalType:    "youth_ratio",
		Label:         "年少人口比",
		Unit:          "%",
		IndicatorCode: "0201010010000020010",
	},
	{
		SignalType:    "senior_ratio",
		Label:         "高齢人口比",
		Unit:          "%",
		IndicatorCode: "0201010010000020030",
	},
	{
		SignalType: "productive_age_ratio",
		Label:      "生産年齢人口比",
		Unit:       "%",
		Derived:    true,
	},
}

func EStatDashboardSignalCatalog() []EStatDashboardSignalDefinition {
	result := make([]EStatDashboardSignalDefinition, len(eStatDashboardSignalCatalog))
	copy(result, eStatDashboardSignalCatalog)
	return result
}

func eStatDashboardIndicatorCodes() []string {
	codes := make([]string, 0, len(eStatDashboardSignalCatalog))
	for _, definition := range eStatDashboardSignalCatalog {
		if definition.Derived || definition.IndicatorCode == "" {
			continue
		}
		codes = append(codes, definition.IndicatorCode)
	}
	return codes
}

func findEStatDashboardSignalByIndicator(indicatorCode string) (EStatDashboardSignalDefinition, bool) {
	for _, definition := range eStatDashboardSignalCatalog {
		if definition.IndicatorCode == indicatorCode {
			return definition, true
		}
	}
	return EStatDashboardSignalDefinition{}, false
}

func findEStatDashboardSignalByType(signalType string) (EStatDashboardSignalDefinition, bool) {
	for _, definition := range eStatDashboardSignalCatalog {
		if definition.SignalType == signalType {
			return definition, true
		}
	}
	return EStatDashboardSignalDefinition{}, false
}
