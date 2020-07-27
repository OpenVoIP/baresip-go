package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/OpenVoIP/baresip-go/ctrltcp"
)

func main() {
	var (
		gitCommitCode string
		buildDateTime string
		goVersion     string
	)

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

	ctrltcp.ConnectRedis()
	go ctrltcp.HandControlAction()
	info := ctrltcp.InitConn()

	for {
		info.Connect(*host, *port)
		time.Sleep(3 * time.Second)
	}
}
