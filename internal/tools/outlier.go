package tools

import (
	"fmt"
	"math"
)

func init() {
	RegisterTool(
		ToolManifest{
			Name:          "OutlierDetection",
			Description:   "Detects outliers in a dataset using the Z-score method. Points beyond the threshold standard deviations from the mean are flagged as outliers.",
			Category:      CategoryAnalysis,
			Documentation: "OutlierDetection.md",
			Parameters: []ParameterDefinition{
				{Name: "dataset", Type: "[]float64", Description: "The dataset to analyze for outliers", Required: true},
				{Name: "threshold", Type: "float64", Description: "Number of standard deviations from mean to consider as outlier (default: 2.0)", Required: false},
			},
			Examples: []string{
				"Detect anomalous sensor readings using default threshold of 2 standard deviations",
				"Find extreme values with threshold of 3 standard deviations for stricter detection",
			},
		},
		outlierWrapper,
	)
}

func outlierWrapper(params map[string]interface{}) (interface{}, error) {
	datasetRaw, ok := params["dataset"]
	if !ok {
		return nil, fmt.Errorf("missing required parameter: dataset")
	}

	var dataset []float64
	switch v := datasetRaw.(type) {
	case []float64:
		dataset = v
	case []interface{}:
		dataset = make([]float64, len(v))
		for i, val := range v {
			f, ok := val.(float64)
			if !ok {
				return nil, fmt.Errorf("dataset[%d] is not a number", i)
			}
			dataset[i] = f
		}
	default:
		return nil, fmt.Errorf("dataset must be an array of numbers")
	}

	threshold := 2.0
	if thresholdRaw, ok := params["threshold"]; ok {
		switch v := thresholdRaw.(type) {
		case float64:
			threshold = v
		case int:
			threshold = float64(v)
		default:
			return nil, fmt.Errorf("threshold must be a number")
		}
	}

	if threshold <= 0 {
		return nil, fmt.Errorf("threshold must be positive")
	}

	return detectOutliers(dataset, threshold), nil
}

// OutlierResult contains the outlier detection results
type OutlierResult struct {
	OutlierIndices []int     `json:"outlier_indices"`
	OutlierValues  []float64 `json:"outlier_values"`
	Mean           float64   `json:"mean"`
	StdDev         float64   `json:"std_dev"`
	Threshold      float64   `json:"threshold"`
	ZScores        []float64 `json:"z_scores"`
	IsOutlier      []bool    `json:"is_outlier"`
}

// detectOutliers finds outliers using Z-score method
func detectOutliers(dataset []float64, threshold float64) OutlierResult {
	n := len(dataset)
	if n == 0 {
		return OutlierResult{
			OutlierIndices: []int{},
			OutlierValues:  []float64{},
			Mean:           0,
			StdDev:         0,
			Threshold:      threshold,
			ZScores:        []float64{},
			IsOutlier:      []bool{},
		}
	}

	var sum float64
	for _, v := range dataset {
		sum += v
	}
	mean := sum / float64(n)

	var sumSquaredDiff float64
	for _, v := range dataset {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}
	stdDev := math.Sqrt(sumSquaredDiff / float64(n))

	outlierIndices := []int{}
	outlierValues := []float64{}
	zScores := make([]float64, n)
	isOutlier := make([]bool, n)

	for i, v := range dataset {
		if stdDev == 0 {
			zScores[i] = 0
			isOutlier[i] = false
		} else {
			zScores[i] = (v - mean) / stdDev
			if math.Abs(zScores[i]) > threshold {
				isOutlier[i] = true
				outlierIndices = append(outlierIndices, i)
				outlierValues = append(outlierValues, v)
			}
		}
	}

	return OutlierResult{
		OutlierIndices: outlierIndices,
		OutlierValues:  outlierValues,
		Mean:           mean,
		StdDev:         stdDev,
		Threshold:      threshold,
		ZScores:        zScores,
		IsOutlier:      isOutlier,
	}
}
