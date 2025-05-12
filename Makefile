install:
	go mod download
	go build .
	sudo mv ./livelogs /usr/local/bin
	livelogs --version

lint:
	golangci-lint run -E gofmt -E gci --fix;

build:
	go mod download
	mkdir -p bin/livelogs_darwin_amd64
	env GOOS=darwin GOARCH=amd64 go build -o bin/livelogs_darwin_amd64/livelogs
	mkdir -p bin/livelogs_darwin_arm64
	env GOOS=darwin GOARCH=arm64 go build -o bin/livelogs_darwin_arm64/livelogs
	mkdir -p bin/livelogs_linux_amd64
	env GOOS=linux GOARCH=amd64 go build -o bin/livelogs_linux_amd64/livelogs

compressed-builds: build
	cd bin/livelogs_darwin_amd64 && tar -czvf ../livelogs_darwin_amd64.tar.gz livelogs
	cd bin/livelogs_darwin_arm64 && tar -czvf ../livelogs_darwin_arm64.tar.gz livelogs
	cd bin/livelogs_linux_amd64 && tar -czvf ../livelogs_linux_amd64.tar.gz livelogs
