build:
	go build -v -o gowalker

web: build
	./gowalker

release:
	env GOOS=linux GOARCH=amd64 go build -o gowalker
