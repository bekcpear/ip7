package geolite2

import (
	"bytes"
	"encoding/json"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type Result struct {
	ASN     *geoip2.ASN
	City    *geoip2.City
	Country *geoip2.Country
}

// Query will query the ip address from all possible maxmindDatabases,
// and return the combined Result.
func Query(ip net.IP) *Result {
	r := new(Result)
	r.ASN = asn(ip)
	r.City = city(ip)
	r.Country = country(ip)
	return r
}

func (r *Result) MarshalJSON() ([]byte, error) {
	j := []byte("{")
	if r.ASN != nil {
		marshal, err := json.Marshal(r.ASN)
		if err != nil {
			return nil, err
		}
		j = append(j, `"ASN":`...)
		j = append(j, marshal...)
		j = append(j, ',')
	}
	if r.City != nil {
		marshal, err := json.Marshal(r.City)
		if err != nil {
			return nil, err
		}
		j = append(j, `"City":`...)
		j = append(j, marshal...)
		j = append(j, ',')
	}
	if r.Country != nil {
		marshal, err := json.Marshal(r.Country)
		if err != nil {
			return nil, err
		}
		j = append(j, `"Country":`...)
		j = append(j, marshal...)
		j = append(j, ',')
	}
	j = append(bytes.TrimSuffix(j, []byte{','}), '}')

	return j, nil
}
