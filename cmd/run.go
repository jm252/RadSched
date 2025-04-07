package cmd

import (
	"log"
	"strings"
	"radsched/common"
	"radsched/utils"
	"fmt"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run [function name]",
	Short: "Run a specific function on the Rad-Sched scheduler",
	Long:  "This command runs a specific function at the optimal edge location.",
	Args:  cobra.ExactArgs(1), 
	Run: func(cmd *cobra.Command, args []string) {
		functionName := strings.ToLower(args[0])
		withWeight, _ := cmd.Flags().GetBool("with-weight")
		RunFunction(functionName, withWeight)
	},
}

// Run the function by choosing the optimal execution location
func RunFunction(functionName string, withWeight bool) (common.ExecutionInfo) {
	// fetch function information
	functions, err := utils.GetFunctionsAsMap()
	if err != nil {
		log.Fatalf("Failed to fetch function info: %v", err)
	}
	function, exists := functions[functionName]
	if (!exists) {
		log.Fatalf("Function %s is unkown. Please prepare before running", functionName)
	} 

	// calculate optimal location
	var executionInfo common.ExecutionInfo
	if (withWeight) {
		executionInfo = utils.RunOptWeightedLatency(function)
	} else {
		executionInfo = utils.RunOptLatency(function)
	}

	fmt.Printf("Function Name: %s\n", functionName)
	fmt.Printf("Optimal Location: %s\n", executionInfo.OptLocation)
	fmt.Printf("Execution Time: %f\n", executionInfo.ExecutionTime)

	return executionInfo 
}
