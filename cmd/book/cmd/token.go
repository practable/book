/*
Copyright © 2020 Tim Drysdale <timothy.d.drysdale@gmail.com>

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
	"math"
	"os"
	"time"

	"github.com/ory/viper"
	"github.com/practable/book/internal/login"
	"github.com/spf13/cobra"
)

// hostCmd represents the host command
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "relay token generates a new token for authenticating to a relay",
	Long: `Set the operating paramters with environment variables, for example

export BOOK_TOKEN_LIFETIME=60
export BOOK_TOKEN_SECRET=some_secret
export BOOK_TOKEN_ADMIN=false
export BOOK_TOKEN_SUBJECT=some_user
export BOOK_TOKEN_AUDIENCE=https://example.org/book
bearer=$(book token)
`,

	Run: func(cmd *cobra.Command, args []string) {

		viper.SetEnvPrefix("BOOK_TOKEN")
		viper.AutomaticEnv()

		viper.SetDefault("lifetime", "1m")
		viper.SetDefault("admin", "false") // default to safest option
		viper.SetDefault("subject", "book-token-cli")

		audience := viper.GetString("audience")
		admin := viper.GetBool("admin")
		lifetimeStr := viper.GetString("lifetime")
		secret := viper.GetString("secret")
		subject := viper.GetString("subject")

		// check inputs
		ok := true

		if audience == "" {
			fmt.Println("BOOK_TOKEN_AUDIENCE not set")
			ok = false
		}

		if secret == "" {
			fmt.Println("BOOK_TOKEN_SECRET not set")
			ok = false
		}

		var scopes []string

		if admin {
			scopes = append(scopes, "book:admin")
		} else {
			scopes = append(scopes, "book:user")
		}

		lifetime, err := time.ParseDuration(lifetimeStr)

		if err != nil {
			fmt.Print("cannot parse duration in BOOK_LIFETIME=" + lifetimeStr)
			ok = false
		}

		if !ok {
			os.Exit(1)
		}

		iat := time.Now().Unix() - 1 //ensure immediately usable
		nbf := iat
		exp := iat + int64(math.Round(lifetime.Seconds()))

		token := login.New(audience, subject, scopes, iat, nbf, exp)

		bearer, err := login.Sign(token, secret)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(bearer)
		os.Exit(0)

	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)

}
