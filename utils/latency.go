package utils

import (
	"log"
	"math"
	"radsched/common"
	"strconv"
	"strings"
	"math/rand"
)

// experiment locations
var expLocations = []string{"us-west-1", "us-east-1", "us-east-2"} 

// Choose and return optimal executiuon location based on latency  
func RunOptLatency(function common.FunctionInfo) common.ExecutionInfo {
	// get time from datacenter to edges
	edges, err := GetEdges()
	if (err != nil) {
		log.Fatalf("Failed to load edge data: %v", err)
	}
	edgeToDatacenters := edges[function.Datacenter]

	// get time from client to all nodes 
	locations, err := GetLocations()
	if (err != nil) {
		log.Fatalf("Failed to load edge data: %v", err)
	}

	executionTime, err := strconv.ParseFloat(strings.Split(function.ExecutionTime, "m")[0], 64)
	if (err != nil) {
		log.Fatalf("Error converting string to float64: %v", err)
	}
	datacenterRuntime := locations[function.Datacenter] + executionTime
	
	// get optimal node 
	var optEdge string
	optEdgeTime := math.MaxFloat64
	for edge, clientToEdge := range locations {
		currentEdgeTime := clientToEdge + max(executionTime, edgeToDatacenters[edge])
		if (currentEdgeTime < optEdgeTime) {
			optEdgeTime = currentEdgeTime
			optEdge = edge
		}
	}

	if (optEdgeTime > datacenterRuntime) {
		return common.ExecutionInfo{
			OptLocation: function.Datacenter,
			ExecutionTime: datacenterRuntime,
		}
	}
	
	return common.ExecutionInfo{
		OptLocation: optEdge,
		ExecutionTime: optEdgeTime,
	}
}

// Choose and return optimal executiuon location based on latency and consistency
func RunOptWeightedLatency(function common.FunctionInfo) common.ExecutionInfo {
	// get time from datacenter to edges
	edges, err := GetEdges()
	if (err != nil) {
		log.Fatalf("Failed to load edge data: %v", err)
	}
	edgeToDatacenters := edges[function.Datacenter]
	edgeToDatacentersExp := RemoveLocationsExcept(edgeToDatacenters, expLocations)

	// get time from client to all nodes 
	locations, err := GetLocations()
	if (err != nil) {
		log.Fatalf("Failed to load edge data: %v", err)
	}
	locationsExp := RemoveLocationsExcept(locations, expLocations)

	executionTime, err := strconv.ParseFloat(strings.Split(function.ExecutionTime, "m")[0], 64)
	if (err != nil) {
		log.Fatalf("Error converting string to float64: %v", err)
	}
	datacenterRuntime := locationsExp[function.Datacenter] + executionTime

	// get eligible nodes
	eligibleNodes := make([]common.ExecutionInfo, 0); 
	for edge, clientToEdge := range locationsExp {
		currentEdgeTime := clientToEdge + max(executionTime, edgeToDatacentersExp[edge])
		if (currentEdgeTime < datacenterRuntime) {
			eligibleNodes = append(eligibleNodes, common.ExecutionInfo{
										OptLocation: edge, 
										ExecutionTime: currentEdgeTime})
		}
	}

	// if no eligble nodes, run in datacenter
	if (len(eligibleNodes) == 0) {
		return common.ExecutionInfo{
			OptLocation: function.Datacenter,
			ExecutionTime: datacenterRuntime,
		}
	}

	// With probability epsilon, choose random from eligible nodes, otherwise compute optimal
	var optEdge string
	var optEdgeTime float64
	weightedLatency := math.MaxFloat64
	epsilon, err := GetEpsilon(function.FunctionName, SMOOTH)
	if err != nil {
		log.Fatalf("Error calcuating epsilon: %v", err)
	}
	isExploreAction := getAction(epsilon)
	
	if isExploreAction {
		randEdge := rand.Intn(len(eligibleNodes))
		optEdge = eligibleNodes[randEdge].OptLocation
		optEdgeTime = eligibleNodes[randEdge].ExecutionTime
	} else {
		for _, edgeInfo := range eligibleNodes {
			weighting, err := getConsistencyWeight(edgeInfo.OptLocation, function.FunctionName)
			if (err != nil) {
				log.Fatalf("Error converting string to float64: %v", err)
			}
			currentWeightedLatency := edgeInfo.ExecutionTime * weighting; 
			if (currentWeightedLatency < weightedLatency) {
				weightedLatency = currentWeightedLatency
				optEdgeTime = edgeInfo.ExecutionTime;
				optEdge = edgeInfo.OptLocation; 
			}
		}
	}

	return common.ExecutionInfo{
		OptLocation: optEdge,
		ExecutionTime: optEdgeTime,
	}
}

func getAction(epsilon float64) (bool) {
	return rand.Float64() < epsilon
}