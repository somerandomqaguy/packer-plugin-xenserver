//go:generate packer-sdc struct-markdown

package common

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	xenapi "github.com/terra-farm/go-xen-api-client"
)

type CommonConfig struct {
	// The XenServer username used to access the remote machine.
	Username string `mapstructure:"remote_username" required:"true"`
	// The XenServer password for access to the remote machine.
	Password string `mapstructure:"remote_password" required:"true"`
	// (string) - The host of the Xenserver / XCP-ng pool primary.
	// Typically these will be specified through environment variables as seen in the
	// [examples](../../examples/centos8.json).
	HostIp   string `mapstructure:"remote_host" required:"true"`
	// This is the name of the new virtual
	// machine, without the file extension. By default this is
	// "packer-BUILDNAME-TIMESTAMP", where "BUILDNAME" is the name of the build.
	VMName             string   `mapstructure:"vm_name"`
	// The description of the new virtual
	// machine. By default this is the empty string.
	VMDescription      string   `mapstructure:"vm_description"`
	// The name of the SR to packer will save the built VM to
	SrName             string   `mapstructure:"sr_name" required:"true"`
	// The ISO-SR to packer will use. The ISO will be uploaded to the SR.
	SrISOName          string   `mapstructure:"sr_iso_name" required:"true"`

	FloppyFiles        []string `mapstructure:"floppy_files"`
	NetworkNames       []string `mapstructure:"network_names"`
	ExportNetworkNames []string `mapstructure:"export_network_names"`

	HostPortMin uint `mapstructure:"host_port_min"`
	HostPortMax uint `mapstructure:"host_port_max"`

	BootCommand     []string `mapstructure:"boot_command"`
	ShutdownCommand string   `mapstructure:"shutdown_command"`

	RawBootWait string `mapstructure:"boot_wait"`
	BootWait    time.Duration

	// The name of the XenServer Tools ISO. Defaults to "xs-tools.iso".
	ToolsIsoName string `mapstructure:"tools_iso_name"`

	HTTPDir     string `mapstructure:"http_directory"`
	HTTPPortMin uint   `mapstructure:"http_port_min"`
	HTTPPortMax uint   `mapstructure:"http_port_max"`

	//	SSHHostPortMin    uint   `mapstructure:"ssh_host_port_min"`
	//	SSHHostPortMax    uint   `mapstructure:"ssh_host_port_max"`
	SSHKeyPath  string `mapstructure:"ssh_key_path"`
	SSHPassword string `mapstructure:"ssh_password"`
	SSHPort     uint   `mapstructure:"ssh_port"`
	SSHUser     string `mapstructure:"ssh_username"`
	SSHConfig   `mapstructure:",squash"`

	RawSSHWaitTimeout string `mapstructure:"ssh_wait_timeout"`
	SSHWaitTimeout    time.Duration

	OutputDir string `mapstructure:"output_directory"`
	Format    string `mapstructure:"format"`
	// Determine when to keep the VM and when to clean it up. This
	// can be "always", "never" or "on_success". By default this is "never", and Packer
	// always deletes the VM regardless of whether the process succeeded and an artifact
	// was produced. "always" asks Packer to leave the VM at the end of the process
	// regardless of success. "on_success" requests that the VM only be cleaned up if an
	// artifact was produced. The latter is useful for debugging templates that fail.
	KeepVM    string `mapstructure:"keep_vm"`


	IPGetter  string `mapstructure:"ip_getter"`
}

