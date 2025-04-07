package utils

import (
	"context"
	"encoding/json"
	"time"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type SyntheticWorkloadResult struct {
	Success       bool    `json:"success"`
	ExecutionTime float64 `json:"execution_time"`
	TotalRuntime  float64 `json:"total_runtime"` // New field for total runtime
}

func RunSyntheticWorkload(region string, sleepTime int) (SyntheticWorkloadResult, error) {
	startTime := time.Now()

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return SyntheticWorkloadResult{}, err
	}
	client := lambda.NewFromConfig(cfg)

	payload, err := json.Marshal(map[string]interface{}{
		"sleep_time": sleepTime,
	})
	if err != nil {
		return SyntheticWorkloadResult{}, err
	}

	input := &lambda.InvokeInput{
		FunctionName: aws.String("SyntheticWorkload"),
		Payload:      payload,
	}
	output, err := client.Invoke(context.TODO(), input)
	if err != nil {
		return SyntheticWorkloadResult{}, err
	}

	var lambdaResponse struct {
		StatusCode int    `json:"statusCode"`
		Body       string `json:"body"`
	}
	if err := json.Unmarshal(output.Payload, &lambdaResponse); err != nil {
		return SyntheticWorkloadResult{}, err
	}
	var result SyntheticWorkloadResult
	if err := json.Unmarshal([]byte(lambdaResponse.Body), &result); err != nil {
		return SyntheticWorkloadResult{}, err
	}
	totalRuntime := time.Since(startTime).Seconds() * 1000
	result.TotalRuntime = totalRuntime
	result.ExecutionTime = result.ExecutionTime * 1000

	return result, nil
}