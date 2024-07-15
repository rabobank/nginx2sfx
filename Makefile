BINARY=nginx2sfx

COMMIT_HASH := `git rev-parse --short=8 HEAD 2>/dev/null`
BUILD_TIME := `date +%FT%T%z`
LDFLAGS=-s -w -X github.com/rabobank/${BINARY}/conf.CommitHash=${COMMIT_HASH} -X github.com/rabobank/${BINARY}/conf.BuildTime=${BUILD_TIME}

all: deps linux

clean:
	go clean
	if [ -f ./target/linux_amd64/${BINARY} ] ; then rm ./target/linux_amd64/${BINARY} ; fi

deps:
	go get -v ./...

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-d ${LDFLAGS}" -o ./target/linux_amd64/${BINARY} .
