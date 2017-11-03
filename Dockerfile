FROM golang:1.9-alpine

RUN apk add --no-cache gcc git libc-dev

ENV DOCKER_MACHINE_VERSION v0.12.2
ENV DOCKER_MACHINE_PKG github.com/docker/machine
RUN git clone --depth 1 -b "$DOCKER_MACHINE_VERSION" https://$DOCKER_MACHINE_PKG.git "$GOPATH/src/$DOCKER_MACHINE_PKG" \
    && go build -v -o "$GOPATH/bin/docker-machine" $DOCKER_MACHINE_PKG/cmd

WORKDIR $GOPATH/src/github.com/joyent/docker-machine-driver-triton

COPY *.go ./
COPY ./vendor/ ./vendor/
RUN go install -v ./...

CMD ["sh"]
