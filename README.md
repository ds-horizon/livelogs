# livelogs

<img src="./docs/livetail.png" alt="Livelogs" title="Livelogs: Internal CLI to check logs on environments." width="50%" height="50%">

A powerful CLI tool for real-time log streaming and historical log analysis across multiple environments and cloud providers. Livelogs provides seamless access to service logs with advanced filtering, search capabilities, and support for both application and Auto Scaling Group (ASG) logs.

## 🚀 Features

- **Real-time log streaming** from Kafka brokers
- **Historical log analysis** with flexible time range queries
- **Multi-environment support** (prod, staging, development, etc.)
- **Multi-cloud compatibility** (AWS, GCP)
- **Organization-aware** (d11, dp, hulk)
- **Advanced filtering** with Linux operations and regex support
- **Tag-based filtering** for structured log analysis
- **ASG log monitoring** for infrastructure events
- **Secure SSH tunneling** to central log agents
- **Auto-update notifications** via GitHub releases

## 📦 Installation

### Via Homebrew (Recommended)

```shell
brew install dream11/tools/livelogs
livelogs configure
```

> **Supported Platforms:** macOS (Intel & Apple Silicon), Linux AMD64

### Manual Installation

1. Download the latest release from [GitHub Releases](https://github.com/dream-sports-labs/livelogs/releases)
2. Extract and move to your PATH:
   ```shell
   tar -xzf livelogs_*.tar.gz
   sudo mv livelogs /usr/local/bin
   ```
3. Verify installation:
   ```shell
   livelogs --version
   ```

## 🛠️ Development Setup

### Prerequisites

- **Go 1.22+** - [Download here](https://golang.org/dl/)
- **Git** - For cloning the repository

### Setup Steps

1. **Clone the repository:**
   ```shell
   git clone https://github.com/dream11/livelogs
   cd livelogs
   ```

2. **Install dependencies:**
   ```shell
   go mod download
   ```

3. **Verify the setup:**
   ```shell
   go run main.go --version
   ```

### Build & Install from Source

#### Quick Install
```shell
make install
```

#### Manual Build
```shell
# Build for current platform
go build .

# Install globally
sudo mv ./livelogs /usr/local/bin

# Verify installation
livelogs --version
```

#### Cross-Platform Build
```shell
# Build for all supported platforms
make build

# Create compressed distributions
make compressed-builds
```

## 📋 Usage

### Command Structure

All commands follow the pattern:
```shell
livelogs <command> [flags]
```

### Primary Commands

#### `logs` - Stream/View Service Logs

The main command for accessing service logs with comprehensive filtering options.

**Basic Usage:**
```shell
livelogs logs --env <environment> [additional-flags]
```

**Required Flags:**
- `--env, -e` : Environment name (mandatory)

**Optional Flags:**
- `--service_name, -s` : Filter by specific service name
- `--component_name, -c` : Filter by component name
- `--component_type` : Component type (`application` or `asg`) [default: application]
- `--asg_name` : ASG name (for ASG logs)
- `--org, -o` : Organization (`d11`, `dp`, `hulk`) [default: d11]
- `--cloud_provider` : Cloud provider (`aws`, `gcp`) [default: aws]
- `--account, -a` : Account type (`prod`, `load`, `stag`)
- `--start_time` : Start time for historical logs (format: "2006-01-02 15:04:05" IST)
- `--end_time` : End time for historical logs (format: "2006-01-02 15:04:05" IST)
- `--since` : Time duration from now (e.g., `10m`, `1h`, `30s`)
- `--linux_operation, -l` : Linux operations for log processing
- `--show_tags` : Comma-separated list of ddtags to display
- `--verbose, -v` : Enable verbose logging for debugging

### 📖 Examples

#### Real-time Log Streaming
```shell
# Stream logs for a service in production
livelogs logs --service_name demo-service --component_name demo-component --env prod

# Stream logs with organization and cloud provider
livelogs logs --service_name demo-service --env prod --org d11 --cloud_provider aws
```

#### Historical Log Analysis
```shell
# View logs from the last 30 minutes
livelogs logs --service demo-service --component_name demo-component --env prod --since 30m

# View logs from the last 2 hours
livelogs logs --service demo-service --component_name demo-component --env prod --since 2h

# View logs for a specific time range
livelogs logs --service demo-service  --component_name demo-component --env prod \
  --start_time "2023-12-01 10:00:00" \
  --end_time "2023-12-01 11:00:00"
```

#### Advanced Filtering
```shell
# Filter logs using Linux operations
livelogs logs --service demo-service --component_name demo-component --env prod \
  --linux_operation 'grep "error" | grep -v "user_error"'

# Show only specific tags
livelogs logs --service demo-service --component_name demo-component --env prod \
  --show_tags "version,region,instance_id"

# Combine multiple filters
livelogs logs --service demo-service --component_name demo-component --env prod --since 1h \
  --linux_operation 'grep -i "exception"' \
  --show_tags "service,host"
```

#### ASG Log Monitoring
```shell
# Monitor Auto Scaling Group events
livelogs logs --component_type asg --asg_name my-asg --account prod

# ASG logs with time filter
livelogs logs --component_type asg --asg_name my-asg --account prod --since 15m
```

#### Multi-Organization Examples
```shell
# Dream11 production logs
livelogs logs --service my-service --component_name demo-component --env prod --org d11

# DreamPay production logs
livelogs logs --service payment-service --component_name demo-component --env prod --org dp
```

#### Short Flag Examples
```shell
# Using short flags for quick access
livelogs logs -s demo-service -c demo-component -e prod -o d11

# Combine short and long flags
livelogs logs -s demo-service -c demo-component -e prod --since 1h -l 'grep "error"'
```

### 🔍 Log Output Format

#### Application Logs
```
service_name    hostname               ddtags                                                message
demo-service    i-086fd8e72a71a55c3    {"version":"1.2.3", "sourcecategory":"sourcecode"}    Application started successfully
```

#### ASG Logs
```json
{
   "accountId": "1234567",
   "autoScalingGroupName": "test-asg",
   "details": "{\"Availability Zone\":\"us-east-1a\"",
   "activityId": "31b66071-3270-a01b-db34-a70fab729c8e",
   "requestId": "31b66071-3270-a01b-db34-a70fab729c8e",
   "progress": "50",
   "event": "autoscaling:EC2_INSTANCE_LAUNCH",
   "statusCode": "InProgress",
   "description": "Launching a new EC2 instance: i-02736tf4dcb",
   "cause": "At 2025-07-18T06:50:27Z a user request explicitly set group desired capacity changing the desired capacity from 35 to 44.",
   "startTime": "2025-07-18T06:50:39.528Z",
   "endTime": "2025-07-18T06:51:10.904Z",
   "ec2InstanceId": "i-02736tf4dcb"
}
```

## 🔧 Configuration

### Environment-Based Account Mapping

Livelogs automatically determines the account type based on environment names:
- `prod`, `uat` environments → `prod` account
- Other environments → Based on explicit `--account` flag

### Supported Organizations

| Org Code | Description |
|----------|-------------|
| `d11`    | Dream11 (default) |
| `dp`     | DreamPay |
| `hulk`   | Hulk |

### Supported Cloud Providers

| Provider | Description |
|----------|-------------|
| `aws`    | Amazon Web Services (default) |
| `gcp`    | Google Cloud Platform |

## 🤝 Contributing

### Code Conventions

1. **Naming Standards:**
   - Variables and functions: `camelCase`
   - Exported symbols: `ExportedName`
   - CLI parameters: `parameter-name`

2. **Project Layout:**
   - Follow [Go project layout standards](https://github.com/golang-standards/project-layout)

3. **CLI Parameter Example:**
   ```go
   logsCmd.Flags().StringP("service-name", "s", "", "service name")
   logsCmd.Flags().StringP("env", "e", "", "environment name")
   ```

### Development Workflow

#### Code Formatting & Linting

1. **Install linting tools:**
   ```shell
   brew install golangci-lint
   ```
2. **Upgrade to its latest version:**
   ```shell
   brew upgrade golangci-lint
   ```

3. **Run linter:**
   ```shell
   make lint
   ```

### Note
> **Note:** To fix gci errors run following commands:
```bash
go install github.com/daixiang0/gci@v0.11.0
gci -w -local github.com/daixiang0/gci main.go
gci write --skip-generated -s standard -s default .
```

#### Pre-commit Hooks

1. **Install pre-commit:**
   ```shell
   pip install pre-commit
   ```

2. **Setup hooks:**
   ```shell
   cd livelogs && pre-commit install
   ```

3. **Make commits:**
   ```shell
   git add .
   git commit -m "Your commit message"
   ```

> Pre-commit hooks will automatically validate Go code and suggest improvements.

### Version Management

Livelogs follows [Semantic Versioning](https://semver.org/) (`x.y.z`):

- **Patch** (`x.y.z+1`): Bug fixes and patches
- **Minor** (`x.y+1.0`): New features, backward compatible
- **Major** (`x+1.0.0`): Breaking changes, major features

**Location:** Version is maintained in [`app/app.go`](./app/app.go) in the `App` variable.

**Example:**
```go
var App = models.Application{
    Name:    "livelogs",
    Version: "0.2.2",
}
```

> Update the version responsibly as it triggers release creation on main branch merges.

## 🆘 Troubleshooting

### Common Issues

#### Connection Problems
```shell
# Enable verbose logging for debugging
livelogs logs --service my-service --component_name my-component --env prod --verbose
```

#### DNS Resolution Issues
- Ensure you're connected to the correct VPN
- Verify network connectivity to log infrastructure

#### Permission Errors
- Check if your user has access to the specified environment
- Verify organization permissions

#### Historical Log Limits
```
Error: We store only past X minutes data for livelogs
```
- Use Grafana for older logs (link provided in error message)
- Reduce the time range for your query

### Getting Help

1. **Enable verbose mode** for detailed error information:
   ```shell
   livelogs logs --verbose [other-flags]
   ```

2. **Check version and update**:
   ```shell
   livelogs --version
   brew upgrade dream11/tools/livelogs
   ```

3. **Review command syntax**:
   ```shell
   livelogs logs --help
   ```

## 📚 API Reference

### Command Line Interface

The CLI uses the [Cobra](https://github.com/spf13/cobra) framework for command-line parsing and provides a user-friendly interface.

#### Global Flags
- `--version` : Display version information
- `--help` : Show command help

#### Logs Command Flags
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--env` | `-e` | string | - | Environment name (required) |
| `--service_name` | `-s` | string | - | Service name filter |
| `--component_name` | `-c` | string | - | Component name filter |
| `--component_type` | - | string | `application` | Component type (`application`, `asg`) |
| `--asg_name` | - | string | - | ASG name for ASG logs |
| `--org` | `-o` | string | `d11` | Organization code |
| `--cloud_provider` | - | string | `aws` | Cloud provider |
| `--account` | `-a` | string | auto | Account type |
| `--start_time` | - | string | - | Start time for historical logs |
| `--end_time` | - | string | - | End time for historical logs |
| `--since` | - | string | - | Duration from now |
| `--linux_operation` | `-l` | string | - | Linux operations for processing |
| `--show_tags` | - | string | - | Comma-separated ddtags to show |
| `--verbose` | `-v` | bool | `false` | Verbose logging |

## 🏢 Maintainers

**Dream11 Engineering Team**

- Repository: [github.com/dream-sports-labs/livelogs](https://github.com/dream11/livelogs)
- Issues: [GitHub Issues](https://github.com/dream-sports-labs/livelogs/issues)

---

**Made with ❤️ by Dream11 Engineering**
