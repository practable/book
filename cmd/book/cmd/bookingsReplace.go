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
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/ory/viper"
	"github.com/spf13/cobra"
	apiclient "github.com/timdrysdale/interval/internal/client/client"
	"github.com/timdrysdale/interval/internal/client/client/admin"
	"github.com/timdrysdale/interval/internal/client/models"
	"gopkg.in/yaml.v2"
)

// bookingsReplaceCmd represents the replace bookings command
var bookingsReplaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "Replace the bookings in the booking server",
	Long: `Replace the bookings in the booking server

example usage:

book bookings replace bookings.yaml

The bookings must be in a file, in yaml format.
`,
	Run: func(cmd *cobra.Command, args []string) {

		viper.SetEnvPrefix("BOOKCLIENT")
		viper.AutomaticEnv()
		viper.SetDefault("host", "book.practable.io")
		viper.SetDefault("scheme", "https")

		host := viper.GetString("host")
		scheme := viper.GetString("scheme")
		token := viper.GetString("token")

		if token == "" {
			fmt.Println("BOOKCLIENT_TOKEN not set")
			os.Exit(1)
		}

		if len(os.Args) < 4 {
			fmt.Println("usage: book bookings replace <file.yaml>")
			os.Exit(1)
		}

		f := os.Args[3]
		mfest, err := ioutil.ReadFile(f)
		if err != nil {
			fmt.Printf("Error: failed to read bookings from file %s because %s\n", f, err.Error())
			os.Exit(1)
		}

		cfg := apiclient.DefaultTransportConfig().WithHost(host).WithSchemes([]string{scheme})
		auth := httptransport.APIKeyAuth("Authorization", "header", token)
		bc := apiclient.NewHTTPClientWithConfig(nil, cfg)
		timeout := 10 * time.Second

		// convert yaml file to models.Bookings
		var bm models.Bookings
		err = yaml.Unmarshal(mfest, &bm)
		if err != nil {
			fmt.Printf("Error: failed to parse bookings because %s\n", err.Error())
			os.Exit(1)
		}

		params := admin.NewReplaceBookingsParams().WithTimeout(timeout).WithBookings(bm)
		_, err = bc.Admin.ReplaceBookings(params, auth)
		if err != nil {
			fmt.Printf("Error: failed to replace bookings because %s\n", err.Error())
			os.Exit(1)
		}

		// print nothing so that we can tell successful replacement
		// admin can always get the status with a separate command

		os.Exit(0)

	},
}

func init() {
	bookingsCmd.AddCommand(bookingsReplaceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
