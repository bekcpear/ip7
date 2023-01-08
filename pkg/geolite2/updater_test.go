package geolite2

import (
	"testing"

	"oss.ac/ip7/pkg/config"

	"github.com/spf13/cobra"
)

func Test_Update(t *testing.T) {
	config.Initialize(&cobra.Command{})
	config.Cfg.URLFmt = "http://test.bitbili.net/ip7/db?type=%s&license_key=%s&suffix=tar.gz"
	config.Cfg.Databases[0].Type = tCity
	config.Cfg.Databases[0].LicenseKey = "jdklji1JKD.Lj-~i1jk728931_dsa"

	Update()
}
