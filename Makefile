PROJECTNAME=$(shell basename "$(PWD)")
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

install:
	go mod download

dev:
	go build -o $(GOBIN)/app ./main.go && $(GOBIN)/app