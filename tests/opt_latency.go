package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"radsched/cmd"
	"radsched/common"
	"radsched/utils"
)

func algorithm_test() {
		// test functions
		TEST_FUNCTION_1 := "Function9"
		TEST_FUNCTION_2 := "Function4"
	
		functionMap, err := utils.GetFunctionsAsMap()
		if err != nil {
			log.Print(err)
		}
		function1 := functionMap[strings.ToLower(TEST_FUNCTION_1)]
		function2 := functionMap[strings.ToLower(TEST_FUNCTION_2)]
	
		// get runtime if run in datacenter
		datacenterMap, err := utils.GetLocations()
		if err != nil {
			log.Print(err)
		}
		clientDatacenter1Time := datacenterMap[function1.Datacenter]
		clientDatacenter2Time := datacenterMap[function2.Datacenter]
		function1ExuctionTime, err := strconv.ParseFloat(strings.Split(function1.ExecutionTime, "m")[0], 64)
		function2ExuctionTime, err := strconv.ParseFloat(strings.Split(function2.ExecutionTime, "m")[0], 64)
	
		function1DatacenterRuntime := clientDatacenter1Time + function1ExuctionTime
		function2DatacenterRuntime := clientDatacenter2Time + function2ExuctionTime
	
		// get runtime if run on edge with RadSched
		cmd := exec.Command("./radsched", "run", TEST_FUNCTION_1)
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			log.Print(err)
		}
		output := out.String()
		fmt.Println("Function 9 Datacenter Total Runtime: ", function1DatacenterRuntime)
		fmt.Println("Function 9 RadSched Total Runtime: ", output)
	
		cmd = exec.Command("./radsched", "run", TEST_FUNCTION_2)
		var out1 bytes.Buffer
		cmd.Stdout = &out1
		err = cmd.Run()
		if err != nil {
			log.Print(err)
		}
		output = out1.String()
		fmt.Println("Function 4 Datacenter Total Runtime: ", function2DatacenterRuntime)
		fmt.Println("Function 4 RadSched Total Runtime: ", output)
}

// 1) excute function at radsched optimal location
// 2) execute function at datacenter 
// 3) compare results
func opt_latency_test() {
	// get test function
	TEST_FUNCTIONS := make([]string, 0, len(common.TEST_FUNCTION_MAP))
	for function := range common.TEST_FUNCTION_MAP {
		TEST_FUNCTIONS = append(TEST_FUNCTIONS, function)
	}
	functionMap, _:= utils.GetFunctionsAsMap()

	// get radsched optimal locations
	opt_locations := make([]string, len(TEST_FUNCTIONS))
	for i := 0; i < len(TEST_FUNCTIONS); i++ {
		radSchedResult := cmd.RunFunction(TEST_FUNCTIONS[i], false)
		opt_locations[i] = radSchedResult.OptLocation; 
	}

	// run function in radsched optimal location and in datacenter
	for i := 0; i < len(TEST_FUNCTIONS); i++ {
		sleepTime := common.TEST_FUNCTION_MAP[TEST_FUNCTIONS[i]]
		fmt.Println("Running for", TEST_FUNCTIONS[i], " for ", sleepTime, "ms ", "at location", opt_locations[i], " and datacenter", functionMap[TEST_FUNCTIONS[i]].Datacenter)
		radResult, err := utils.RunSyntheticWorkload(opt_locations[i], int(sleepTime))
		if (err != nil) {
			log.Println(err)
		}
		dcResult, err := utils.RunSyntheticWorkload(functionMap[TEST_FUNCTIONS[i]].Datacenter, int(sleepTime))
		if (err != nil) {
			log.Println(err)
		}
		fmt.Println("Function: ", TEST_FUNCTIONS[i])
		fmt.Println()
		fmt.Println("RadSched optimal location, ", opt_locations[i])
		fmt.Println("RadSched optimal runtime, ", radResult.ExecutionTime)
		fmt.Println()
		fmt.Println("Datacenter location, ", functionMap[TEST_FUNCTIONS[i]].Datacenter)
		fmt.Println("Datacenter runtime, ", dcResult.ExecutionTime)
	}
}


type TestResults struct {
	InputTime float64
	ExecutionTime float64
	TotalRuntime float64
	OverheadTime float64
	ExpectedTransitTime float64
}
var datacenter_results = make(map[string]TestResults)

func test_execution_across_datacenters() {
	datacenters := common.Datacenters
	
	// calcualte average runtime in all datacenter for the same function
	sleepTime := 1
	for i := 0; i < 10; i++ {
		for _, region := range datacenters {
			fmt.Printf("Testing Lambda in region: %s\n", region)
			result, err := utils.RunSyntheticWorkload(region, sleepTime)
			if err != nil {
				log.Printf("Error in region %s: %v\n", region, err)
				continue
			}
			fmt.Printf("Lambda execution time in region %s: %.2f ms\n", region, result.ExecutionTime)
			fmt.Printf("Lambda total runtime in region %s: %.2f ms\n", region, result.TotalRuntime)
			if (i == 0) {
				datacenter_results[region] = TestResults{
					InputTime: float64(sleepTime * 1000),
					ExecutionTime: result.ExecutionTime,
					TotalRuntime: result.TotalRuntime,
				}
			} else {
				currentResults := datacenter_results[region]
				currentResults.ExecutionTime = (currentResults.ExecutionTime*float64(i) + result.ExecutionTime) / float64(i+1)
				currentResults.TotalRuntime = (currentResults.TotalRuntime*float64(i) + result.TotalRuntime) / float64(i+1)
				datacenter_results[region] = currentResults
			}
		}
	}

	// calculate min, max, and median 

	// calculate amnount of overhead 
	rtt_map, _ := utils.GetLocations()
	for region, result := range datacenter_results {
		transit := rtt_map[region]
		result.ExpectedTransitTime = transit 
		result.OverheadTime = result.TotalRuntime - result.ExecutionTime
		datacenter_results[region] = result
	}

	for key, value := range datacenter_results {
		fmt.Println(key)
		fmt.Println(value)
	}
}

func main() {
	test_execution_across_datacenters()
	// opt_latency_test()
}

// 
