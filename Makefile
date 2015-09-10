all:
	go build -o hook-to-rest hook-to-rest.go

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o hook-to-rest hook-to-rest.go

clean:
	go clean
	rm -rf hook-to-test
