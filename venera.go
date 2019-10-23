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

	"racoondev.tk/gitea/racoon/venera/internal/bot"
	"racoondev.tk/gitea/racoon/venera/internal/dispatcher"
	"racoondev.tk/gitea/racoon/venera/internal/storage"
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

	database, err := storage.Connect(utils.Configuration.GetConnectionString())
	if err != nil {
		logger.Critical(err)
		os.Exit(1)
	}

	wgBot := sync.WaitGroup{}
	ctx, shutdownBot := context.WithCancel(context.Background())
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt, os.Kill)

	if err := dispatcher.Initialize(logger, database); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}

	err = bot.Initialize(ctx, logger, &wgBot, utils.Configuration.Telegram.Token,
		utils.Configuration.Telegram.TrustedUser)
	if err != nil {
		logger.Critical(err)
		os.Exit(1)
	}

	wgDispatcher := sync.WaitGroup{}
	dispatcher.RunServer(logger, &wgDispatcher)

	<-signalChannel

	shutdownBot()
	wgBot.Wait()

	dispatcher.Stop()
	wgDispatcher.Wait()

}
