build:
	go install -v
	cp '$(GOPATH)/bin/gowalker' .

web: build
	./gowalker

release:
	env GOOS=linux GOARCH=amd64 go build -o gowalker
