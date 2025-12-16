package config

// EnvironmentConfig defines the configuration for each environment
type EnvironmentConfig struct {
	CPU             float64
	Memory          float64
	DesiredCount    float64
	MinCount        float64
	MaxCount        float64
	DBInstanceClass string
}

// Environments contains the configuration for all environments
// All environments deploy to us-east-1 only (single region)
var Environments = map[string]EnvironmentConfig{
	"production": {
		CPU:             1024,
		Memory:          2048,
		DesiredCount:    3,
		MinCount:        2,
		MaxCount:        10,
		DBInstanceClass: "r6g.large",
	},
	"sandbox": {
		CPU:             1024,
		Memory:          2048,
		DesiredCount:    2,
		MinCount:        1,
		MaxCount:        5,
		DBInstanceClass: "r6g.medium",
	},
	"demo": {
		CPU:             512,
		Memory:          1024,
		DesiredCount:    1,
		MinCount:        1,
		MaxCount:        2,
		DBInstanceClass: "r6g.medium",
	},
	"uat": {
		CPU:             256,
		Memory:          512,
		DesiredCount:    1,
		MinCount:        1,
		MaxCount:        2,
		DBInstanceClass: "t3.small",
	},
}
