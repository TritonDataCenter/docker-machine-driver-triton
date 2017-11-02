FROM golang:1.9-alpine

RUN apk add --no-cache gcc git libc-dev curl

ENV DOCKER_MACHINE_VERSION v0.12.2
RUN git clone --depth 1 -b "$DOCKER_MACHINE_VERSION" https://github.com/docker/machine.git "$GOPATH/src/github.com/docker/machine" \
    && go build -v -o "$GOPATH/bin/docker-machine" github.com/docker/machine/cmd

ENV DEP_VERSION 0.3.2
ENV DEP_CHECKSUM 322152b8b50b26e5e3a7f6ebaeb75d9c11a747e64bbfd0d8bb1f4d89a031c2b5
RUN curl -Lso /tmp/dep "https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64" \
    && echo "${DEP_CHECKSUM}  /tmp/dep" | sha256sum -c \
    && mv /tmp/dep /usr/local/bin/dep \
    && chmod 0755 /usr/local/bin/dep

WORKDIR $GOPATH/src/github.com/joyent/docker-machine-driver-triton
COPY *.go ./
COPY Gopkg.* ./
RUN dep ensure \
    && go install -v ./...

CMD ["sh"]
