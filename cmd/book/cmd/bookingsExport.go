/*
Copyright © 2022 Tim Drysdale <timothy.d.drysdale@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/ory/viper"
	apiclient "github.com/practable/book/internal/client/client"
	"github.com/practable/book/internal/client/client/admin"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// bookingsExportCmd represents the bookings export commmand
var bookingsExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the manifest from the booking server",
	Long: `Export the manifest from the booking server

example usage:
export BOOK_CLIENT_SCHEME=http
export BOOK_CLIENT_HOST=example.org
export BOOK_CLIENT_BASE_PATH=/book/api/v1
export BOOK_CLIENT_TOKEN=$somesecret
export BOOK_CLIENT_FORMAT=yaml
book bookings export

The exported manifest is printed to stdout, and can be piped to a file if required.  
`,
	Run: func(cmd *cobra.Command, args []string) {

		viper.SetEnvPrefix("BOOK_CLIENT")
		viper.AutomaticEnv()
		viper.SetDefault("host", "localhost")
		viper.SetDefault("scheme", "http")
		viper.SetDefault("format", "yaml")
		viper.SetDefault("base_path", "/api/v1")

		basePath := viper.GetString("base_path")
		host := viper.GetString("host")
		scheme := viper.GetString("scheme")
		token := viper.GetString("token")
		format := strings.ToLower(viper.GetString("format"))

		if token == "" {
			fmt.Println("BOOK_CLIENT_TOKEN not set")
			os.Exit(1)
		}

		switch format {
		case "json", "yaml", "yml":
		default:
			fmt.Println("format can be json or yaml, but not " + format)
			os.Exit(1)
		}

		cfg := apiclient.DefaultTransportConfig().WithHost(host).WithSchemes([]string{scheme}).WithBasePath(basePath)
		auth := httptransport.APIKeyAuth("Authorization", "header", token)
		bc := apiclient.NewHTTPClientWithConfig(nil, cfg)
		timeout := 10 * time.Second
		params := admin.NewExportBookingsParams().WithTimeout(timeout)
		status, err := bc.Admin.ExportBookings(params, auth)
		if err != nil {
			fmt.Printf("Error: failed to export bookings because %s\n", err.Error())
			os.Exit(1)
		}

		switch format {

		case "json":
			mj, err := json.Marshal(status.Payload)
			if err != nil {
				fmt.Printf("Error: failed to marshal exported bookings because %s\n", err.Error())
				os.Exit(1)
			}
			fmt.Println(string(mj))
		default:
			my, err := yaml.Marshal(status.Payload)
			if err != nil {
				fmt.Printf("Error: failed to marshal exported bookings because %s\n", err.Error())
				os.Exit(1)
			}
			fmt.Println(string(my))
		}
		os.Exit(0)
	},
}

func init() {
	bookingsCmd.AddCommand(bookingsExportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
