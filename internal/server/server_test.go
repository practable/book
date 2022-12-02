package server

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/interval/internal/config"
)

var debug bool
var cfg config.ServerConfig

func init() {
	debug = false
	if debug {
		log.SetReportCaller(true)
		log.SetLevel(log.TraceLevel)
		log.SetFormatter(&logrus.TextFormatter{FullTimestamp: false, DisableColors: true})
		defer log.SetOutput(os.Stdout)

	} else {
		log.SetLevel(log.WarnLevel)
		var ignore bytes.Buffer
		logignore := bufio.NewWriter(&ignore)
		log.SetOutput(logignore)
	}

}
func TestMain(m *testing.M) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port, err := freeport.GetFreePort()
	if err != nil {
		panic(err)
	}

	host := "http://[::]:" + strconv.Itoa(port)

	cfg = config.ServerConfig{
		Host:        host,
		Port:        port,
		StoreSecret: "somesecret",
	}

	go Run(ctx, cfg)

	time.Sleep(time.Second)

	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestLogin(t *testing.T) {

}
