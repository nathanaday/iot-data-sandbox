package timeseries

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

const (
	// Common timestamp column names
	TimestampCol = "timestamp"
	ValueCol     = "value"
)

type TimeSeriesData struct {
	DataFrame  dataframe.DataFrame
	StartTime  time.Time
	EndTime    time.Time
	RowCount   int
	TimeLabel  string
	ValueLabel string
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func LoadAndValidateCSV(filePath string) (*TimeSeriesData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	df := dataframe.ReadCSV(file)

	if df.Err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", df.Err)
	}

	if err := validateStructure(df); err != nil {
		return nil, err
	}

	normalizedDF, timeLabel, valueLabel, err := normalizeTimestamps(df)
	if err != nil {
		return nil, err
	}

	if err := validateValues(normalizedDF); err != nil {
		return nil, err
	}

	tsData := &TimeSeriesData{
		DataFrame:  normalizedDF,
		RowCount:   normalizedDF.Nrow(),
		TimeLabel:  timeLabel,
		ValueLabel: valueLabel,
	}

	if normalizedDF.Nrow() > 0 {
		startTime, endTime, err := getTimeRange(normalizedDF)
		if err != nil {
			return nil, err
		}
		tsData.StartTime = startTime
		tsData.EndTime = endTime
	}

	return tsData, nil
}

func validateStructure(df dataframe.DataFrame) error {
	cols := df.Names()

	if len(cols) < 2 {
		return &ValidationError{Message: "CSV must contain at least 2 columns (timestamp and value)"}
	}

	hasTimestamp := false

	for _, col := range cols {
		colLower := strings.ToLower(col)
		if colLower == TimestampCol || colLower == "time" || colLower == "date" || colLower == "datetime" {
			hasTimestamp = true
			break
		}
	}

	if !hasTimestamp {
		return &ValidationError{Message: "CSV must contain a timestamp column (timestamp, time, date, or datetime)"}
	}

	return nil
}

func normalizeTimestamps(df dataframe.DataFrame) (dataframe.DataFrame, string, string, error) {
	cols := df.Names()

	// Default labels if no headers are found
	timeLabel := "time"
	valueLabel := "value"

	if len(cols) == 0 {
		return df, timeLabel, valueLabel, &ValidationError{Message: "no columns found in CSV"}
	}

	// Find the timestamp column
	timestampColName := ""
	for _, col := range cols {
		colLower := strings.ToLower(col)
		if colLower == TimestampCol || colLower == "time" || colLower == "date" || colLower == "datetime" {
			timestampColName = col
			timeLabel = col
			break
		}
	}

	if timestampColName == "" {
		return df, timeLabel, valueLabel, &ValidationError{Message: "no timestamp column found"}
	}

	// Find the value column (first non-timestamp column)
	valueColName := ""
	for _, col := range cols {
		if col != timestampColName {
			valueColName = col
			valueLabel = col
			break
		}
	}

	if valueColName == "" {
		return df, timeLabel, valueLabel, &ValidationError{Message: "no value column found"}
	}

	// Normalize timestamps
	timestampSeries := df.Col(timestampColName)
	records := timestampSeries.Records()

	normalizedTimestamps := make([]string, len(records))

	for i, record := range records {
		if i == 0 {
			normalizedTimestamps[i] = "timestamp"
			continue
		}

		parsedTime, err := parseTimestamp(record)
		if err != nil {
			return df, timeLabel, valueLabel, fmt.Errorf("invalid timestamp at row %d: %w", i, err)
		}
		normalizedTimestamps[i] = parsedTime.Format(time.RFC3339)
	}

	newTimestampSeries := series.New(normalizedTimestamps, series.String, "timestamp")

	// Extract value series and rename to "value"
	origValueSeries := df.Col(valueColName)
	valueRecords := origValueSeries.Records()
	newValueSeries := series.New(valueRecords, origValueSeries.Type(), "value")

	newDF := dataframe.New(
		newTimestampSeries,
		newValueSeries,
	)

	return newDF, timeLabel, valueLabel, nil
}

