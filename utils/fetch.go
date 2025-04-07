package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"radsched/common"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)


func LoadFunctions() ([]common.FunctionInfo, error) {
	resp, err := http.Get("http://localhost:8000/bootstrap")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var functions []common.FunctionInfo
	if err := json.NewDecoder(resp.Body).Decode(&functions); err != nil {
		return nil, err
	}
	return functions, nil
}

func StoreFunctions(functions []common.FunctionInfo ) (error) {
	file, err := os.Create("/Users/yoni_mindel/Desktop/Thesis/radsched/function_registry.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") 
	if err := encoder.Encode(functions); err != nil {
		return err
	}
	return nil
}

func GetClientToEdgeRTT() ([]common.LocationInfo, error) {
	cmd := exec.Command("/usr/local/bin/python3", "/Users/yoni_mindel/Desktop/Thesis/radsched/ping_edges.py")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to execute ping_regions.py: %v", err)
		return nil, err
	}

	var rttData map[string]float64
	if err := json.Unmarshal(output, &rttData); err != nil {
		log.Fatalf("Failed to unmarshal RTT data: %v", err)
		return nil, err
	}

	var locations []common.LocationInfo
	for region, rtt := range rttData {
		locations = append(locations, common.LocationInfo{
			LocationName: region,
			RoundTripTime: strings.TrimSpace(fmt.Sprintf("%.2f ms", rtt)),
		})
	}

	return locations, nil
}

