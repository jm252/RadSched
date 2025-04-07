package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	epsilonfile = "epsilon.json"
	ADAPTIVE = "ADAPTIVE" 
	SMOOTH = "SMOOTH" 
)

var (
	epsilonInit = 0.5
	epsilonMin = 0.1   // Minimum exploration rate
	epsilonMax = 0.9   // Maximum exploration
	alpha 	   = 0.3   // Learning rate
)

// Get and update epsilon for a given function and adjustment method
func GetEpsilon(function string, adjustMethod string) (float64, error) {	
	epsilonData, err := LoadEpsilon()
	if (err != nil) {
		return 0.0, err
	}
	epsilon0, exists := epsilonData[strings.ToLower(function)] 
	if (!exists || len(epsilonData) == 0) {
		epsilonData[strings.ToLower(function)] = epsilonInit
		SaveEpsilon(epsilonData)
	}

	// get function hit ratio data
	hitRatios, err := FetchHitRatioByFunction() 
	if err != nil {
		log.Fatalf("Error fetching hit ratio data %v", err)
	}
	functionHitRatio, exists := hitRatios[strings.ToLower(function)]
	if !exists {
		return epsilon0, nil
	}
	successRate := float64(functionHitRatio.NumSuccess) / float64(functionHitRatio.NumAttempts)

	// adjust epsilon proportionally to ratio and num attempts
	var epsilonNew float64
	switch adjustMethod {
	case ADAPTIVE:
		epsilonNew = EpsilonAdjustAdaptive(successRate)
	case SMOOTH:
		epsilonNew = EpsilonAdjustAdaptiveSmooth(epsilon0, successRate)
	default:
		return 0.0, fmt.Errorf("invalid epsilon adjustment method: %s", adjustMethod)
	}	
	
	epsilonData[function] = epsilonNew
	SaveEpsilon(epsilonData)

	return epsilonNew, nil
}

// Get new epsilon inversely proportional to success rate
func EpsilonAdjustAdaptive(successRate float64) float64 {
	epsilonNew := 1 - successRate

	if epsilonNew > epsilonMax {
		epsilonNew = epsilonMax
	}
	if epsilonNew < epsilonMin {
		return epsilonMin
	}
	
	return epsilonNew
}

// Update epsilon inversely proportional to success rate with a smoothing factor
func EpsilonAdjustAdaptiveSmooth(epsilon0 float64, successRate float64) float64 {
	targetEpsilon := 1 - successRate
	epsilonNew := epsilon0 + alpha *(targetEpsilon - epsilon0)

	if epsilonNew > epsilonMax {
		epsilonNew = epsilonMax
	}
	if epsilonNew < epsilonMin {
		return epsilonMin
	}
	
	return epsilonNew
}

// Fetches epsilon data 
func LoadEpsilon() (map[string]float64, error) {
	// If file doesn't exist, return an empty map
	if _, err := os.Stat(epsilonfile); os.IsNotExist(err) {
		return make(map[string]float64), nil
	}

	// Read file contents
	data, err := os.ReadFile(epsilonfile)
	if err != nil {
		return nil, fmt.Errorf("failed to read epsilon file: %v", err)
	}

	// Place contents in map
	epsilonData := make(map[string]float64)
	err = json.Unmarshal(data, &epsilonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse epsilon JSON: %v", err)
	}

	return epsilonData, nil
}

// Writes epsilon values to file 
func SaveEpsilon(epsilonData map[string]float64) error {
	data, err := json.MarshalIndent(epsilonData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal epsilon JSON: %v", err)
	}

	err = os.WriteFile(epsilonfile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write epsilon file: %v", err)
	}

	return nil
}