func parseTimestamp(ts string) (time.Time, error) {
	ts = strings.TrimSpace(ts)

	// Try Unix timestamp (seconds)
	if val, err := strconv.ParseInt(ts, 10, 64); err == nil {
		// Unix timestamp in seconds
		if val > 1e12 {
			// Milliseconds
			return time.Unix(0, val*int64(time.Millisecond)), nil
		}
		return time.Unix(val, 0), nil
	}

	// Try Unix timestamp (float seconds)
	if val, err := strconv.ParseFloat(ts, 64); err == nil {
		sec := int64(val)
		nsec := int64((val - float64(sec)) * 1e9)
		return time.Unix(sec, nsec), nil
	}

	// Try ISO 8601 / RFC3339
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02",
		"01/02/2006 15:04:05",
		"01/02/2006",
		"1/2/2006 15:04:05",
		"1/2/2006",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, ts); err == nil {
			return t, nil
		}
	}

	// Try Julian Day
	if val, err := strconv.ParseFloat(ts, 64); err == nil && val > 2400000 && val < 2500000 {
		jd := val
		jd = jd + 0.5
		z := int(jd)
		f := jd - float64(z)

		var a int
		if z < 2299161 {
			a = z
		} else {
			alpha := int((float64(z) - 1867216.25) / 36524.25)
			a = z + 1 + alpha - alpha/4
		}

		b := a + 1524
		c := int((float64(b) - 122.1) / 365.25)
		d := int(365.25 * float64(c))
		e := int(float64(b-d) / 30.6001)

		day := b - d - int(30.6001*float64(e))
		month := e - 1
		if e > 13 {
			month = e - 13
		}
		year := c - 4716
		if month < 3 {
			year = c - 4715
		}

		seconds := f * 86400.0
		hour := int(seconds / 3600)
		minute := int((seconds - float64(hour*3600)) / 60)
		second := int(seconds - float64(hour*3600) - float64(minute*60))

		return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC), nil
	}

	return time.Time{}, &ValidationError{Message: fmt.Sprintf("unsupported timestamp format: %s", ts)}
}

func validateValues(df dataframe.DataFrame) error {
	valueSeries := df.Col(ValueCol)
	if valueSeries.Err != nil {
		return &ValidationError{Message: "value column not found after normalization"}
	}

	records := valueSeries.Records()
	for i := 1; i < len(records); i++ {
		if _, err := strconv.ParseFloat(records[i], 64); err != nil {
			return &ValidationError{Message: fmt.Sprintf("invalid value at row %d: must be a number", i)}
		}
	}

	return nil
}

func getTimeRange(df dataframe.DataFrame) (time.Time, time.Time, error) {
	timestampSeries := df.Col("timestamp")
	records := timestampSeries.Records()

	if len(records) <= 1 {
		return time.Time{}, time.Time{}, &ValidationError{Message: "insufficient data"}
	}

	var minTime, maxTime time.Time

	for i := 1; i < len(records); i++ {
		t, err := time.Parse(time.RFC3339, records[i])
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("failed to parse timestamp at row %d: %w", i, err)
		}

		if i == 1 || t.Before(minTime) {
			minTime = t
		}
		if i == 1 || t.After(maxTime) {
			maxTime = t
		}
	}

	return minTime, maxTime, nil
}

func FilterByTimeRange(tsData *TimeSeriesData, startTime, endTime *time.Time) (*TimeSeriesData, error) {
	df := tsData.DataFrame

	if startTime == nil && endTime == nil {
		return tsData, nil
	}

	timestampSeries := df.Col("timestamp")
	records := timestampSeries.Records()

	mask := make([]bool, len(records))
	mask[0] = true

	for i := 1; i < len(records); i++ {
		t, err := time.Parse(time.RFC3339, records[i])
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}

		include := true
		if startTime != nil && t.Before(*startTime) {
			include = false
		}
		if endTime != nil && t.After(*endTime) {
			include = false
		}

		mask[i] = include
	}

	filteredDF := df.Filter(
		dataframe.F{
			Colname:    "timestamp",
			Comparator: series.In,
			Comparando: getFilteredTimestamps(records, mask),
		},
	)

	result := &TimeSeriesData{
		DataFrame: filteredDF,
		RowCount:  filteredDF.Nrow(),
	}

	if filteredDF.Nrow() > 0 {
		start, end, err := getTimeRange(filteredDF)
		if err == nil {
			result.StartTime = start
			result.EndTime = end
		}
	}

	return result, nil
}

func getFilteredTimestamps(records []string, mask []bool) []string {
	var filtered []string
	for i, include := range mask {
		if include && i > 0 {
			filtered = append(filtered, records[i])
		}
	}
	return filtered
}
