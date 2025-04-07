package main

import (
	"radsched/cmd"
	// "radsched/utils"
	// "fmt"
)

func main() {
	cmd.InitRootCmd()
	cmd.Execute()

	// result, err := utils.RunSyntheticWorkload("us-east-2", 2)
	// if (err != nil) {
	// 	fmt.Println(err)
	// }
	// fmt.Println(result)
}