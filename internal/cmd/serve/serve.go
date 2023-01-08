package serve

import (
	"net"

	"oss.ac/ip7/pkg/serve"

	"github.com/spf13/cobra"
)

func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run an HTTP API service",
		Run: func(cmd *cobra.Command, args []string) {
			serve.Serve(cmd)
		},
	}
	cmd.Flags().SortFlags = false
	fs := cmd.PersistentFlags()
	fs.SortFlags = false
	fs.IPP("listen", "l", net.IPv4(127, 0, 0, 1), "listen address")
	fs.IntP("port", "p", 3789, "listen port")
	fs.String("checking-header", "", "an extra request header which should be set by client"+
		"\n"+" to True to make the request valid")

	return cmd
}
