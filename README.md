# docker-machine-driver-triton
A Docker Machine driver for Triton.

## Requirements
* [Docker](https://www.docker.com/products/overview#/install_the_platform)
* [Docker Machine](https://docs.docker.com/machine/install-machine)
* [Go](https://golang.org/doc/install)

You need a Triton account to use this driver. See [this page](https://www.joyent.com/) to create an account on the Triton Public Cloud.

## Installation from source
To get the code and compile the binary, run:
```bash
go get -u github.com/joyent/docker-machine-driver-triton
```

Then put the driver in a directory filled in your PATH environment variable or run:
```bash
export PATH=$PATH:$GOPATH/bin
```
This will allow the docker-machine command to find the docker-machine-driver-triton binary.

## How to use

### Driver-specific command line flags

#### Flags description
* **`--triton-account` : The username of the Triton account to use when using the Triton Cloud API. (required)**
* **`--triton-key-id` : The fingerprint of the public key of the SSH key pair to use for authentication with the Triton Cloud API. (required)**
* `--triton-key-path` : Path to the file in which the private key of triton_key_id is stored.
* `--triton-url` : The URL of the Triton Cloud API to use.
* `--triton-image` : The name of the Triton image to use.
* `--triton-package` : The Triton package to use.
* `--triton-ssh-user`: The username to connect to SSH with.

#### Flags usage
|             Option             |          Environment         |            Default value            |
|--------------------------------|------------------------------|-------------------------------------|
| `--triton-account`             | `TRITON_ACCOUNT`             |                                     |
| `--triton-key-id`              | `TRITON_KEY_ID`              |                                     |
| `--triton-key-path`            | `TRITON_KEY_PATH`            | "~/.ssh/id_rsa"                     |
| `--triton-url`                 | `TRITON_URL`                 | "https://us-east-1.api.joyent.com"  |
| `--triton-image`               |                              | "debian-8"                          |
| `--triton-package`             |                              | "g3-standard-0.25-kvm"              |
| `--triton-ssh-user`            | `TRITON_SSH_USER`            | "root"                              |

### Provisioning examples
An example:
```bash
docker-machine create -d triton \
--triton-account nima@jalali.net \
--triton-key-id 68:9f:9a:c4:76:3a:f4:62:77:47:3e:47:d4:34:4a:b7 \
test-node
```

An example using environment variables:
```bash
export TRITON_ACCOUNT="nima@jalali.net"
export TRITON_KEY_ID="68:9f:9a:c4:76:3a:f4:62:77:47:3e:47:d4:34:4a:b7"
docker-machine create -d triton test-node
```

An example using a Ubuntu Image:
```bash
docker-machine create -d triton \
--triton-account nima@jalali.net \
--triton-key-id 68:9f:9a:c4:76:3a:f4:62:77:47:3e:47:d4:34:4a:b7 \
--triton-image ubuntu-certified-16.10@20170619.1 \
--triton-ssh-user ubuntu \
test-node
```
