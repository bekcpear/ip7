package geolite2

import (
	"ip7/pkg/config"
	"os"
	"path"
	"testing"
)

func Test_Update(t *testing.T) {
	config.Cfg = &config.Config{
		Databases: []*config.GeoLite2DatabaseConfig{
			&config.GeoLite2DatabaseConfig{
				Type:       tCity,
				AutoUpdate: true,
				LicenseKey: "",
			},
		},
		AutoUpdate: true,
		LicenseKey: "",
		URLFmt:     "http://127.0.0.1/GeoLite2-%s%s_20230106.tar.gz",
	}
	cDir, _ := os.UserCacheDir()
	config.Cfg.CacheDir = path.Join(cDir, "ip7")

	Update()
}
