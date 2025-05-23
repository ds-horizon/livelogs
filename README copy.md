# livelogs
 # <img src="./docs/livetail.png" alt="Livelogs" title="Livelogs: Internal CLI to check logs on environments." width="50%" height="50%">

Interface to check logs for services with any environment.
## Installation

```shell
brew install dream11/tools/livelogs
```

> Supporting Darwin(AMD64 & ARM64) and Linux AMD64

## Development Setup

1. [Download](https://golang.org/dl/go1.18.3.darwin-amd64.pkg) and run the Go installer.

2. Verify Go: `go version`

3. Clone this repo: `git clone https://github.com/dream11/livelogs`

4. Enter the repository: `cd livelogs`

5. Install dependencies: `go mod download`

6. Verify the cli: `go run main.go --version`

### Build & Install

1. Run `make install`

Or,

1. Build the executable: `go build .`

2. Move to binary to system path: `sudo mv ./livelogs /usr/local/bin`

3. Verify the cli: `livelogs --version`

## Contribution guide

### Code conventions

1. All variables and functions to be named as per Go's standards (camel case).
   1. Only the variables & functions, that are to be used across packages should be named in exported convention `ExportedName`, rest all names should be in unexported convention `unexportedName`.
   2. All defined command line parameters should follow the following convention - `parameter-name`
      Example: 
      ```go
      logsCmd.Flags().StringP("service", "s", "", "service name")
	  logsCmd.Flags().StringP("env", "e", "", "environment name")
      ```

2. The project must follow the following [layout](https://github.com/golang-standards/project-layout).

### Formatting the code

1. Install Lint tool: `brew install golangci-lint`

2. Upgrade to its latest version: `brew upgrade golangci-lint`

3. Run linter: `make lint`

### Note
> **Note:** To fix gci errors run following commands:
```bash
go install github.com/daixiang0/gci@v0.11.0
gci -w -local github.com/daixiang0/gci main.go
gci write --skip-generated -s standard -s default .
```

> All these linting checks are also ensured in pre-commit checks provided below.

### Making commits

1. Install pre-commit: `pip install pre-commit`

2. Setup pre-commit: `cd livelogs && pre-commit install`

3. Now, make commits.

> Now on every commit that you make, pre-commit hook will validate the `go` code and will suggest changes if any.

### Managing the version

Livelogs application version is in the semantic version form i.e. `x.y.z` where, 
`x` is the major version, `y` is the minor version and `z` is the patch version.

The version is maintained in [livelogs/app/app.go](./app/app.go) inside variable named `App`.

Example: if the current version is `0.0.1`, then in case of 

1. a bug fix or a patch, the patch version is upgraded i.e. `1.0.1`
2. a minor change in some existing feature, the minor version is upgraded i.e. `1.1.0`
3. a major feature addition, the major version is upgraded i.e. `2.0.0`

> Update the version responsibly as this version will be used to create a release against any main branch merge.

## Commands

### Structure

All commands are formatted as: livelogs `<verb>` `<options>`

Here,

1. `verb` - The action to be performed. Supported verbs are -

    - `logs` - For checking logs of service
    - `update` - For updating

2. `options` - Extra properties required to support the commands. Example: `logs` -
    ```shell
    livelogs logs --service=<service_name> --env=<environment_name>
    livelogs logs --service=demo-1234 --env=prod
    livelogs logs --service demo-1234 --env prod
    livelogs logs -s=demo-1234 -e=prod
    livelogs logs -s demo-1234 -e prod
    ```
## Command Line User Interface

To interact with user via Command Line,

```go
import (
    "github.com/dream11/livelogs/internal/ui"
)
```

### Logging

```go
var Logger ui.Logger

func main() {
    Logger.Info("string")
    Logger.Success("string")
    Logger.Warn("string")
    Logger.Output("string")
    Logger.Debug("string")
    Logger.Error("string") // This should be followed by an exit call
}
```
