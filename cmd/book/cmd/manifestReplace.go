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
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/ory/viper"
	"github.com/spf13/cobra"
	apiclient "github.com/timdrysdale/interval/internal/client/client"
	"github.com/timdrysdale/interval/internal/client/client/admin"
	cmodels "github.com/timdrysdale/interval/internal/client/models"
	"github.com/timdrysdale/interval/internal/store"
	"gopkg.in/yaml.v2"
)

// checkCmd represents the check command
var manifestReplaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "Replace the manifest in the booking server",
	Long: `Replace the manifest in the booking server

example usage:

book manifest replace manifest.json

The manifest must be in a file, default is json. Yaml not currently available
`,
	Run: func(cmd *cobra.Command, args []string) {

		viper.SetEnvPrefix("BOOKCLIENT")
		viper.AutomaticEnv()
		viper.SetDefault("host", "book.practable.io")
		viper.SetDefault("scheme", "https")
		viper.SetDefault("format", "json")

		host := viper.GetString("host")
		scheme := viper.GetString("scheme")
		token := viper.GetString("token")
		format := strings.ToLower(viper.GetString("format"))

		if token == "" {
			fmt.Println("BOOKCLIENT_TOKEN not set")
			os.Exit(1)
		}

		if len(os.Args) < 4 {
			fmt.Println("usage: book manifest replace <file>")
			os.Exit(1)
		}

		switch format {

		case "json", "yaml", "yml":

		default:
			fmt.Println("format can be json or yaml, but not " + format)
			os.Exit(1)
		}

		f := os.Args[3]
		mfest, err := ioutil.ReadFile(f)
		if err != nil {
			fmt.Printf("Error: failed to read manifest from file %s because %s\n", f, err.Error())
			os.Exit(1)
		}

		storeManifest := store.Manifest{}
		clientManifest := cmodels.Manifest{}

		switch format {
		case "json":

			err = json.Unmarshal(mfest, &storeManifest)
			if err != nil {
				fmt.Printf("Error: failed to unmarshal manifest into store format for checking because: %s\n", err.Error())
				os.Exit(1)
			}
			err = json.Unmarshal(mfest, &clientManifest)
			if err != nil {
				fmt.Printf("Error: failed to unmarshal manifest into client format for uploading because: %s\n", err.Error())
				os.Exit(1)
			}

		case "yaml", "yml":

			fmt.Println("YAML format unavailable at present")
			os.Exit(1)

			err = yaml.Unmarshal(mfest, &storeManifest)
			if err != nil {
				fmt.Printf("Error: failed to unmarshal manifest into store format for checking because: %s\n", err.Error())
				os.Exit(1)
			}
			err = yaml.Unmarshal(mfest, &clientManifest)
			if err != nil {
				fmt.Printf("Error: failed to unmarshal manifest into client format for uploading because: %s\n", err.Error())
				os.Exit(1)
			}

		}

		// check manifest before uploading
		err, msgs := store.CheckManifest(storeManifest)

		if err != nil {
			fmt.Println(err.Error())
			for k, v := range msgs {
				fmt.Println(strconv.Itoa(k) + ": " + v)
			}
			os.Exit(1)
		}

		// upload

		cfg := apiclient.DefaultTransportConfig().WithHost(host).WithSchemes([]string{scheme})
		auth := httptransport.APIKeyAuth("Authorization", "header", token)
		bc := apiclient.NewHTTPClientWithConfig(nil, cfg)
		timeout := 10 * time.Second
		params := admin.NewReplaceManifestParams().WithTimeout(timeout).WithManifest(&clientManifest)
		status, err := bc.Admin.ReplaceManifest(params, auth)
		if err != nil {
			fmt.Printf("Error: failed to replace manifest because %s\n", err.Error())
			fmt.Println(status.Payload)
			os.Exit(1)
		}

		// print nothing so that we can tell successful replacement
		// admin can always get the status with a separate command

		os.Exit(0)

	},
}

func init() {
	manifestCmd.AddCommand(manifestReplaceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}