func (c *CommonConfig) Prepare(ctx *interpolate.Context, pc *common.PackerConfig) []error {
	var err error
	var errs []error

	// Set default values

	if c.HostPortMin == 0 {
		c.HostPortMin = 5900
	}

	if c.HostPortMax == 0 {
		c.HostPortMax = 6000
	}

	if c.RawBootWait == "" {
		c.RawBootWait = "5s"
	}

	if c.ToolsIsoName == "" {
		c.ToolsIsoName = "xs-tools.iso"
	}

	if c.HTTPPortMin == 0 {
		c.HTTPPortMin = 8000
	}

	if c.HTTPPortMax == 0 {
		c.HTTPPortMax = 9000
	}

	if c.RawSSHWaitTimeout == "" {
		c.RawSSHWaitTimeout = "200m"
	}

	if c.FloppyFiles == nil {
		c.FloppyFiles = make([]string, 0)
	}

	/*
		if c.SSHHostPortMin == 0 {
			c.SSHHostPortMin = 2222
		}

		if c.SSHHostPortMax == 0 {
			c.SSHHostPortMax = 4444
		}
	*/

	if c.SSHPort == 0 {
		c.SSHPort = 22
	}

	if c.RawSSHWaitTimeout == "" {
		c.RawSSHWaitTimeout = "20m"
	}

	if c.OutputDir == "" {
		c.OutputDir = fmt.Sprintf("output-%s", pc.PackerBuildName)
	}

	if c.VMName == "" {
		c.VMName = fmt.Sprintf("packer-%s-{{timestamp}}", pc.PackerBuildName)
	}

	if c.Format == "" {
		c.Format = "xva"
	}

	if c.KeepVM == "" {
		c.KeepVM = "never"
	}

	if c.IPGetter == "" {
		c.IPGetter = "auto"
	}

	// Validation

	if c.Username == "" {
		errs = append(errs, errors.New("remote_username must be specified."))
	}

	if c.Password == "" {
		errs = append(errs, errors.New("remote_password must be specified."))
	}

	if c.HostIp == "" {
		errs = append(errs, errors.New("remote_host must be specified."))
	}

	if c.HostPortMin > c.HostPortMax {
		errs = append(errs, errors.New("the host min port must be less than the max"))
	}

	if c.HTTPPortMin > c.HTTPPortMax {
		errs = append(errs, errors.New("the HTTP min port must be less than the max"))
	}

	c.BootWait, err = time.ParseDuration(c.RawBootWait)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed to parse boot_wait: %s", err))
	}

	if c.SSHKeyPath != "" {
		if _, err := os.Stat(c.SSHKeyPath); err != nil {
			errs = append(errs, fmt.Errorf("ssh_key_path is invalid: %s", err))
		} else if _, err := FileSigner(c.SSHKeyPath); err != nil {
			errs = append(errs, fmt.Errorf("ssh_key_path is invalid: %s", err))
		}
	}

	/*
		if c.SSHHostPortMin > c.SSHHostPortMax {
			errs = append(errs,
				errors.New("ssh_host_port_min must be less than ssh_host_port_max"))
		}
	*/

	if c.SSHUser == "" {
		errs = append(errs, errors.New("An ssh_username must be specified."))
	}

	c.SSHWaitTimeout, err = time.ParseDuration(c.RawSSHWaitTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed to parse ssh_wait_timeout: %s", err))
	}

	switch c.Format {
	case "xva", "xva_compressed", "vdi_raw", "vdi_vhd", "none":
	default:
		errs = append(errs, errors.New("format must be one of 'xva', 'vdi_raw', 'vdi_vhd', 'none'"))
	}

	switch c.KeepVM {
	case "always", "never", "on_success":
	default:
		errs = append(errs, errors.New("keep_vm must be one of 'always', 'never', 'on_success'"))
	}

	switch c.IPGetter {
	case "auto", "tools", "http":
	default:
		errs = append(errs, errors.New("ip_getter must be one of 'auto', 'tools', 'http'"))
	}

	return errs
}

// steps should check config.ShouldKeepVM first before cleaning up the VM
func (c CommonConfig) ShouldKeepVM(state multistep.StateBag) bool {
	switch c.KeepVM {
	case "always":
		return true
	case "never":
		return false
	case "on_success":
		// only keep instance if build was successful
		_, cancelled := state.GetOk(multistep.StateCancelled)
		_, halted := state.GetOk(multistep.StateHalted)
		return !(cancelled || halted)
	default:
		panic(fmt.Sprintf("Unknown keep_vm value '%s'", c.KeepVM))
	}
}

func (config CommonConfig) GetSR(c *Connection) (xenapi.SRRef, error) {
	var srRef xenapi.SRRef
	if config.SrName == "" {
		hostRef, err := c.GetClient().Session.GetThisHost(c.session, c.session)

		if err != nil {
			return srRef, err
		}

		pools, err := c.GetClient().Pool.GetAllRecords(c.session)

		if err != nil {
			return srRef, err
		}

		for _, pool := range pools {
			if pool.Master == hostRef {
				return pool.DefaultSR, nil
			}
		}

		return srRef, errors.New(fmt.Sprintf("failed to find default SR on host '%s'", hostRef))

	} else {
		// Use the provided name label to find the SR to use
		srs, err := c.GetClient().SR.GetByNameLabel(c.session, config.SrName)

		if err != nil {
			return srRef, err
		}

		switch {
		case len(srs) == 0:
			return srRef, fmt.Errorf("Couldn't find a SR with the specified name-label '%s'", config.SrName)
		case len(srs) > 1:
			return srRef, fmt.Errorf("Found more than one SR with the name '%s'. The name must be unique", config.SrName)
		}

		return srs[0], nil
	}
}

func (config CommonConfig) GetISOSR(c *Connection) (xenapi.SRRef, error) {
	var srRef xenapi.SRRef
	if config.SrISOName == "" {
		return srRef, errors.New("sr_iso_name must be specified in the packer configuration")

	} else {
		// Use the provided name label to find the SR to use
		srs, err := c.GetClient().SR.GetByNameLabel(c.session, config.SrName)

		if err != nil {
			return srRef, err
		}

		switch {
		case len(srs) == 0:
			return srRef, fmt.Errorf("Couldn't find a SR with the specified name-label '%s'", config.SrName)
		case len(srs) > 1:
			return srRef, fmt.Errorf("Found more than one SR with the name '%s'. The name must be unique", config.SrName)
		}

		return srs[0], nil
	}
}
