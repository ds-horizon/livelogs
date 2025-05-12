package constants

import "time"

const (
	ArgumentServiceName           = "service_name"
	ArgumentComponentName         = "component_name"
	ArgumentComponentType         = "component_type"
	ArgumentEnv                   = "env"
	ArgumentOrg                   = "org"
	ArgumentCloudProvider         = "cloud_provider"
	ArgumentAccount               = "account"
	ArgumentStartTime             = "start_time"
	ArgumentEndTime               = "end_time"
	ArgumentSince                 = "since"
	ArgumentLinuxOperation        = "linux_operation"
	ArgumentShowTags              = "show_tags"
	ArgumentVerbose               = "verbose"
	LogSearchConfig               = "log_search_config"
	GlobalLogsCommandTimeout      = 10 * time.Minute
	EnvLivelogsUser               = "livelogs-user"
	CentralLiveLogAgentSshTimeout = 5 * time.Second
	CentralLiveLogAgentHost       = "http://log-central-orchestrator.dss-platform.com"
	KafkaBrokerPort               = "9092"
	EmptyJSON                     = "{}"
	CentralLiveLogAgentName       = "central-livelogs"
	AwsMetadataUrl                = "http://169.254.169.254/latest/meta-data"
	GcpMetadataUrl                = "http://metadata.google.internal/computeMetadata/v1/"
)
