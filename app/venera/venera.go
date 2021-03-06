package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"

	"github.com/ccding/go-logging/logging"

	"github.com/racoon-devel/venera/internal/bot"
	"github.com/racoon-devel/venera/internal/dispatcher"
	"github.com/racoon-devel/venera/internal/storage"
	"github.com/racoon-devel/venera/internal/utils"
)

const version = "1.3"

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

	if err := storage.Connect(utils.Configuration.GetConnectionString()); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}

	wgBot := sync.WaitGroup{}
	ctx, shutdownBot := context.WithCancel(context.Background())
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt, os.Kill)

	if err := dispatcher.Initialize(logger); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}

	err := bot.Initialize(ctx, logger, &wgBot, utils.Configuration.Telegram.Token,
		utils.Configuration.Telegram.TrustedUser)
	if err != nil {
		logger.Critical(err)
		os.Exit(1)
	}

	wgDispatcher := sync.WaitGroup{}
	dispatcher.RunServer(logger, &wgDispatcher)

	<-signalChannel

	logger.Info("Dispatcher shutdowning...")
	dispatcher.Stop()
	wgDispatcher.Wait()
	logger.Info("Dispatcher shutdowned")

	logger.Info("Bot shutdowning...")
	shutdownBot()
	wgBot.Wait()
	logger.Info("Bot shutdowned")
}
