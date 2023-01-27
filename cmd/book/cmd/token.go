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
	"fmt"
	"os"
	"time"

	"github.com/ory/viper"
	"github.com/spf13/cobra"
	"github.com/practable/book/internal/login"
)

// tokenCmd represents the token command
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "session token generates a new token for authenticating to book",
	Long: `Set the operating paramters with environment variables, for example

export BOOK_CLIENT_SECRET=somesecret
export BOOK_CLIENT_TOKEN_TTL=300
export BOOK_CLIENT_TOKEN_ADMIN=true
export BOOK_CLIENT_TOKEN_AUD=https://book.example.io
export BOOK_CLIENT_TOKEN_SUB=someuser
bearer=$(book token)

If you want to set a future NBF date, then specify the NBF in RFC3339 format
export BOOK_CLIENT_TOKEN_NBF=2022-10-12T07:20:50Z
`,

	Run: func(cmd *cobra.Command, args []string) {

		viper.SetEnvPrefix("BOOK_CLIENT")
		viper.AutomaticEnv()

		viper.SetDefault("token_ttl", "1m")
		viper.SetDefault("token_admin", "false")
		viper.SetDefault("token_aud", "https://book.practable.io")

		admin := viper.GetBool("token_admin")
		aud := viper.GetString("token_aud")
		ttl := viper.GetString("token_ttl")
		nbfstr := viper.GetString("token_nbf")
		secret := viper.GetString("secret")
		sub := viper.GetString("token_sub")

		// check inputs

		if aud == "" {
			fmt.Println("BOOK_CLIENT_TOKEN_AUD not set")
			os.Exit(1)
		}
		if sub == "" {
			fmt.Println("BOOK_CLIENT_TOKEN_SUB not set")
			os.Exit(1)
		}
		if secret == "" {
			fmt.Println("BOOK_CLIENT_SECRET not set")
			os.Exit(1)
		}
		if ttl == "" {
			fmt.Println("BOOK_CLIENT_TOKEN_TTL not set")
			os.Exit(1)
		}

		iat := time.Now().Unix() - 1 // need immediately usable tokens for testing
		nbf := iat                   //update below if NBF is specified

		if nbfstr != "" {
			t, e := time.Parse(
				time.RFC3339,
				nbfstr)
			if e != nil {
				fmt.Printf("BOOK_CLIENT_TOKEN_NBF time format error: %s\n", e.Error())
			}
			// ensure future date
			if t.After(time.Now()) {
				nbf = t.Unix()
			}
		}

		d, err := time.ParseDuration(ttl)

		if err != nil {
			fmt.Printf("BOOK_CLIENT_TOKEN_TTL duration format error: %s\n", err.Error())
		}

		exp := nbf + int64(d/time.Second)

		var scopes []string

		if admin {
			scopes = []string{"booking:admin"}
		} else {
			scopes = []string{"booking:user"}
		}

		token := login.New(aud, sub, scopes, iat, nbf, exp)
		stoken, err := login.Sign(token, secret)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(stoken)
		os.Exit(0)

	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tokenCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tokenCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
