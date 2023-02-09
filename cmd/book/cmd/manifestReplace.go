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
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/ory/viper"
	apiclient "github.com/practable/book/internal/client/client"
	"github.com/practable/book/internal/client/client/admin"
	cmodels "github.com/practable/book/internal/client/models"
	"github.com/practable/book/internal/convert"
	"github.com/practable/book/internal/store"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var manifestReplaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "Replace the manifest in the booking server",
	Long: `Replace the manifest in the booking server

example usage:

export BOOK_CLIENT_TOKEN=$SECRET
export BOOK_CLIENT_SCHEME=https
export BOOK_CLIENT_HOST=example.org
export BOOK_CLIENT_BASE_PATH=/book/api/v1
export BOOK_CLIENT_FORMAT=YAML
book manifest replace manifest.yaml

The manifest must be in a file, default type is YAML.
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

		clientManifest := cmodels.Manifest{}
		storeManifest := store.Manifest{}

		switch format {

		case "yaml", "yml":

			clientManifest, storeManifest, err = convert.YAMLToManifests(mfest)

		case "json":

			clientManifest, storeManifest, err = convert.JSONToManifests(mfest)

		}

		if err != nil {
			fmt.Printf("Error: failed to unmarshal manifest into client format for uploading because %s\n", err.Error())
			os.Exit(1)
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

		cfg := apiclient.DefaultTransportConfig().WithSchemes([]string{scheme}).WithHost(host).WithBasePath(basePath)
		auth := httptransport.APIKeyAuth("Authorization", "header", token)
		bc := apiclient.NewHTTPClientWithConfig(nil, cfg)
		timeout := 10 * time.Second
		params := admin.NewReplaceManifestParams().WithTimeout(timeout).WithManifest(&clientManifest)
		_, err = bc.Admin.ReplaceManifest(params, auth)
		if err != nil {
			fmt.Printf("Error: failed to replace manifest because %s\n", err.Error())
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
