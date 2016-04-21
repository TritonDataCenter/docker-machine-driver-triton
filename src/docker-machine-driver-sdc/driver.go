package main

import (
	"errors"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	_ "github.com/joyent/gosdc"
)

var (
	errUnimplemented = errors.New("UNIMPLEMENTED")
)

type Driver struct {
	*drivers.BaseDriver
}

func NewDriver(hostName, storePath string) Driver {
	return Driver{
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
	return "sdc"
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags that can be set, their descriptions and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return nil
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

// SetConfigFromFlags configures the driver with the object that was returned by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
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
