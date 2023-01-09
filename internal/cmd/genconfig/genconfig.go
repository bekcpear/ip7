package genconfig

import (
	"encoding/json"
	"log"
	"os"

	"github.com/spf13/cobra"
	"oss.ac/ip7/pkg/config"
)

func NewGenConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genconfig",
		Short: "generate an example configuration",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := json.MarshalIndent(config.Cfg, "", "  ")
			if err != nil {
				log.Fatalln(err)
			}
			_, err = os.Stdout.Write(c)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	return cmd
}
