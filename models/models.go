package models

type Application struct {
	Name    string
	Version string
}

type VectorLogsStruct struct {
	Ddsource      string      `json:"ddsource"`
	Ddtags        interface{} `json:"ddtags"`
	Hostname      string      `json:"hostname"`
	Message       interface{} `json:"message"`
	Service       string      `json:"service_name"`
	SourceType    string      `json:"source_type"`
	Env           string      `json:"env"`
	ComponentName string      `json:"component_name"`
}

type PayloadStruct struct {
	Hostname      string `json:"hostname"`
	Org           string `json:"org"`
	Env           string `json:"env"`
	ServiceName   string `json:"service_name"`
	ComponentName string `json:"component_name"`
	Account       string `json:"account"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Since         string `json:"since"`
}

type UserLogStruct struct {
	Hostname string `json:"hostname"`
	Command  string `json:"command"`
}

type UserLogCommandStruct struct {
	Org           string `json:"org"`
	Env           string `json:"env"`
	ServiceName   string `json:"service_name"`
	ComponentName string `json:"component_name"`
	Account       string `json:"account"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Since         string `json:"since"`
}

type LogSearchConfig struct {
	KafkaBrokerHost       string `json:"kafkaBrokerHost"`
	Topic                 string `json:"topic"`
	MaxRetentionMinutes   int    `json:"maxRetentionMinutes"`
	LogSearchGrafanaUrl   string `json:"logSearchGrafanaUrl"`
	LiveLogAgentHost      string `json:"liveLogAgentHost"`
	LiveLogAgentSshPemKey string `json:"liveLogAgentSshPemKey"`
	LiveLogAgentSshUser   string `json:"liveLogAgentSshPemUser"`
	LiveLogAgentSshPort   int    `json:"liveLogAgentSshPemPort"`
	IsLowerEnv            bool   `json:"isLowerEnv"`
	Tenant                string `json:"tenant"`
}

type LogsCommandArgs struct {
	Env             string
	Account         string
	ServiceName     string
	ComponentName   string
	ComponentType   string
	Org             string
	CloudProvider   string
	StartTime       string
	EndTime         string
	Since           string
	LinuxOperation  string
	AllowedDdTags   bool
	ShowTags        string
	LogSearchConfig LogSearchConfig
}
