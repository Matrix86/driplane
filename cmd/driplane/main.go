package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Matrix86/cloudwatcher"
	"github.com/Matrix86/driplane/core"
	"github.com/Matrix86/driplane/utils"

	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/tui"
)

var (
	helpFlag   bool
	debugFlag  bool
	dryRunFlag bool
	rulePath   string
	jsPath     string
	configFile string

	mainOrchestrator *core.Orchestrator
	quitSignal       = false
)

// Signal stops feeders on SIGINT or SIGTERM signal interception
func Signal() {
	sChan := make(chan os.Signal, 1)
	signal.Notify(sChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		s := <-sChan
		switch s {
		case os.Interrupt, syscall.SIGTERM:
			log.Debug("CTRL-C detected")
			if mainOrchestrator != nil {
				quitSignal = true
				mainOrchestrator.StopFeeders()
			}
			return
		}
	}
}

func Update(cfg *core.Configuration) {
	log.Debug("Auto-update enabled")
	s, err := cloudwatcher.New("local", cfg.Get("general.rules_path"), time.Second)
	if err != nil {
		log.Error("AutoUpdate: %s", err)
		return
	}

	config := map[string]string{
		"disable_fsnotify": "false",
	}

	err = s.SetConfig(config)
	if err != nil {
		log.Error("AutoUpdate: %s", err)
		return
	}

	err = s.Start()
	if err != nil {
		log.Error("AutoUpdate: %s\n", err)
		return
	}

	defer s.Close()
	for {
		select {
		case v := <-s.GetEvents():
			log.Info("Event '%s' on File '%s'...restarting", v.TypeString(), v.Key)
			if mainOrchestrator != nil {
				log.Debug("Stopping")
				mainOrchestrator.StopFeeders()
			}

		case e := <-s.GetErrors():
			log.Error("AutoUpdate: %s\n", e)
		}
	}
}

func main() {
	flag.StringVar(&configFile, "config", "", "Set configuration file.")
	flag.StringVar(&rulePath, "rules", "", "Path of the rules' directory.")
	flag.StringVar(&jsPath, "js", "", "Path of the js plugins.")
	flag.BoolVar(&helpFlag, "help", false, "This help.")
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug logs.")
	flag.BoolVar(&dryRunFlag, "dry-run", false, "Only test the rules syntax.")
	flag.Parse()

	appName := fmt.Sprintf("%s v%s", core.Name, core.Version)
	appBuild := fmt.Sprintf("(built for %s %s with %s)", runtime.GOOS, runtime.GOARCH, runtime.Version())
	appAuthor := fmt.Sprintf("Author: %s", core.Author)

	fmt.Printf("%s %s\n%s\n", tui.Bold(appName), tui.Dim(appBuild), tui.Dim(appAuthor))

	if helpFlag {
		flag.Usage()
		return
	}

	if configFile == "" {
		log.Error("you need to set a configuration file")
		flag.Usage()
		return
	}

	log.Output = ""
	log.Level = log.INFO
	log.OnFatal = log.ExitOnFatal
	log.Format = "[{datetime}] {level:color}{level:name}{reset} {message}"

	config, err := core.LoadConfiguration(configFile)
	if err != nil {
		log.Fatal("error loading file '%s': %v", configFile, err)
	}

	if debugFlag || config.Get("general.debug") == "true" {
		log.Level = log.DEBUG
		config.Set("debug", "true")
	}

	if rulePath != "" {
		if !utils.DirExists(rulePath) {
			log.Fatal("rules directory not found: '%s'", rulePath)
		}
		config.Set("general.rules_path", rulePath)
	}

	if config.Get("general.rules_path") == "" {
		log.Error("you need to set up a directory containing the *.rule files using -rules flag or 'rules_path' on the config file")
		return
	}

	if config.Get("general.log_path") != "" {
		log.Output = config.Get("general.log_path")
		if err := log.Open(); err != nil {
			fmt.Printf("log file opening: %v\n", err)
			os.Exit(1)
		}
		defer log.Close()
	}

	if _, err := os.Stat(config.Get("general.rules_path")); os.IsNotExist(err) {
		log.Fatal("rule directory '%s' doesn't exists", config.Get("general.rules_path"))
	}

	if config.Get("debug") == "true" {
		log.Debug("Configurations:")
		for k, v := range config.GetConfig() {
			log.Debug(" %s -> %s", k, v)
		}
	}

	if config.Get("update.enable") == "true" {
		log.Debug("Auto-update enabled")
		go Update(config)
	} else {
		log.Debug("Auto-update disabled")
	}

	go Signal()

	for !quitSignal {
		mainOrchestrator, err = core.NewOrchestrator(config)
		if err != nil {
			log.Fatal("%s", err)
		}

		if dryRunFlag {
			os.Exit(0)
		}

		log.Debug("Trying to start orchestrator")
		mainOrchestrator.StartFeeders()
		mainOrchestrator.WaitFeeders()

		log.Debug("Stopping")
		mainOrchestrator.StopFeeders()
	}
}
