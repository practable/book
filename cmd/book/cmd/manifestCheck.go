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

	"github.com/spf13/cobra"
	"github.com/timdrysdale/interval/internal/store"
	"gopkg.in/yaml.v2"
)

// checkCmd represents the check command
var manifestCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check a manifest for correctness",
	Long: `Check a manifest for correctness, without affecting the manifest loaded into the booking server.

example usage:
(no environment variables are required for this command)
book manifest check manifest.yaml

The manifest must be in a file, in yaml format.
`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(os.Args) < 4 {
			fmt.Println("usage: book manifest check <file.yaml>")
			os.Exit(1)
		}

		f := os.Args[3]
		mfest, err := ioutil.ReadFile(f)
		if err != nil {
			fmt.Printf("Error: failed to read manifest from file %s because %s\n", f, err.Error())
			os.Exit(1)
		}

		m := store.Manifest{}

		err = yaml.Unmarshal(mfest, &m)
		if err != nil {
			fmt.Printf("Error: failed to unmarshal manifest from file because %s\n", err.Error())
			os.Exit(1)
		}

		err, msgs := store.CheckManifest(m)

		if err != nil {
			fmt.Println(err.Error())
			for k, v := range msgs {
				fmt.Println(strconv.Itoa(k) + ": " + v)
			}
		}

		os.Exit(0)

	},
}

func init() {
	manifestCmd.AddCommand(manifestCheckCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
