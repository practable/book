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
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ory/viper"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/timdrysdale/interval/internal/config"
	"github.com/timdrysdale/interval/internal/server"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the booking server",
	Long: `Book offers a REST-like API for booking experiments. 
Set configuration parameters with environment variables.

The main parameters have these defaults:

export BOOK_FQDN=https://book.practable.io
export BOOK_PORT=4000
export BOOK_LOG_FILE=/some/logging/location/book.log
export BOOK_PERSIST_DIR=/var/lib/book/

If the BOOK_LOG_FILE is not set, or the file cannot be opened, then logging goes to stderr. Setting it to stdout sends logging to stdout.

You must set the secrets for the booking system and the relay:

export BOOK_ADMIN_SECRET=replace-me-with-some-long-secret
export BOOK_RELAY_SECRET=replace-me-with-another-long-secret

The logging level can be set by:

export BOOK_LOG_LEVEL=warn

Logs default to json format but can also be set to text format (e.g. for development)

export BOOK_LOG_FORMAT=text

Note that persisting bookings to /var/lib/book will require write permission to that directory, 
which can be obtained by running at with elevated permissions e.g. systemd service, or running
as a user which has write priviledges to that directory. Else, specify a user-space directory.

ADVANCED SETTINGS:
You should not need to alter the default values for the following settings, 
but they are available to change if you know what you are doing:

export BOOK_ACCESS_TOKEN_TTL=1h
export BOOK_TIDY_EVERY=1h
export BOOK_MIN_USERNAME_LENGTH=6

After setting the env vars and permissions as required, run with:

$ book serve
`,
	Run: func(cmd *cobra.Command, args []string) {

		viper.SetEnvPrefix("BOOK")
		viper.AutomaticEnv()

		viper.SetDefault("access_token_ttl", "1h")
		viper.SetDefault("fqdn", "https://book.practable.io")
		viper.SetDefault("log_file", "/var/log/book/book.log")
		viper.SetDefault("persist_dir", "/var/lib/book/")
		viper.SetDefault("port", 4000)
		viper.SetDefault("tidy_every", "1h")
		viper.SetDefault("log_level", "warn")
		viper.SetDefault("log_stderr", false)
		viper.SetDefault("log_format", "json")
		viper.SetDefault("min_username_length", 6)

		accessTokenTTL := viper.GetString("access_token_ttl")
		adminSecret := viper.GetString("admin_secret")
		logLevel := viper.GetString("log_level")
		fqdn := viper.GetString("fqdn")
		logFile := viper.GetString("log_file")
		logFormat := viper.GetString("log_format")
		persistDir := viper.GetString("persist_dir")
		port := viper.GetInt("port")
		relaySecret := viper.GetString("relay_secret")
		tidyEvery := viper.GetString("tidy_every")
		minUsernameLength := viper.GetInt("min_username_length")

		// Sanity checks

		if adminSecret == "" || relaySecret == "" {
			fmt.Println("You must set both BOOK_ADMIN_SECRET and BOOK_RELAY_SECRET")
			os.Exit(1)
		}

		accessTokenTTLDuration, err := time.ParseDuration(accessTokenTTL)

		if err != nil {
			fmt.Println("Specify BOOK_ACCESS_TOKEN_TTL duration as string, e.g. 1h, 30m etc")
			os.Exit(1)
		}

		tidyEveryDuration, err := time.ParseDuration(tidyEvery)

		if err != nil {
			fmt.Println("Specify BOOK_TIDY_EVERY duration as string, e.g. 1h, 30m etc")
			os.Exit(1)
		}

		// Set up logging

		switch strings.ToLower(logLevel) {
		case "trace":
			log.SetLevel(log.TraceLevel)
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "info":
			log.SetLevel(log.InfoLevel)
		case "warn":
			log.SetLevel(log.WarnLevel)
		case "error":
			log.SetLevel(log.ErrorLevel)
		case "fatal":
			log.SetLevel(log.FatalLevel)
		case "panic":
			log.SetLevel(log.PanicLevel)
		default:
			fmt.Println("BOOK_LOG_LEVEL can be trace, debug, info, warn, error, fatal or panic but not " + logLevel)
			os.Exit(1)
		}

		switch strings.ToLower(logFormat) {
		case "json":
			log.SetFormatter(&log.JSONFormatter{})
		case "text":
			log.SetFormatter(&log.TextFormatter{})
		default:
			fmt.Println("BOOK_LOG_FORMAT can be json or text but not " + logLevel)
			os.Exit(1)
		}

		if strings.ToLower(logFile) == "stdout" {

			log.SetOutput(os.Stdout) //

		} else {

			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err == nil {
				log.SetOutput(file)
			} else {
				log.Infof("Failed to log to %s, logging to default stderr", logFile)
			}
		}

		// Report useful info

		log.Infof("book version: %s\n", versionString())
		log.Infof("FQDN:[%s]\n", fqdn)
		log.Infof("Listening port: %d\n", port)
		log.Infof("Persistance Directory: [%s]\n", persistDir)
		log.Infof("Persistance NOT IMPLEMENTED\n")
		log.Debugf("Admin secret=[%s...%s]\n", adminSecret[:2], adminSecret[len(adminSecret)-2:])
		log.Debugf("Relay secret=[%s...%s]\n", relaySecret[:2], relaySecret[len(relaySecret)-2:])
		log.Infof("Access token TTL=[%s]\n", accessTokenTTL)
		log.Infof("Tidy every=[%s]\n", tidyEvery)
		log.Infof("Log file=[%s]\n", logFile)
		log.Infof("Log level=[%s]\n", logLevel)

		// Start the server

		c := make(chan os.Signal, 1)

		signal.Notify(c, os.Interrupt)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			for range c {
				cancel()
				<-ctx.Done()
				os.Exit(0)
			}
		}()
		cfg := config.ServerConfig{
			Host:                fqdn,
			Port:                port,
			StoreSecret:         []byte(adminSecret),
			RelaySecret:         []byte(relaySecret),
			MinUserNameLength:   minUsernameLength,
			AccessTokenLifetime: accessTokenTTLDuration,
			Now:                 func() time.Time { return time.Now() },
			PruneEvery:          tidyEveryDuration,
		}

		server.Run(ctx, cfg)

	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
