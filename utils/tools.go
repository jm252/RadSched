package utils

func RemoveLocationsExcept(locations map[string]float64, exceptions []string) (map[string]float64) {
	exceptionSet := make(map[string]int)
	for _, key := range exceptions {
		exceptionSet[key] = 1
	}
	for key := range locations {
		if _, exists := exceptionSet[key]; !exists {
			delete(locations, key)
		}
	}
	return locations
}