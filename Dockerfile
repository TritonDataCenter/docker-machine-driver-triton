FROM golang:1.6-alpine

RUN apk add --no-cache gcc git libc-dev

ENV DOCKER_MACHINE_VERSION v0.7.0

RUN git clone --depth 1 -b "$DOCKER_MACHINE_VERSION" https://github.com/docker/machine.git "$GOPATH/src/github.com/docker/machine"
RUN go build -v -o "$GOPATH/bin/docker-machine" github.com/docker/machine/cmd

RUN go get -d -v github.com/joyent/gosdc/...

WORKDIR $GOPATH/src/github.com/joyent/docker-machine-driver-triton
COPY *.go ./

RUN go install -v ./...

CMD ["sh"]
