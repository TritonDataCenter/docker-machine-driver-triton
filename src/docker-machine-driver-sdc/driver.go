package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	stdlog "log"
	"os"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"

	"github.com/joyent/gocommon/client"
	"github.com/joyent/gosdc/cloudapi"
	"github.com/joyent/gosign/auth"
)

const (
	driverName = "triton"
	flagPrefix = driverName + "-"
	envPrefix  = "TRITON_"
)

var (
	defaultSdcAccount = ""
	defaultSdcKeyFile = os.Getenv("HOME") + "/.ssh/id_rsa"
	defaultSdcKeyId   = ""
	defaultSdcUrl     = "https://us-east-1.api.joyent.com"

	// https://docs.joyent.com/public-cloud/instances/virtual-machines/images/linux/debian#debian-8
	defaultSdcImage   = "debian-8"
	defaultSdcPackage = "g3-standard-0.25-kvm"

	errUnimplemented = errors.New("UNIMPLEMENTED")
)

type Driver struct {
	*drivers.BaseDriver

	// authentication/access parameters
	SdcAccount string
	SdcKeyFile string
	SdcKeyId   string
	SdcUrl     string

	// machine creation parameters
	SdcImage   string
	SdcPackage string

	// machine state
	SdcMachineId string
}

// SetConfigFromFlags configures the driver with the object that was returned by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.SdcAccount = opts.String(flagPrefix + "account")
	d.SdcKeyFile = opts.String(flagPrefix + "key-file")
	d.SdcKeyId = opts.String(flagPrefix + "key-id")
	d.SdcUrl = opts.String(flagPrefix + "url")

	d.SdcImage = opts.String(flagPrefix + "image")
	d.SdcPackage = opts.String(flagPrefix + "package")

	d.SetSwarmConfigFromFlags(opts)

	if d.SdcAccount == "" {
		return fmt.Errorf("%s driver requires the --%saccount/%sACCOUNT option", driverName, flagPrefix, envPrefix)
	}
	if d.SdcKeyFile == "" {
		return fmt.Errorf("%s driver requires the --%skey-file/%sKEY_FILE option", driverName, flagPrefix, envPrefix)
	}
	if d.SdcKeyId == "" {
		return fmt.Errorf("%s driver requires the --%skey-id/%sKEY_ID option", driverName, flagPrefix, envPrefix)
	}
	if d.SdcUrl == "" {
		return fmt.Errorf("%s driver requires the --%surl/%sURL option", driverName, flagPrefix, envPrefix)
	}

	if d.SdcImage == "" {
		return fmt.Errorf("%s driver requires the --%simage option", driverName, flagPrefix)
	}
	if d.SdcPackage == "" {
		return fmt.Errorf("%s driver requires the --%spackage option", driverName, flagPrefix)
	}

	return nil
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags that can be set, their descriptions and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: envPrefix + "URL",
			Name:   flagPrefix + "url",
			Usage:  "URL of the CloudAPI endpoint",
			Value:  defaultSdcUrl,
		},
		mcnflag.StringFlag{
			EnvVar: envPrefix + "ACCOUNT",
			Name:   flagPrefix + "account",
			Usage:  "Login name/username",
			Value:  defaultSdcAccount,
		},
		mcnflag.StringFlag{
			EnvVar: envPrefix + "KEY_ID",
			Name:   flagPrefix + "key-id",
			Usage:  fmt.Sprintf(`The fingerprint of $%sKEY_FILE (ssh-keygen -l -E md5 -f $%sKEY_FILE | awk '{ gsub(/^[^:]+:/, "", $2); print $2 }')`, envPrefix, envPrefix),
			Value:  defaultSdcKeyId,
		},
		mcnflag.StringFlag{
			EnvVar: envPrefix + "KEY_FILE",
			Name:   flagPrefix + "key-file",
			Usage:  fmt.Sprintf("An SSH private key file that has been added to $%sACCOUNT", envPrefix),
			Value:  defaultSdcKeyFile,
		},

		mcnflag.StringFlag{
			Name:  flagPrefix + "image",
			Usage: `VM image to provision ("debian-8", "debian-8@20150527", "ca291f66", etc)`,
			Value: defaultSdcImage,
		},
		mcnflag.StringFlag{
			Name:  flagPrefix + "package",
			Usage: `VM instance size to create ("g3-standard-0.25-kvm", "g3-standard-0.5-kvm", etc)`,
			Value: defaultSdcPackage,
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
		stdlog.New(os.Stderr, "", stdlog.LstdFlags),
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

	log.Debugf("machine name: %s", machine.Name)

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

		Image:   d.SdcImage,
		Package: d.SdcPackage,
	})
	if err != nil {
		return err
	}

	d.SdcMachineId = machine.Id

	return nil
}

