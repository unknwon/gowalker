build:
	go install -v
	cp '$(GOPATH)/bin/gowalker' .

web: build
	./gowalker
