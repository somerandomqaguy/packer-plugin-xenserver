//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package common

import (
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	CommonConfig        `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	// The maximum number of VCPUs for the VM.
	// By default this is 1.
	VCPUsMax       uint              `mapstructure:"vcpus_max"`

	// The number of startup VCPUs for the VM.
	// By default this is 1.
	VCPUsAtStartup uint              `mapstructure:"vcpus_atstartup"`

	// The size, in megabytes, of the amount of memory to
	// allocate for the VM. By default, this is 1024 (1 GB).
	VMMemory       uint              `mapstructure:"vm_memory"`
	// The size, in megabytes, of the hard disk to create
	// for the VM. By default, this is 40000 (about 40 GB).
	DiskSize       uint              `mapstructure:"disk_size"`

	CloneTemplate  string            `mapstructure:"clone_template"`
	// The platform args.
	// Defaults to
	//   ```javascript
	//   {
	// 	  "viridian": "false",
	// 	  "nx": "true",
	// 	  "pae": "true",
	// 	  "apic": "true",
	// 	  "timeoffset": "0",
	// 	  "acpi": "1",
	// 	  "cores-per-socket": "1"
	//   }
	//   ```
	VMOtherConfig  map[string]string `mapstructure:"vm_other_config"`


	ISOChecksum     string   `mapstructure:"iso_checksum"` // FIXME: Change these to ISOConfig struct from common
	ISOUrls         []string `mapstructure:"iso_urls"`
	ISOUrl          string   `mapstructure:"iso_url"`
	ISOName         string   `mapstructure:"iso_name"`

	PlatformArgs map[string]string `mapstructure:"platform_args"`

	RawInstallTimeout string        `mapstructure:"install_timeout"`
	InstallTimeout    time.Duration ``
	SourcePath        string        `mapstructure:"source_path"`

	Firmware string `mapstructure:"firmware"`

	ctx interpolate.Context
}

func (c Config) GetInterpContext() *interpolate.Context {
	return &c.ctx
}
