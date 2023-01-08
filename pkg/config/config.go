package config

import (
	"encoding/json"
	"log"
	"os"
	"path"

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

	if !licKey.Changed {
		for _, k := range []string{"IP7_LICENSE_KEY", "ip7_license_key"} {
			key := os.Getenv(k)
			if key != "" {
				err := licKey.Value.Set(key)
				if err != nil {
					log.Fatalln(err)
				}
				break
			}
		}
	}

	if !urlFormat.Changed {
		err := urlFormat.Value.Set(URLFmt)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if configFile.Changed {
		var err error
		Cfg, err = parseConfigFile(configFile.Value.String())
		if err != nil {
			log.Fatalln(err)
		}
		if dbPath.Changed {
			log.Println("a database is specified through the command line argument, skipping the databases config within the config file")
			Cfg.Databases = []*GeoLite2DatabaseConfig{&GeoLite2DatabaseConfig{
				Type: dbType.Value.String(),
				Path: dbPath.Value.String(),
			}}
		}
	} else {
		Cfg = &Config{
			Databases: []*GeoLite2DatabaseConfig{&GeoLite2DatabaseConfig{
				Type: dbType.Value.String(),
				Path: dbPath.Value.String(),
			}},
			LicenseKey: licKey.Value.String(),
			URLFmt:     urlFormat.Value.String(),
		}
		Cfg.AutoUpdate, _ = fs.GetBool(autoUpdate.Name)
	}

	if Cfg != nil {
		autoUpdateVal, _ := fs.GetBool(autoUpdate.Name)
		licKeyVal, _ := fs.GetString(licKey.Name)
		if Cfg.LicenseKey == "" && licKeyVal != "" {
			Cfg.LicenseKey = licKeyVal
		}
		if Cfg.URLFmt == "" {
			Cfg.URLFmt = URLFmt
		}
		if urlFormat.Changed {
			Cfg.URLFmt = urlFormat.Value.String()
		}
		if autoUpdate.Changed {
			Cfg.AutoUpdate = autoUpdateVal
		}
		for _, db := range Cfg.Databases {
			if db.LicenseKey == "" && licKeyVal != "" {
				db.LicenseKey = licKeyVal
			}
			if autoUpdate.Changed {
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
}
