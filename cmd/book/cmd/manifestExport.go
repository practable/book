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
	"github.com/spf13/cobra"
	apiclient "github.com/practable/book/internal/client/client"
	"github.com/practable/book/internal/client/client/admin"
	"gopkg.in/yaml.v2"
)

// checkCmd represents the check command
var manifestExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the manifest from the booking server",
	Long: `Export the manifest from the booking server

example usage:
export BOOK_CLIENT_HOST=localhost:4000
export BOOK_CLIENT_SCHEME=http
export BOOK_CLIENT_TOKEN=$secret
export BOOK_CLIENT_FORMAT=yaml
book manifest export

The exported manifest is printed to stdout, and can be piped to a file if required.  
`,
	Run: func(cmd *cobra.Command, args []string) {

		viper.SetEnvPrefix("BOOK_CLIENT")
		viper.AutomaticEnv()
		viper.SetDefault("host", "book.practable.io")
		viper.SetDefault("scheme", "https")
		viper.SetDefault("format", "yaml")

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

		cfg := apiclient.DefaultTransportConfig().WithHost(host).WithSchemes([]string{scheme})
		auth := httptransport.APIKeyAuth("Authorization", "header", token)
		bc := apiclient.NewHTTPClientWithConfig(nil, cfg)
		timeout := 10 * time.Second
		params := admin.NewExportManifestParams().WithTimeout(timeout)
		status, err := bc.Admin.ExportManifest(params, auth)
		if err != nil {
			fmt.Printf("Error: failed to export manifest because %s\n", err.Error())
			os.Exit(1)
		}

		switch format {

		case "json":
			mj, err := json.Marshal(status.Payload)
			if err != nil {
				fmt.Printf("Error: failed to marshal exported manifest because %s\n", err.Error())
				os.Exit(1)
			}
			fmt.Println(string(mj))
		default:
			my, err := yaml.Marshal(status.Payload)
			if err != nil {
				fmt.Printf("Error: failed to marshal exported manifest because %s\n", err.Error())
				os.Exit(1)
			}
			fmt.Println(string(my))
		}
		os.Exit(0)
	},
}

func init() {
	manifestCmd.AddCommand(manifestExportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
