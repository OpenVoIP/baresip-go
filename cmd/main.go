package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/OpenVoIP/baresip-go/ctrltcp"
	"github.com/OpenVoIP/baresip-go/pkg/httpserver"
)

var (
	gitCommitCode string
	buildDateTime string
	goVersion     string
)

var enableHTTP bool

func main() {
	var (
		host    = flag.String("host", "127.0.0.1", "Server host (valid values: 0.0.0.0)")
		port    = flag.Int("port", 4444, "TCP port")
		debug   = flag.Bool("debug", false, "Log")
		version = flag.Bool("version", false, "Version")
	)
	flag.Parse()

	if *version {
		fmt.Printf("gitCommitCode: %s, buildDateTime: %s %s", gitCommitCode, buildDateTime, goVersion)
		return
	}

	if *debug {
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
		log.Info("log level DEBUG")
	} else {
		log.Info("log level INFO")
	}

	if enableHTTP {
		go httpserver.CreateServer()
	}

	ctrltcp.ConnectRedis()
	go ctrltcp.HandControlAction()
	info := ctrltcp.InitConn()

	for {
		info.Connect(*host, *port)
		time.Sleep(3 * time.Second)
	}
}
