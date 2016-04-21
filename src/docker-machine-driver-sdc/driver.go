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
)

var (
	defaultSdcAccount = ""
	defaultSdcKeyFile = os.Getenv("HOME") + "/.ssh/id_rsa"
	defaultSdcKeyId   = ""
	defaultSdcUrl     = "https://us-east-1.api.joyent.com"

	// "debian-8/20150702"
	// https://docs.joyent.com/public-cloud/instances/virtual-machines/images/linux/debian#debian-8-20150702
	defaultSdcImage = "2f56d126-20d0-11e5-9e5b-5f3ef6688aba"

	// "g3-standard-0.25-kvm"
	defaultSdcPackage = "f13b10ca-1a63-4903-b29e-8615ea3858a6"

	errUnimplemented = errors.New("UNIMPLEMENTED")
)

type Driver struct {
	*drivers.BaseDriver

	// authentication/access parameters
	SdcAccount string
	SdcKeyFile string
	SdcKeyId   string
	SdcUrl     string

	// machine state
	SdcMachineId string
}

// SetConfigFromFlags configures the driver with the object that was returned by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.SdcAccount = opts.String(driverName + "-account")
	d.SdcKeyFile = opts.String(driverName + "-key-file")
	d.SdcKeyId = opts.String(driverName + "-key-id")
	d.SdcUrl = opts.String(driverName + "-url")
	d.SetSwarmConfigFromFlags(opts)

	if d.SdcAccount == "" {
		return fmt.Errorf("%s driver requires the --%s-account/SDC_ACCOUNT option", driverName, driverName)
	}
	if d.SdcKeyFile == "" {
		return fmt.Errorf("%s driver requires the --%s-key-file/SDC_KEY_FILE option", driverName, driverName)
	}
	if d.SdcKeyId == "" {
		return fmt.Errorf("%s driver requires the --%s-key-id/SDC_KEY_ID option", driverName, driverName)
	}
	if d.SdcUrl == "" {
		return fmt.Errorf("%s driver requires the --%s-url/SDC_URL option", driverName, driverName)
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
			Usage:  `The fingerprint of $SDC_KEY_FILE (ssh-keygen -l -E md5 -f $SDC_KEY_FILE | awk '{ gsub(/^[^:]+:/, "", $2); print $2 }')`,
			Value:  defaultSdcKeyId,
		},
		mcnflag.StringFlag{
			EnvVar: "SDC_KEY_FILE",
			Name:   driverName + "-key-file",
			Usage:  "An SSH private key file that has been added to $SDC_ACCOUNT",
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
func (d *Driver) getMachine() (*cloudapi.Machine, error) {
	client, err := d.client()
	if err != nil {
		return nil, err
	}
	machine, err := client.GetMachine(d.SdcMachineId)
	if err != nil {
		return nil, err
	}

	// update d.IPAddress since we know the value (saves later work)
	d.IPAddress = machine.PrimaryIP

	return machine, nil
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
	client, err := d.client()
	if err != nil {
		return err
	}

	machine, err := client.CreateMachine(cloudapi.CreateMachineOpts{
		Name: d.MachineName,

		// TODO configurable in some way
		Image:   defaultSdcImage,
		Package: defaultSdcPackage,
	})
	if err != nil {
		return err
	}

	d.SdcMachineId = machine.Id

	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
func (d *Driver) GetIP() (string, error) {
	if d.IPAddress != "" {
		return d.IPAddress, nil
	}
	machine, err := d.getMachine()
	if err != nil {
		return "", err
	}
	return machine.PrimaryIP, nil
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetSSHKeyPath() string {
	// TODO remove this and somehow install the key provided by docker-machine into the VM
	return d.SdcKeyFile
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	// https://github.com/docker/machine/blob/v0.7.0/libmachine/state/state.go

	machine, err := d.getMachine()
	if err != nil {
		return state.Error, err
	}

	// https://github.com/joyent/smartos-live/blob/master/src/vm/man/vmadm.1m.md#vm-states
	switch machine.State {
	case "configured":
		return state.Starting, nil
	case "installed":
		return state.Starting, nil
	case "ready":
		return state.Starting, nil
	case "running":
		return state.Running, nil
	case "shutting_down":
		return state.Stopping, nil
	case "down":
		return state.Stopped, nil
	}

	return state.Error, fmt.Errorf("unknown SDC machine state: %s", machine.State)
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	client, err := d.client()
	if err != nil {
		return err
	}
	return client.StopMachine(d.SdcMachineId)
}

// Remove a host
func (d *Driver) Remove() error {
	client, err := d.client()
	if err != nil {
		return err
	}
	return client.DeleteMachine(d.SdcMachineId)
}

// Restart a host. This may just call Stop(); Start() if the provider does not have any special restart behaviour.
func (d *Driver) Restart() error {
	client, err := d.client()
	if err != nil {
		return err
	}
	return client.RebootMachine(d.SdcMachineId)
}

// Start a host
func (d *Driver) Start() error {
	client, err := d.client()
	if err != nil {
		return err
	}
	return client.StartMachine(d.SdcMachineId)
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	// TODO
	return errUnimplemented
}
