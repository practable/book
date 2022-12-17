/*
Copyright © 2021 Tim Drysdale <timothy.d.drysdale@gmail.com>

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
	apiclient "github.com/timdrysdale/interval/internal/client/client"
	"github.com/timdrysdale/interval/internal/client/client/admin"
)

// setstatusCmd represents the setstatus command
var statusSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the lock status and message of the day",
	Long: `Set server details with environment variables 
and pass settings as arguments. 
For example:
export BOOKCLIENT_HOST=book.practable.io
export BOOKCLIENT_SCHEME=https
export BOOKCLIENT_TOKEN=$secret
book setstatus lock "Bookings are currently closed for maintenance"
book setstatus unlock "Bookings are open"
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
			fmt.Println("usage: book setstatus [lock/unlock] message")
			os.Exit(1)
		}

		lockword := strings.ToLower(os.Args[2])
		lock := false

		if lockword == "lock" {
			lock = true
		} else if lockword != "unlock" {
			fmt.Printf("lock status of %s not understood\n", lockword)
			fmt.Println("usage: book setstatus [lock/unlock] message")
			os.Exit(1)
		}

		message := os.Args[3]

		cfg := apiclient.DefaultTransportConfig().WithHost(host).WithSchemes([]string{scheme})
		auth := httptransport.APIKeyAuth("Authorization", "header", token)
		bc := apiclient.NewHTTPClientWithConfig(nil, cfg)
		timeout := 10 * time.Second
		params := admin.NewSetLockParams().WithTimeout(timeout).WithLock(lock).WithMsg(&message)
		status, err := bc.Admin.SetLock(params, auth)
		if err != nil {
			fmt.Printf("Error: failed to set status because %s\n", err.Error())
			os.Exit(1)
		}

		pretty, err := json.MarshalIndent(status.Payload, "", "\t")
		if err != nil {
			fmt.Printf("Error: failed to format the response because %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Println(string(pretty))
		os.Exit(0)

	},
}

func init() {
	statusCmd.AddCommand(statusSetCmd)
}