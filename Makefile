all: build

build:
	go build -o main main.go parser.go tester.go

test: build
	./main TEST
