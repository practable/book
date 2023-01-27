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
	"strconv"
	"strings"
	"time"

	"net/http"
	_ "net/http/pprof" //ok in production, probably? https://medium.com/google-cloud/continuous-profiling-of-go-programs-96d4416af77b

	"github.com/ory/viper"
	"github.com/practable/book/internal/config"
	"github.com/practable/book/internal/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		viper.SetDefault("check_every", "1m")
		viper.SetDefault("disable_cancel_after_use", "false")
		viper.SetDefault("fqdn", "https://book.practable.io")
		viper.SetDefault("log_file", "/var/log/book/book.log")
		viper.SetDefault("log_level", "warn")
		viper.SetDefault("log_format", "json")
		viper.SetDefault("min_username_length", 6)
		viper.SetDefault("persist_dir", "/var/lib/book/")
		viper.SetDefault("port", 4000)
		viper.SetDefault("profile", "true")
		viper.SetDefault("profile_port", 6060)
		viper.SetDefault("request_timeout", "1m")
		viper.SetDefault("tidy_every", "1h")

		accessTokenTTL := viper.GetString("access_token_ttl")
		adminSecret := viper.GetString("admin_secret")
		checkEvery := viper.GetString("check_every")
		disableCancelAfterUse := viper.GetBool("disable_cancel_after_use")
		logLevel := viper.GetString("log_level")
		fqdn := viper.GetString("fqdn")
		logFile := viper.GetString("log_file")
		logFormat := viper.GetString("log_format")
		persistDir := viper.GetString("persist_dir")
		port := viper.GetInt("port")
		profile := viper.GetBool("profile")
		profilePort := viper.GetInt("profile_port")
		relaySecret := viper.GetString("relay_secret")
		requestTimeout := viper.GetString("request_timeout")

		tidyEvery := viper.GetString("tidy_every")
		minUsernameLength := viper.GetInt("min_username_length")

		// Sanity checks

		if adminSecret == "" || relaySecret == "" {
			fmt.Println("You must set both BOOK_ADMIN_SECRET and BOOK_RELAY_SECRET")
			os.Exit(1)
		}

		accessTokenTTLDuration, err := time.ParseDuration(accessTokenTTL)

		if err != nil {
			fmt.Println("Specify BOOK_ACCESS_TOKEN_TTL duration as string, e.g. 5m, 1h etc")
			os.Exit(1)
		}

		checkEveryDuration, err := time.ParseDuration(checkEvery)

		if err != nil {
			fmt.Println("Specify BOOK_CHECK_EVERY duration as string, e.g. 30s, 1m etc")
			os.Exit(1)
		}

		requestTimeoutDuration, err := time.ParseDuration(requestTimeout)

		if err != nil {
			fmt.Println("Specify BOOK_REQUEST_TIMEOUT duration as string, e.g. 30s, 1m etc")
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
		log.Infof("book version: %s", versionString())
		log.Debugf("Admin secret: [%s...%s]", adminSecret[:4], adminSecret[len(adminSecret)-4:]) // partial reveal of secret in our logs
		log.Debugf("Relay secret: [%s...%s]", relaySecret[:4], relaySecret[len(relaySecret)-4:]) // at debug level only
		log.Infof("Access token TTL: [%s]", accessTokenTTL)
		log.Infof("Check grace period expiries every [%s]", checkEvery)
		log.Infof("Disable cancel after use: %t", disableCancelAfterUse)
		log.Infof("FQDN: [%s]\n", fqdn)
		log.Infof("Listening port: %d", port)
		log.Infof("Log file: [%s]", logFile)
		log.Infof("Log level: [%s]", logLevel)
		log.Infof("Persistance Directory: [%s]", persistDir)
		log.Infof("Persistance NOT IMPLEMENTED")
		log.Infof("Profiling on: [%t]", profile)
		log.Infof("Profile port: [%d]", profilePort)
		log.Infof("Request timeout: [%s]", requestTimeout)
		log.Infof("Tidy every: [%s]", tidyEvery)

		// Optionally start the profiling server
		if profile {
			go func() {
				url := "localhost:" + strconv.Itoa(profilePort)
				err := http.ListenAndServe(url, nil)
				if err != nil {
					log.Errorf(err.Error())
				}
			}()
		}

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
			AccessTokenLifetime:   accessTokenTTLDuration,
			CheckEvery:            checkEveryDuration,
			DisableCancelAfterUse: disableCancelAfterUse,
			Host:                  fqdn,
			MinUserNameLength:     minUsernameLength,
			Now:                   func() time.Time { return time.Now() },
			Port:                  port,
			PruneEvery:            tidyEveryDuration,
			StoreSecret:           []byte(adminSecret),
			RelaySecret:           []byte(relaySecret),
			RequestTimeout:        requestTimeoutDuration,
		}

		s := server.New(cfg)
		s.Run(ctx)

	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
