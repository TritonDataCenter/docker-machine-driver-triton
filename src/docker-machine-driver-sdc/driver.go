package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"

	"github.com/joyent/gocommon/client"
	"github.com/joyent/gosdc/cloudapi"
	"github.com/joyent/gosign/auth"
)

const (
	driverName = "sdc"

	defaultSdcAccount = ""
	defaultSdcKeyFile = ""
	defaultSdcKeyId   = ""
	defaultSdcUrl     = "https://us-east-1.api.joyent.com"
)

var (
	errUnimplemented = errors.New("UNIMPLEMENTED")
)

type Driver struct {
	*drivers.BaseDriver

	SdcAccount string
	SdcKeyFile string
	SdcKeyId   string
	SdcUrl     string
}

// SetConfigFromFlags configures the driver with the object that was returned by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.SdcAccount = opts.String(driverName + "-account")
	d.SdcKeyFile = opts.String(driverName + "-key-file")
	d.SdcKeyId = opts.String(driverName + "-key-id")
	d.SdcUrl = opts.String(driverName + "-url")
	d.SetSwarmConfigFromFlags(opts)

	if d.SdcAccount == "" {
		return fmt.Errorf("%s driver requires the %s-account option", driverName, driverName)
	}
	if d.SdcKeyFile == "" {
		return fmt.Errorf("%s driver requires the %s-key-file option", driverName, driverName)
	}
	if d.SdcKeyId == "" {
		return fmt.Errorf("%s driver requires the %s-key-id option", driverName, driverName)
	}
	if d.SdcUrl == "" {
		return fmt.Errorf("%s driver requires the %s-url option", driverName, driverName)
	}

	return nil
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags that can be set, their descriptions and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "SDC_URL",
			Name:   driverName + "-url",
			Usage:  "URL of the CloudAPI endpoint",
			Value:  defaultSdcUrl,
		},
		mcnflag.StringFlag{
			EnvVar: "SDC_ACCOUNT",
			Name:   driverName + "-account",
			Usage:  "Login name/username",
			Value:  defaultSdcAccount,
		},
		mcnflag.StringFlag{
			EnvVar: "SDC_KEY_ID",
			Name:   driverName + "-key-id",
			Usage:  "The fingerprint of $SDC_KEY_FILE (ssh-keygen -l -f $SDC_KEY_FILE | awk '{ print $2 }')",
			Value:  defaultSdcKeyId,
		},
		mcnflag.StringFlag{
			EnvVar: "SDC_KEY_FILE",
			Name:   driverName + "-key-file",
			Usage:  "An SSH public key file that has been added to $SDC_ACCOUNT",
			Value:  defaultSdcKeyFile,
		},
	}
}

func (d Driver) client() (*cloudapi.Client, error) {
	keyData, err := ioutil.ReadFile(d.SdcKeyFile)
	if err != nil {
		return nil, err
	}
	userAuth, err := auth.NewAuth(d.SdcAccount, string(keyData), "rsa-sha256")
	if err != nil {
		return nil, err
	}

	creds := &auth.Credentials{
		UserAuthentication: userAuth,
		SdcKeyId:           d.SdcKeyId,
		SdcEndpoint:        auth.Endpoint{URL: d.SdcUrl},
	}

	return cloudapi.New(client.NewClient(
		creds.SdcEndpoint.URL,
		cloudapi.DefaultAPIVersion,
		creds,
		log.New(os.Stderr, "", log.LstdFlags),
	)), nil
}

func NewDriver(hostName, storePath string) Driver {
	return Driver{
		SdcAccount: defaultSdcAccount,
		SdcKeyFile: defaultSdcKeyFile,
		SdcKeyId:   defaultSdcKeyId,
		SdcUrl:     defaultSdcUrl,

		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// https://github.com/docker/machine/blob/v0.7.0/libmachine/drivers/drivers.go
// https://github.com/docker/machine/blob/v0.7.0/libmachine/drivers/base.go

// Create a host using the driver's config
func (d *Driver) Create() error {
	return errUnimplemented
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

// GetSSHUsername returns username for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return "", errUnimplemented
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	return "", errUnimplemented
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	// https://github.com/docker/machine/blob/v0.7.0/libmachine/state/state.go
	return state.None, errUnimplemented
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return errUnimplemented
}

// Remove a host
func (d *Driver) Remove() error {
	return errUnimplemented
}

// Restart a host. This may just call Stop(); Start() if the provider does not have any special restart behaviour.
func (d *Driver) Restart() error {
	return errUnimplemented
}

// Start a host
func (d *Driver) Start() error {
	return errUnimplemented
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	return errUnimplemented
}
