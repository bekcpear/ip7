package root

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"oss.ac/ip7/internal/cmd/serve"
	"oss.ac/ip7/pkg/config"
	"oss.ac/ip7/pkg/geolite2"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ip7 [flags] ip-address",
		Short: "An IP address checker, powered by MaxMind GeoLite2 databases.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			config.Initialize(cmd)
			geolite2.Update()
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 1 {
				log.Printf("only one IP address per query, ignore %s\n", strings.Join(args[1:], ","))
			}
			ip := net.ParseIP(args[0])
			if ip == nil {
				log.Fatalf("%s is not an IP address\n", args[0])
			}
			rr, err := json.Marshal(geolite2.Query(ip))
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Printf("%s", string(rr))
		},
		TraverseChildren: true,
	}

	cmd.AddCommand(serve.NewServeCmd())
	cmd.CompletionOptions.HiddenDefaultCmd = true
	cmd.Flags().SortFlags = false

	fs := cmd.PersistentFlags()
	fs.SortFlags = false
	fs.StringP("type", "t", "City", "The type of GeoLite2 Database (ASN, City, Country)")
	fs.StringP("database", "d", "", "The path to GeoLite2 Database (GeoIP2 Binary Format)")
	fs.StringP("license-key", "k", "", "The License key which used to download all GeoLite2 databases,"+
		"\n"+" it can also be get from env variables, 'IP7_LICENSE_KEY' or 'ip7_license_key', with a low priority,"+
		"\n"+" missing license key settings within the configuration file will be replaced by this setting")
	fs.BoolP("auto-update", "u", false, "Auto update the database from maxmind with the license key,"+
		"\n"+" an explicit setting will override the setting within the configuration file")
	fs.StringP("url-format", "m", "", "URL format used to download maxmind databases"+
		"\n"+" an explicit setting will override the setting within the configuration file"+
		"\n"+" (default "+config.URLFmt+")")
	fs.StringP("config", "c", "", "The config file path")

	return cmd
}
