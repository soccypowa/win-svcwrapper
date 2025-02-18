tidy:
	go mod tidy

build:
	go build -o servicewrapper.exe servicewrapper.go

clean:
	del .\servicewrapper.exe

.PHONY: tidy build clean