package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/ccding/go-logging/logging"

	"racoondev.tk/gitea/racoon/venera/internal/dispatcher"
	"racoondev.tk/gitea/racoon/venera/internal/server"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

const version = "0.1"

func main() {
	fmt.Printf("Venera Project v%s\n", version)

	runtime.GOMAXPROCS(runtime.NumCPU())

	logger, _ := logging.WriterLogger("venera", logging.INFO, "%12s [%s][%7s:%3d] %s\n time,levelname,filename,lineno,message",
		"15:04:05.999", os.Stdout, true)

	defer logger.Destroy()

	configPath := flag.String("config", "/etc/venera/venera.conf",
		"set server configuration file")

	verbose := flag.Bool("verbose", false, "print debug information to log")

	flag.Parse()

	if *verbose {
		logger.SetLevel(logging.DEBUG)
	}

	logger.Infof("configuration file '%s' used", *configPath)

	if err := utils.Configuration.Load(*configPath); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}

	logger.Debug(utils.Configuration)

	if err := dispatcher.Init(logger); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}

	server.Run(logger)
}