func StoreLocations(locations []common.LocationInfo) error {
	file, err := os.Create("/Users/yoni_mindel/Desktop/Thesis/radsched/client_edge_rtts.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty-print the JSON
	if err := encoder.Encode(locations); err != nil {
		return err
	}

	return nil
}

func GetFunctionsAsMap()(map[string]common.FunctionInfo, error) {
	file, err := os.Open("/Users/yoni_mindel/Desktop/Thesis/radsched/function_registry.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var functionList []common.FunctionInfo
	if err := json.NewDecoder(file).Decode(&functionList); err != nil {
		return nil, err
	}

	functionMap := make(map[string]common.FunctionInfo)
	for _, function := range functionList {
		functionMap[strings.ToLower(function.FunctionName)] = function
	}

	return functionMap, nil
}

func GetFunctionsAsList()([]common.FunctionInfo, error) {
	file, err := os.Open("/Users/yoni_mindel/Desktop/Thesis/radsched/function_registry.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var functionList []common.FunctionInfo
	if err := json.NewDecoder(file).Decode(&functionList); err != nil {
		return nil, err
	}

	return functionList, nil
}

func GetLocations()(map[string]float64, error) {
	file, err := os.Open("/Users/yoni_mindel/Desktop/Thesis/radsched/client_edge_rtts.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var locationList []common.LocationInfo
	if err := json.NewDecoder(file).Decode(&locationList); err != nil {
		return nil, err
	}

	locationMap := make(map[string]float64)
	for _, location := range locationList {
		rtt, err := strconv.ParseFloat(strings.Split(location.RoundTripTime, " ")[0], 64)
		if (err != nil) {
			log.Fatalf("Error converting string to float64: %v", err)
		}
		locationMap[strings.TrimSpace(strings.ToLower(location.LocationName))] = rtt
	}

	return locationMap, nil
}

func GetEdges()(map[string]map[string]float64, error) {
	edges, err := os.Open("/Users/yoni_mindel/Desktop/Thesis/radsched/edge_datacenter_rtts.json")
	if err != nil {
		return nil, err
	}
	defer edges.Close()

	edgesMap := make(map[string]map[string]float64)
	if err := json.NewDecoder(edges).Decode(&edgesMap); err != nil {
		return nil, err
	}

	edgesLowerMap := make(map[string]map[string]float64)
	for edge, timeMap := range edgesMap {
		lowerEdge := strings.ToLower(edge)
		times := make(map[string]float64)
		for location, rtt := range timeMap {
			times[strings.ToLower(location)] = rtt
		}
		edgesLowerMap[lowerEdge] = times
	}

	return edgesLowerMap, nil
}

func GetEdgeToDataCenterRTT() (map[string]map[string]interface{}, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	client := lambda.NewFromConfig(cfg)

	allRTTData := make(map[string]map[string]interface{})

	for _, region := range common.Datacenters {
		data, err := invokeLambdaInRegion(client, region)
		if err != nil {
			return nil, err
		}
		allRTTData[region] = data
	}

	return allRTTData, nil
}


func invokeLambdaInRegion(client *lambda.Client, region string)  (map[string]interface{}, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client = lambda.NewFromConfig(cfg) 

	input := &lambda.InvokeInput{
		FunctionName: aws.String("PingDatacenters"), 
	}

	output, err := client.Invoke(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var rttData map[string]interface{}
	if err := json.Unmarshal(output.Payload, &rttData); err != nil {
		return nil, err
	}

	return rttData, nil
}

func SaveRTTDataToJSON(data map[string]map[string]interface{}) {
	formattedData := make(map[string]map[string]float64)

	for region, result := range data {
		bodyStr, ok := result["body"].(string)
		if !ok {
			log.Printf("Unexpected body format for region %s", region)
			continue
		}

		var rttMap map[string]float64
		if err := json.Unmarshal([]byte(bodyStr), &rttMap); err != nil {
			log.Printf("Failed to unmarshal RTT data for region %s: %v", region, err)
			continue
		}

		formattedData[region] = rttMap
	}

	file, err := os.Create("/Users/yoni_mindel/Desktop/Thesis/radsched/edge_datacenter_rtts.json")
	if err != nil {
		log.Fatalf("Failed to create RTT data file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(formattedData); err != nil {
		log.Fatalf("Failed to write RTT data to JSON: %v", err)
	}
}


type ConsistencyData struct {
	Edge        string `json:"edge"`
	Function    string `json:"function"`
	NumAttempts int    `json:"num_attempts"`
	NumSuccess  int    `json:"num_success"`
	NumFailure  int    `json:"num_failure"`
}
// getConsistencyWeight fetches the consistency weighting for a specific edge and function
// -1 if error 
// 0 if or no such edge, function combination or no attempts 
// otherwise, num success / num attempts
// func getConsistencyWeight(edge string, function string) (float64, error) {
// 	serverURL := "http://54.219.54.16/cgi-bin/hit_ratio.py"

// 	resp, err := http.Get(serverURL)
// 	if err != nil {
// 		return -1.0, fmt.Errorf("failed to fetch consistency data: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return -1.0, fmt.Errorf("server returned error: %d", resp.StatusCode)
// 	}

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return -1.0, fmt.Errorf("failed to read response body: %v", err)
// 	}

// 	// Use a nested map to match the JSON structure
// 	var data map[string]map[string]map[string]interface{}
// 	err = json.Unmarshal(body, &data)
// 	log.Println(data)
// 	if err != nil {
// 		return -1.0, fmt.Errorf("failed to parse JSON response: %v", err)
// 	}

// 	// Check if the edge exists
// 	edgeData, edgeExists := data[edge]
// 	if !edgeExists {
// 		return 0.5, nil // No such edge, return 0
// 	}

	
// 	// Check if the function exists under this edge
// 	funcData, funcExists := edgeData[function]
// 	if !funcExists {
// 		return 0.5, nil // No such function for this edge, return 0
// 	}

// 	// Extract num_attempts and num_success safely
// 	numAttempts, ok := funcData["num_attempts"].(float64) // JSON numbers are float64
// 	if !ok {
// 		return 0.5, nil // If num_attempts is missing or not a number, return 0
// 	}

// 	numFailure, ok := funcData["num_failure"].(float64)
// 	if !ok {
// 		return 0.5, nil // If num_success is missing or not a number, return 0
// 	}

// 	// Compute and return the consistency weighting
// 	if numAttempts > 0 {
// 		if (numFailure == 0) {
// 			return 0.1, nil
// 		}
// 		return numFailure / numAttempts, nil
// 	}
// 	return 0.5, nil // Function exists, but no attempts were made
// }

func getConsistencyWeight(edge string, function string) (float64, error) {
	file, err := os.Open("/Users/yoni_mindel/Desktop/Thesis/radsched/edge_function_consistency.json")
	if err != nil {
		return -1.0, fmt.Errorf("failed to open edge-function consistency cache: %v", err)
	}
	defer file.Close()

	var data map[string]map[string]FunctionStats
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return -1.0, fmt.Errorf("failed to parse edge-function consistency JSON: %v", err)
	}

	funcsForEdge, edgeExists := data[edge]
	if !edgeExists {
		return 0.5, nil
	}

	stats, funcExists := funcsForEdge[function]
	if !funcExists || stats.NumAttempts == 0 {
		return 0.5, nil
	}

	return float64(stats.NumFailure) / float64(stats.NumAttempts), nil
}

type FunctionStats struct {
	NumAttempts int `json:"num_attempts"`
	NumSuccess  int `json:"num_success"`
	NumFailure  int `json:"num_failure"`
}

func FetchHitRatioByFunction() (map[string]FunctionStats, error) {
	file, err := os.Open("/Users/yoni_mindel/Desktop/Thesis/radsched/function_consistency.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open consistency cache: %v", err)
	}
	defer file.Close()

	var functionData map[string]FunctionStats
	if err := json.NewDecoder(file).Decode(&functionData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return functionData, nil
}


func FetchHitRatioByFunctionRemote() (map[string]FunctionStats, error) {
	url := "http://54.219.54.16/cgi-bin/hit_ratio_v2.py"
	// Make GET request
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse JSON directly into a map[string]FunctionStats
	var functionData map[string]FunctionStats
	err = json.Unmarshal(body, &functionData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return functionData, nil
}

func FetchHitRatioByEdgeRemote() (map[string]map[string]FunctionStats, error) {
	url := "http://54.219.54.16/cgi-bin/hit_ratio.py"
	// Make GET request
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse JSON directly into a ma][string]map[string]FunctionStats
	var edgefunctionData map[string]map[string]FunctionStats
	err = json.Unmarshal(body, &edgefunctionData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return edgefunctionData, nil
}

func StoreFunctionStats(stats map[string]FunctionStats) error {
	file, err := os.Create("/Users/yoni_mindel/Desktop/Thesis/radsched/function_consistency.json")
	if err != nil {
		return fmt.Errorf("failed to create function consistency file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(stats); err != nil {
		return fmt.Errorf("failed to encode function consistency data: %v", err)
	}
	return nil
}

func StoreFunctionStatsByEdge(stats map[string]map[string]FunctionStats) error {
	file, err := os.Create("/Users/yoni_mindel/Desktop/Thesis/radsched/edge_function_consistency.json")
	if err != nil {
		return fmt.Errorf("failed to create edge-function consistency file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(stats); err != nil {
		return fmt.Errorf("failed to encode edge-function consistency data: %v", err)
	}
	return nil
}

func UpdateConsistencyByFunction() error {
	stats, err := FetchHitRatioByFunctionRemote()
	if err != nil {
		return err
	}
	return StoreFunctionStats(stats)
}

func UpdateConsistencyByEdgeFunction() error {
	stats, err := FetchHitRatioByEdgeRemote()
	if err != nil {
		return err
	}
	return StoreFunctionStatsByEdge(stats)
}

