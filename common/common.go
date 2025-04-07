package common

type FunctionInfo struct {
	FunctionName  string `json:"function_name"`
	ExecutionTime string `json:"execution_time"`
	FunctionURL   string `json:"function_url"`
	Datacenter    string `json:"datacenter"`
}

type LocationInfo struct {
	LocationName  string `json:"location_name"`
	RoundTripTime string `json:"round_trip_time"`
}

type ExecutionInfo struct {
	OptLocation   string
	ExecutionTime float64
	Cost          string
}

var All_Datacenters = []string{
	"us-east-1", "us-east-2", "us-west-1", "us-west-2",
	"ca-central-1", "ca-west-1", "eu-west-1", "eu-west-2", "eu-west-3",
	"eu-central-1", "eu-central-2", "eu-south-1", "eu-south-2", "eu-north-1",
	"il-central-1", "me-south-1", "me-central-1", "af-south-1", "ap-east-1",
	"ap-south-1", "ap-south-2", "ap-northeast-1", "ap-northeast-2",
	"ap-northeast-3", "ap-southeast-1", "ap-southeast-2", "ap-southeast-3",
	"ap-southeast-4", "sa-east-1", "cn-north-1", "cn-northwest-1",
	"us-gov-east-1", "us-gov-west-1",
}

var Datacenters = []string{
	"us-west-1", 
	"us-east-1", 
	"us-west-2", 
	"us-east-2", 
	"ap-east-1",
	"ap-south-1", 
	"ap-northeast-1", 
	"ap-northeast-2", 
	"ap-northeast-3",
}

var TEST_FUNCTION_MAP = map[string]float64{
    "function1" : 5,
    "function2" : 25,
    "function3" : 125,
    "function4" : 625,
    "function5" : 3125,
    "function6" : 5,
    "function7" : 25,
    "function8" : 125,
    "function9" : 625,
    "function10" : 3125,
}
