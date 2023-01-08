package config

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/bekcpear/hidepass/pkg/hidepass"
	"github.com/spf13/cobra"
)

const (
	URLFmt = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-%s&license_key=%s&suffix=tar.gz"
)

type GeoLite2DatabaseConfig struct {
	Type       string `json:"type"`
	Path       string `json:"path"`
	AutoUpdate bool   `json:"autoUpdate"`
	LicenseKey string `json:"licenseKey"`
}

type Config struct {
	Databases  []*GeoLite2DatabaseConfig `json:"databases"`
	AutoUpdate bool                      `json:"autoUpdate"`
	LicenseKey string                    `json:"licenseKey"`
	URLFmt     string                    `json:"urlFmt"`
	CacheDir   string                    `json:"cacheDir"`
}

// Cfg is the global configuration variable through the entire program,
// it stores the all GeoLite2 database configurations, including Type and
// LicenseKey.
// You should call Initialize function first to initialize this variable.
var Cfg *Config

func parseConfigFile(file string) (*Config, error) {
	r, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	c := new(Config)
	err = json.Unmarshal(r, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Initialize parse the command line arguments, read license key from env,
// parse the json configuration file and then initialize the Cfg variable.
func Initialize(c *cobra.Command) {
	fs := c.Flags()
	var (
		dbType     = fs.Lookup("type")
		dbPath     = fs.Lookup("database")
		licKey     = fs.Lookup("license-key")
		autoUpdate = fs.Lookup("auto-update")
		urlFormat  = fs.Lookup("url-format")
		configFile = fs.Lookup("config")
	)
	var (
		dbTypeVal     string
		dbPathVal     string
		licKeyVal     string
		autoUpdateVal bool
		urlFormatVal  = URLFmt
	)

	if licKey != nil {
		if !licKey.Changed {
			for _, k := range []string{"IP7_LICENSE_KEY", "ip7_license_key"} {
				key := os.Getenv(k)
				if key != "" {
					licKeyVal = key
					err := licKey.Value.Set(key)
					if err != nil {
						log.Fatalln(err)
					}
					break
				}
			}
		} else {
			licKeyVal = licKey.Value.String()
		}
	}

	if urlFormat != nil && urlFormat.Changed {
		urlFormatVal = urlFormat.Value.String()
	}
	if dbType != nil {
		dbTypeVal = dbType.Value.String()
	}
	if dbPath != nil && dbPath.Changed {
		dbPathVal = dbPath.Value.String()
	}
	if autoUpdate != nil && autoUpdate.Changed {
		autoUpdateVal, _ = fs.GetBool(autoUpdate.Name)
	}

	if configFile != nil && configFile.Changed {
		var err error
		Cfg, err = parseConfigFile(configFile.Value.String())
		if err != nil {
			log.Fatalln(err)
		}
		if dbPath != nil && dbPath.Changed {
			log.Println("a database is specified through the command line argument, skipping the databases config within the config file")
			Cfg.Databases = []*GeoLite2DatabaseConfig{&GeoLite2DatabaseConfig{
				Type: dbTypeVal,
				Path: dbPathVal,
			}}
		}
	} else {
		Cfg = &Config{
			Databases: []*GeoLite2DatabaseConfig{&GeoLite2DatabaseConfig{
				Type: dbTypeVal,
				Path: dbPathVal,
			}},
			LicenseKey: licKeyVal,
			URLFmt:     urlFormatVal,
			AutoUpdate: autoUpdateVal,
		}
	}

	if Cfg != nil {
		if Cfg.LicenseKey == "" && licKeyVal != "" {
			Cfg.LicenseKey = licKeyVal
		}
		if Cfg.URLFmt == "" || (urlFormat != nil && urlFormat.Changed) {
			Cfg.URLFmt = urlFormatVal
		}
		if autoUpdate != nil && autoUpdate.Changed {
			Cfg.AutoUpdate = autoUpdateVal
		}
		for _, db := range Cfg.Databases {
			if db.LicenseKey == "" && licKeyVal != "" {
				db.LicenseKey = licKeyVal
			}
			if autoUpdate != nil && autoUpdate.Changed {
				db.AutoUpdate = autoUpdateVal
			}
		}
		if Cfg.CacheDir == "" {
			cacheDir, err := os.UserCacheDir()
			if err != nil {
				log.Fatalln(err)
			}
			Cfg.CacheDir = path.Join(cacheDir, "ip7")
		}
	} else {
		log.Fatalln("initial configurations failed")
	}

	err := hidepass.SetConfig([]byte(`{"regex": ["(?:license_key=)([a-zA-Z0-9_~\\.-]+)(?:&suffix=)"]}`))
	if err != nil {
		log.Fatalln(err)
	}
}
