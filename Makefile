tidy:
	go mod tidy

build:
	@GOOS='windows' GOARCH='amd64' go build -o build/servicewrapper.exe servicewrapper.go

clean:
	@rm -rf build

.PHONY: tidy build clean