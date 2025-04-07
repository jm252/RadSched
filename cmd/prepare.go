package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"radsched/common"
	"radsched/utils" 
	"github.com/spf13/cobra"
)

type PreparedFunction struct {
	FunctionName  string `json:"function_name"`
	ExecutionTime string `json:"execution_time"`
	FunctionURL   string `json:"function_url"`
	Datacenter    string `json:"datacenter"`
	Date          string `json:"date"`
}

var PrepareCmd = &cobra.Command{
	Use:   "prepare [function name] [function execution time (ms)] [function datacenter]",
	Short: "Prepare a specific function to be executed on the Rad-Sched scheduler",
	Long:  "This command prepares a specific function by adding it to both the local and Radical function repositories.",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		function := common.FunctionInfo{
			FunctionName:  strings.ToLower(args[0]),
			ExecutionTime: args[1],
			Datacenter:    strings.ToLower(args[2]),
		}
		if err := saveToLocalFunctionRegistry(function); err != nil {
			log.Fatalf("Failed to save to local function registry: %v", err)
		}
		if err := registerFunctionWithRadical(function); err != nil {
			log.Fatalf("Failed to register function with Radical: %v", err)
		}
		fmt.Println("Function successfully prepared and registered!")
	},
}

// Registers or updates function in local registry 
func saveToLocalFunctionRegistry(function common.FunctionInfo) error {
	functionsMap, err := utils.GetFunctionsAsMap()
	if err != nil {
		return fmt.Errorf("failed to load existing functions: %v", err)
	}
	functionsList, err := utils.GetFunctionsAsList()
	if err != nil {
		return fmt.Errorf("failed to load existing functions: %v", err)
	}
	_, exists := functionsMap[function.FunctionName] 
	if (exists) {
		for i, existingFunction := range functionsList {
			if strings.ToLower(existingFunction.FunctionName) == function.FunctionName {
				functionsList[i] = function // Update the existing function
				fmt.Printf("Function '%s' updated in the local registry.\n", function.FunctionName)
				break
			}
		}
	} else {
		functionsList = append(functionsList, function)
		fmt.Printf("Function '%s' added to the local registry.\n", function.FunctionName)
		log.Println(functionsList)
	}

	registry, err := os.Create("function_registry")
	if err != nil {
		return err
	}
	defer registry.Close()

	encoder := json.NewEncoder(registry)
	encoder.SetIndent("", "  ") 
	if err := encoder.Encode(functionsList); err != nil {
		return err
	}

	return nil
}

// Registers or updates function in Radical registry 
func registerFunctionWithRadical(function common.FunctionInfo) error {
	preparedFunction := PreparedFunction{
		FunctionName: function.FunctionName,
		ExecutionTime: function.ExecutionTime,
		FunctionURL: "http",
		Datacenter: function.Datacenter,
		Date: "0000-00-00",
	}
	functionData, err := json.Marshal(preparedFunction)
	if err != nil {
		return fmt.Errorf("failed to marshal function data: %v", err)
	}

	resp, err := http.Post("http://localhost:8000/register", "application/json", bytes.NewBuffer(functionData))
	if err != nil {
		return fmt.Errorf("failed to send request to Radical: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("radical registration failed, status code: %d", resp.StatusCode)
	}

	fmt.Printf("Function '%s' registered with Radical.\n", function.FunctionName)
	return nil
}
