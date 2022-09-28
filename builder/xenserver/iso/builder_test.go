package iso

import (
	"reflect"
	"testing"
	"strings"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/common"
)


func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"remote_host":       "localhost",
		"remote_username":   "admin",
		"remote_password":   "admin",
		"vm_name":           "foo",
		"iso_checksum":      "md5:A221725EE181A44C67E25BD6A2516742",
		"iso_url":           "http://www.google.com/",
		"shutdown_command":  "yes",
		"ssh_username":      "foo",

		common.BuildNameConfigKey: "foo",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Error("Builder must implement builder.")
	}
}

func TestBuilderPrepare_Defaults(t *testing.T) {
	var b Builder
	config := testConfig()
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.ToolsIsoName != "xs-tools.iso" {
		t.Errorf("bad tools ISO name: %s", b.config.ToolsIsoName)
	}

	if b.config.CloneTemplate != "Other install media" {
		t.Errorf("bad clone template: %s", b.config.CloneTemplate)
	}

	if b.config.VMName == "" {
		t.Errorf("bad vm name: %s", b.config.VMName)
	}

	if b.config.Format != "xva" {
		t.Errorf("bad format: %s", b.config.Format)
	}

	if b.config.KeepVM != "never" {
		t.Errorf("bad keep instance: %s", b.config.KeepVM)
	}
}

func TestBuilderPrepare_DiskSize(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "disk_size")
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	if b.config.DiskSize != 40000 {
		t.Fatalf("bad size: %d", b.config.DiskSize)
	}

	config["disk_size"] = 60000
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.DiskSize != 60000 {
		t.Fatalf("bad size: %d", b.config.DiskSize)
	}
}

func TestBuilderPrepare_Format(t *testing.T) {
	var b Builder
	config := testConfig()

	// Bad
	config["format"] = "foo"
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["format"] = "vdi_raw"
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_HTTPPort(t *testing.T) {
	var b Builder
	config := testConfig()

	// Bad
	config["http_port_min"] = 1000
	config["http_port_max"] = 500
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Bad
	config["http_port_min"] = -500
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["http_port_min"] = 500
	config["http_port_max"] = 1000
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_ISOChecksum(t *testing.T) {
	var b Builder
	config := testConfig()
	var warns []string
	var err error

	// Test bad with empty iso_checksum string
	config["iso_checksum"] = ""
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatalf("should have checksum error with an empty string: %s", err)
	}

	// Test bad
	bad_checksums := []string {
		"md5:A221725EE181A44C6742BAD",
		"A221725EE181A44C6742BAD",
		"Z221725EE181A44C67E25BD6A2516BAD",
	}

	for _, bad_checksum := range bad_checksums {
		config["iso_checksum"] = bad_checksum
		b = Builder{}
		_, warns, err = b.Prepare(config)
		if len(warns) > 0 {
			t.Fatalf("%s bad: %#v", bad_checksum, warns)
		}
		if err == nil {
			t.Fatalf("%s should have checksum error: %s", bad_checksum, err)
		}
		config["iso_checksum"] = "BLAH"
	}


	// Test good
	good_checksums := []string {
		"sha512:1F0E0CE0036C7EAACA84ECB41A93F352029B3BAFDF83E9E469E5E26980075231C553ABA90E5687E36F63F05915C317D8FA4BE33BBC505112BA64FFD754D382A1",
		"1F0E0CE0036C7EAACA84ECB41A93F352029B3BAFDF83E9E469E5E26980075231C553ABA90E5687E36F63F05915C317D8FA4BE33BBC505112BA64FFD754D382A1",
		"sha256:BA4F78A4C2E928D49829AABFBF204305D6D24C7F189DD071CDE25A4D490F1219",
		"BA4F78A4C2E928D49829AABFBF204305D6D24C7F189DD071CDE25A4D490F1219",
		"sha1:69F180CA9D93DAE6670360F38D0E7D6228993F7E",
		"69F180CA9D93DAE6670360F38D0E7D6228993F7E",
		"md5:A221725EE181A44C67E25BD6A2516742",
		"A221725EE181A44C67E25BD6A2516742",
		"none",
	}

	for _, good_checksum := range good_checksums {
		config["iso_checksum"] = good_checksum
		b = Builder{}
		_, warns, err = b.Prepare(config)
		if len(warns) > 0 {
			t.Fatalf("%s bad: %#v", good_checksum, warns)
		}
		if err != nil {
			t.Fatalf("%s should not have checksum error: %s", good_checksum, err)
		}
		config["iso_checksum"] = "BLAH"

		//make sure lower case works too
		config["iso_checksum"] = strings.ToLower(good_checksum)
		b = Builder{}
		_, warns, err = b.Prepare(config)
		if len(warns) > 0 {
			t.Fatalf("%s bad: %#v", good_checksum, warns)
		}
		if err != nil {
			t.Fatalf("%s should not have checksum error: %s", good_checksum, err)
		}
		config["iso_checksum"] = "BLAH"
	}
}

func TestBuilderPrepare_ISOChecksumTypeDeprecation(t *testing.T) {
	var b Builder
	config := testConfig()

	config["iso_checksum_type"] = "md5"
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}

	if err == nil {
		t.Fatalf("should have error")
	}

	if ! strings.Contains(err.Error(), "Deprecated configuration key: 'iso_checksum_type'.") {
		t.Fatalf("No deprecration error found: %s", err)
	}
}

func TestBuilderPrepare_ISOUrl(t *testing.T) {
	var b Builder
	config := testConfig()
	delete(config, "iso_url")
	delete(config, "iso_urls")

	// Test both epty
	config["iso_url"] = ""
	b = Builder{}
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test iso_url set
	config["iso_url"] = "http://www.packer.io"
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}

	expected := []string{"http://www.packer.io"}

	// Test both set
	config["iso_url"] = "http://www.packer.io"
	config["iso_urls"] = []string{"http://www.packer.io"}
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test just iso_urls set
	delete(config, "iso_url")
	config["iso_urls"] = []string{
		"http://www.packer.io",
		"http://www.hashicorp.com",
	}

	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}

	expected = []string{
		"http://www.packer.io",
		"http://www.hashicorp.com",
	}
	if !reflect.DeepEqual(b.config.ISOUrls, expected) {
		t.Fatalf("bad: %#v", b.config.ISOUrls)
	}

}

func TestBuilderPrepare_ISOName(t *testing.T) {
	var b Builder
	config := testConfig()
	config["iso_name"] = "my_iso"

	b = Builder{}
	_, warns, err := b.Prepare(config)

	//error out if iso_url or iso_name are set.
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Errorf("should have error: %s", err)
	}

	delete(config, "iso_url")
	config["iso_urls"] = []string{"http://www.hashicorp.com"}
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Errorf("should have error: %s", err)
	}

	// test good
	delete(config, "iso_urls")
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_KeepVM(t *testing.T) {
	var b Builder
	config := testConfig()

	// Bad
	config["keep_vm"] = "foo"
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["keep_vm"] = "always"
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}