// https://github.com/joyent/node-triton/blob/aeed6d91922ea117a42eac0cef4a3df67fbfed2f/lib/common.js#L306
func uuidToShortId(s string) string {
	return strings.SplitN(s, "-", 2)[0]
}
func iso8859(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	client, err := d.client()
	if err != nil {
		return err
	}

	// TODO find a better "Ping" method
	_, err = client.CountMachines()
	if err != nil {
		return err
	}

	if _, err := client.GetImage(d.SdcImage); err != nil {
		// apparently isn't a valid ID, but might be a name like "debian-8" (so let's do a lookup)
		// https://github.com/joyent/node-triton/blob/aeed6d91922ea117a42eac0cef4a3df67fbfed2f/lib/tritonapi.js#L368
		nameVersion := strings.SplitN(d.SdcImage, "@", 2)
		name, version := nameVersion[0], ""
		if len(nameVersion) == 2 {
			version = nameVersion[1]
		}
		filter := cloudapi.NewFilter()
		filter.Add("state", "all")
		if version != "" {
			filter.Add("name", name)
			filter.Add("version", version)
		}
		images, imagesErr := client.ListImages(filter)
		if imagesErr != nil {
			return imagesErr
		}
		nameMatches, shortIdMatches := []cloudapi.Image{}, []cloudapi.Image{}
		for _, image := range images {
			if name == image.Name {
				nameMatches = append(nameMatches, image)
			}
			if name == uuidToShortId(image.Id) {
				shortIdMatches = append(shortIdMatches, image)
			}
		}
		if len(nameMatches) == 1 {
			log.Infof("resolved image %q to %q (exact name match)", d.SdcImage, nameMatches[0].Id)
			d.SdcImage = nameMatches[0].Id
		} else if len(nameMatches) > 1 {
			mostRecent := nameMatches[0]
			published, timeErr := iso8859(mostRecent.PublishedAt)
			if timeErr != nil {
				return timeErr
			}
			for _, image := range nameMatches[1:] {
				newPublished, timeErr := iso8859(image.PublishedAt)
				if timeErr != nil {
					return timeErr
				}
				if published.Before(newPublished) {
					mostRecent = image
					published = newPublished
				}
			}
			log.Infof("resolved image %q to %q (most recent of %d name matches)", d.SdcImage, mostRecent.Id, len(nameMatches))
			d.SdcImage = mostRecent.Id
		} else if len(shortIdMatches) == 1 {
			log.Infof("resolved image %q to %q (exact short id match)", d.SdcImage, shortIdMatches[0].Id)
			d.SdcImage = shortIdMatches[0].Id
		} else {
			if len(shortIdMatches) > 1 {
				log.Warnf("image %q is an ambiguous short id", d.SdcImage)
			}
			return err
		}
	}

	// GetPackage (and CreateMachine) both support package names and UUIDs interchangeably
	if _, err := client.GetPackage(d.SdcPackage); err != nil {
		return err
	}

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
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s:%d", ip, engine.DefaultPort), nil
}

func (d *Driver) GetSSHKeyPath() string {
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
	case "configured", "provisioning":
		return state.Starting, nil
	case "failed", "receiving":
		return state.Error, nil
	case "running":
		return state.Running, nil
	case "shutting_down", "stopping":
		return state.Stopping, nil
	case "down", "stopped":
		return state.Stopped, nil
	}

	return state.Error, fmt.Errorf("unknown SDC machine state: %s", machine.State)
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	// TODO find something more forceful than "Stop"
	return d.Stop()
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
	client, err := d.client()
	if err != nil {
		return err
	}
	return client.StopMachine(d.SdcMachineId)
}
