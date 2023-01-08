package geolite2

import (
	"encoding/json"
	"ip7/pkg/config"
	"log"
	"os"
	"path"
	"time"
)

const cacheExpire = 7 * 24 * 60 * 60 // seconds, 7 days

const (
	dbName = "data"
	tsName = "ts"
)

type cacheTS struct {
	BuildEpoch uint
	CacheTime  int64
}

func cachedDatabase(c *config.GeoLite2DatabaseConfig) ([]byte, error) {
	log.Printf("try to get database %s via cache ...\n", c.Type)
	dirPath := path.Join(config.Cfg.CacheDir, c.Type)
	tsPath := path.Join(dirPath, tsName)
	read := func(p string) ([]byte, error) {
		_, err := os.Stat(p)
		if os.IsNotExist(err) {
			log.Printf("no cached database for type %s\n", c.Type)
			return nil, nil
		} else if err != nil {
			return nil, err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	tsData, err := read(tsPath)
	if tsData == nil {
		return nil, err
	}
	ts := new(cacheTS)
	err = json.Unmarshal(tsData, ts)
	if err != nil {
		return nil, err
	}
	if c.AutoUpdate && time.Now().Unix() > ts.CacheTime+cacheExpire {
		log.Printf("cache expired for type %s\n", c.Type)
		return nil, nil
	}

	dbPath := path.Join(dirPath, dbName)
	return read(dbPath)
}

func cacheDatabase(c *config.GeoLite2DatabaseConfig, db []byte, ts *cacheTS) error {
	dirPath := path.Join(config.Cfg.CacheDir, c.Type)
	log.Printf("caching database to %s ...", dirPath)
	dbPath := path.Join(dirPath, dbName)
	tsPath := path.Join(dirPath, tsName)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}

	write := func(p string, b []byte) error {
		f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		_, err = f.Write(b)
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
		return nil
	}

	err = write(dbPath, db)
	if err != nil {
		return err
	}

	tsData, err := json.Marshal(ts)
	if err != nil {
		return err
	}
	return write(tsPath, tsData)
}
