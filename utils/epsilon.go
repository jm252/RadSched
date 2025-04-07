package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
)

const (
	epsilonfile = "epsilon.json"
	EDS = "EDS" // Exponential Decay Scaling
	ASA = "ASA" // Adaptive Simulated Annealing 
	ADAPTIVE = "ADAPTIVE" 
	SMOOTH = "SMOOTH" 
)

var (
	epsilonInit = 0.5
	epsilonMin = 0.1   // Minimum exploration rate
	epsilonMax = 0.9   // maximum exploration
	alpha 	   = 0.3 	// learning rate
	l     	   = 0.005 // Decay rate l 
	g      	   = 5.0   // Smoothing factor g
)


func GetEpsilon(function string, adjustMethod string) (float64, error) {	
	// get current epsilon
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
	numAttempts := functionHitRatio.NumAttempts

	// adjust epsilon proportionally to ratio and num attempts
	var epsilonNew float64
	switch adjustMethod {
	case ADAPTIVE:
		epsilonNew = EpsilonAdjustAdaptive(successRate)
	case SMOOTH:
		epsilonNew = EpsilonAdjustAdaptiveSmooth(epsilon0, successRate)
	case EDS:
		epsilonNew = EpsilonAdjustEDS(epsilon0, successRate, float64(numAttempts))
	case ASA:
		epsilonNew = EpsilonAdjustASA(epsilon0, successRate, numAttempts)
	default:
		return 0.0, fmt.Errorf("invalid epsilon adjustment method: %s", adjustMethod)
	}	
	
	// save epsilon 
	epsilonData[function] = epsilonNew
	SaveEpsilon(epsilonData)

	// return epsilon
	return epsilonNew, nil
}

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

func EpsilonAdjustEDS(epsilon0 float64, successRate float64, numAttempts float64) float64 {
	// epsilonNew := epsilon0 * math.Exp(l * float64(numAttempts)) * (1 - successRate + g)
	log.Println(successRate)
	log.Println(numAttempts)
	epsilonNew := epsilonMin + (epsilonMax - epsilonMin) * math.Exp(-l * (float64(numAttempts) + g * (1 - successRate)))

	// Ensure epsilon does not exceed 1
	if epsilonNew > epsilonMax {
		epsilonNew = epsilonMax
	}
	if epsilonNew < epsilonMin {
		return epsilonMin
	}
	
	return epsilonNew
}

func EpsilonAdjustASA(epsilon0, successRate float64, numAttempts int) float64 {
	epsilonNew := epsilon0 * math.Exp(-l * float64(numAttempts) * successRate)

	if epsilonNew < epsilonMin {
		epsilonNew = epsilonMin
	}

	return epsilonNew
}

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

// Saveepsilon writes the epsilon values back to the JSON file
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