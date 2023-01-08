package geolite2

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"oss.ac/ip7/pkg/config"

	"github.com/bekcpear/hidepass/pkg/hidepass"
	"github.com/oschwald/geoip2-golang"
)

func get(url string) []byte {
	cli := http.DefaultClient
	log.Println(hidepass.Hide(fmt.Sprintf("getting %s ...\n", url)))
	resp, err := cli.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}
	closeBody := func(r *http.Response) {
		err := r.Body.Close()
		if err != nil {
			log.Println("cannot close the body")
		}
	}
	b, e := io.ReadAll(resp.Body)
	defer closeBody(resp)
	if e != nil {
		log.Fatalln(e)
	}
	if resp.StatusCode != 200 {
		log.Printf("Body: %s\n", string(b))
		log.Printf("Status: %s\n", resp.Status)
		return nil
	}
	return b
}

func getDatabase(c *config.GeoLite2DatabaseConfig) ([]byte, error) {
	log.Printf("updating database %s through network ...\n", c.Type)
	t := get(fmt.Sprintf(config.Cfg.URLFmt, c.Type, c.LicenseKey))
	if t == nil {
		return nil, fmt.Errorf("no tarball body")
	}
	hashTxt := get(fmt.Sprintf(config.Cfg.URLFmt+".sha256", c.Type, c.LicenseKey))
	if hashTxt == nil {
		return nil, fmt.Errorf("get hash sum failed")
	}

	realHash := sha256.Sum256(t)
	hash := make([]byte, 32)
	_, err := hex.Decode(hash, bytes.Split(hashTxt, []byte{' '})[0])
	if err != nil {
		return nil, err
	}
	log.Printf("calculated sha256 sum: %x\n", realHash)
	log.Printf("downloaded sha256 sum: %x\n", hash)
	if bytes.Compare(realHash[:], hash) != 0 {
		return nil, fmt.Errorf("the downloaded tarball does not match the sha256 sum, %s, skip it", string(hash))
	}
	gr, err := gzip.NewReader(bytes.NewReader(t))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tr := tar.NewReader(gr)
	var db []byte
	for n, e := tr.Next(); n != nil && e == nil; n, e = tr.Next() {
		matched, err := regexp.MatchString(`\.mmdb$`, n.Name)
		if err != nil {
			log.Fatalln(err)
		}
		if matched {
			db = make([]byte, n.Size)
			i := 0
			for {
				ii, err := tr.Read(db[i:])
				i += ii
				if err == io.EOF {
					break
				} else if err != nil {
					return nil, err
				}
			}
			break
		}
	}
	return db, nil
}

func getDBReader(c *config.GeoLite2DatabaseConfig) (*geoip2.Reader, error) {
	cacheIt := false
	db, err := cachedDatabase(c)
	if err != nil {
		log.Println(err)
	}
	if db == nil {
		db, err = getDatabase(c)
		if err != nil {
			return nil, err
		}
		cacheIt = true
	}
	dbReader, err := geoip2.FromBytes(db)
	if err != nil {
		return nil, err
	}
	if cacheIt {
		now := time.Now().Unix()
		err = cacheDatabase(c, db, &cacheTS{dbReader.Metadata().BuildEpoch, now})
		if err != nil {
			return nil, err
		}
	}
	return dbReader, nil
}

// Update will loop all config.GeoLite2DatabaseConfig within config.Cfg and fetch the maxmind
// databases through the network or open local databases if the corresponding Path is set.
func Update() {
	for _, db := range config.Cfg.Databases {
		if db.Type != "" && db.Path == "" {
			if !isValidDatabaseType(db.Type) {
				log.Printf("invalid type setting: %s, skipping ...\n", db.Type)
				continue
			}
			dbReader, err := getDBReader(db)
			if err != nil {
				log.Println(err)
			} else {
				latestDatabase[db.Type] = dbReader
			}
		} else if db.Path != "" {
			initDatabaseViaFile(db)
		} else {
			log.Println("skipping an invalid database config with neither Type nor Path ...")
		}
	}
}
