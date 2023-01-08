package geolite2

import (
	"log"
	"net"

	"oss.ac/ip7/pkg/config"

	"github.com/oschwald/geoip2-golang"
)

const (
	tASN     = "ASN"
	tCity    = "City"
	tCountry = "Country"
)

var latestDatabase = make(map[string]*geoip2.Reader)

func isValidDatabaseType(t string) bool {
	switch t {
	case tASN, tCity, tCountry:
		return true
	}
	return false
}

func asn(ip net.IP) *geoip2.ASN {
	db := latestDatabase[tASN]
	if db != nil {
		r, err := db.ASN(ip)
		if err != nil {
			log.Println(err)
			return nil
		}
		return r
	}
	return nil
}
func city(ip net.IP) *geoip2.City {
	db := latestDatabase[tCity]
	if db != nil {
		r, err := db.City(ip)
		if err != nil {
			log.Println(err)
			return nil
		}
		return r
	}
	return nil
}
func country(ip net.IP) *geoip2.Country {
	db := latestDatabase[tCountry]
	if db != nil {
		r, err := db.Country(ip)
		if err != nil {
			log.Println(err)
			return nil
		}
		return r
	}
	return nil
}

func initDatabaseViaFile(c *config.GeoLite2DatabaseConfig) {
	log.Printf("init database from local file '%s' ...\n", c.Path)
	r, err := geoip2.Open(c.Path)
	if err != nil {
		log.Fatalln(err)
	}
	updateType := func(t string) {
		if c.Type != t {
			if c.Type != "" {
				log.Printf(
					"the recorded type '%s' for '%s' does not match the actual type '%s', correct it ...\n",
					c.Type, c.Path, t)
			}
			c.Type = t
		}
		latestDatabase[t] = r
	}
	switch r.Metadata().DatabaseType {
	case "GeoLite2-ASN":
		updateType(tASN)
	case "GeoLite2-City":
		updateType(tCity)
	case "GeoLite2-Country":
		updateType(tCountry)
	default:
		log.Fatalf("unsupported database type '%s' for file '%s'\n",
			r.Metadata().DatabaseType, c.Path)
	}
}
