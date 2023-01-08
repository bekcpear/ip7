package serve

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"oss.ac/ip7/pkg/geolite2"

	"github.com/spf13/cobra"
)

type result struct {
	Status  int
	Query   string
	Result  *geolite2.Result
	Message string
}

var checkingHeader string

func myHandle(w http.ResponseWriter, req *http.Request) {
	sourceIP := req.Header.Get("X-Real-IP")
	if sourceIP == "" {
		addr, err := net.ResolveTCPAddr("tcp", req.RemoteAddr)
		if err != nil {
			log.Printf("cannot resolve the remote addr: %s\n", req.RemoteAddr)
		}
		sourceIP = addr.IP.String()
	}
	log.Printf("request '%s' from %s (%s)", req.URL, req.RemoteAddr, sourceIP)

	var (
		status = 200
		rr     []byte
		err    error
	)

	if req.URL.Path == "/" || req.URL.Path == "/index.html" {
		w.Header().Set("Content-Type", "text/html")
		rr = []byte(`<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width,initial-scale=1,viewport-fit=cover,shrink-to-fit=no" />
</head>
<body>
<p>
This is a testing web service, for private use only.
</p>
<p>
This service includes GeoLite2 data created by MaxMind, available from
<a href="https://www.maxmind.com">https://www.maxmind.com</a>.
</p>
</body>
</html>`)
	} else {
		w.Header().Set("Content-Type", "application/json")
		r := new(result)
		qStr := strings.TrimLeft(req.URL.Path, "/")

		agree := true
		if checkingHeader != "" {
			agree, _ = strconv.ParseBool(req.Header.Get(checkingHeader))
		}

		if !agree {
			status = 406
			r.Query = qStr
			r.Message = "illegal request"
		} else {
			var q net.IP
			if req.URL.Path == "/self" {
				r.Query = sourceIP
				q = net.ParseIP(sourceIP)
			} else {
				r.Query = qStr
				q = net.ParseIP(qStr)
				if q == nil {
					status = 400
					r.Message = "invalid IP address"
				}
			}

			if q != nil {
				r.Result = geolite2.Query(q)
			}
		}

		r.Status = status
		rr, err = json.Marshal(r)
		if err != nil {
			log.Println(err)
			status = 500
			rr = []byte("internal error")
		}
	}

	w.WriteHeader(status)
	ws, err := w.Write(rr)
	if err != nil {
		log.Println(err)
	}

	log.Printf("response: %d (%d bytes) to %s (%s)", status, ws, req.RemoteAddr, sourceIP)
}

func Serve(c *cobra.Command) {
	fs := c.Flags()

	listen, err := fs.GetIP("listen")
	if err != nil {
		log.Fatalln(err)
	}
	port, err := fs.GetInt("port")
	if err != nil {
		log.Fatalln(err)
	}
	checkingHeader, err = fs.GetString("checking-header")
	if err != nil {
		log.Fatalln(err)
	}

	address := &net.TCPAddr{
		IP:   listen,
		Port: port,
	}
	listener, err := net.Listen(address.Network(), address.String())
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("listened on %s\n", address)

	// schedule to update databases every day
	t := time.NewTicker(time.Hour * 24)
	defer t.Stop()
	tDone := make(chan bool)
	go func() {
		log.Printf("scheduled to update database every day ...")
		for {
			select {
			case <-tDone:
				return
			case <-t.C:
				geolite2.Update()
			}
		}
	}()

	http.DefaultServeMux.HandleFunc("/", myHandle)
	err = http.Serve(listener, nil)
	log.Println(err)
	tDone <- true
}
