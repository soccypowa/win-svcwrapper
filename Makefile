tidy:
	@go mod tidy

build:
	@GOOS='windows' GOARCH='amd64' go build -o servicewrapper.exe servicewrapper.go

clean:
	@rm -rf servicewrapper.exe

.PHONY: tidy build clean