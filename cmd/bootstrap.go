package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"radsched/utils"
)

// Define the bootstrap command
var BootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Fetch the most up-to-date function data",
	Long:  "This command fetches the most recent function data from the /bootstrap endpoint.",
	Run:   bootstrap,
}

func bootstrap(cmd *cobra.Command, args []string) {
	// get registered functions
	functions, err := utils.LoadFunctions()
	if err != nil {
		log.Fatalf("Error fetching function data: %v\n", err)
		return
	}
	err = utils.StoreFunctions(functions)
	if err != nil {
		log.Fatalf("Failed to save function data to JSON: %v", err)
	}
	log.Println("Successfully saved the latest function data to function_registry.json")

	// get client to edge times
	locations, err := utils.GetClientToEdgeRTT()
	if err != nil {
		log.Fatalf("Failed to retreive client to edge data: %v", err)
	}
	err = utils.StoreLocations(locations)
	if err != nil {
		log.Fatalf("Failed to save client to edge data to JSON: %v", err)
	}
	log.Println("Successfully saved the client to edge RTT data")

	// get edge to datacenter times
	data, err := utils.GetEdgeToDataCenterRTT()
	if err != nil {
		log.Fatalf("Error invoking Lambda: %v", err)
	}
	utils.SaveRTTDataToJSON(data)
	log.Println("Successfully saved the edge to datacenter RTT data")

	// get and store consistency data - by function
	if err := utils.UpdateConsistencyByFunction(); err != nil {
		log.Fatalf("Failed to update global consistency data: %v", err)
	}
	log.Println("Updated function-level consistency stats")

	// get and store consistency data - by edge -> function
	if err := utils.UpdateConsistencyByEdgeFunction(); err != nil {
		log.Fatalf("Failed to update edge-function consistency data: %v", err)
	}
	log.Println("Updated edge-function consistency stats")

	
}
