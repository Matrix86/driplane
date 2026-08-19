package main

import (
	"context"
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
	"github.com/Matrix86/driplane/web"

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
	webFlag    bool
	webAddress string

	mainSupervisor *core.Supervisor
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
			if mainSupervisor != nil {
				mainSupervisor.Stop()
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
			if mainSupervisor != nil {
				mainSupervisor.Reload()
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
	flag.BoolVar(&webFlag, "web", false, "Enable the embedded web interface.")
	flag.StringVar(&webAddress, "web-address", "", "Address of the embedded web interface (default 127.0.0.1:8080). Setting this enables the interface on its own.")
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

	if webFlag {
		config.Set("web.enable", "true")
	}
	if webAddress != "" {
		config.Set("web.address", webAddress)
		config.Set("web.enable", "true")
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

	// -dry-run must exit here, before the debug configuration dump below and
	// before web.Start further down. Its only job is to parse the rules and
	// exit; it must not print the daemon's configuration or bind the web
	// interface's port. Both matter beyond tidiness: the web UI's /api/test
	// handler runs a dry-run of this same binary, with this same config
	// file, as a subprocess, and returns its combined output verbatim to the
	// browser. If this check ran after the debug dump, then with
	// general.debug: true (an ordinary development setup) every /api/test
	// call would print every config value — including feeder credentials
	// such as twitter.consumerKey — into the editor's test panel. If it ran
	// after web.Start, every /api/test call would also spawn a second web
	// server that fails to bind the already-listening port (or, with
	// web.token left empty, prints a freshly generated token into that same
	// panel). Keep this check first, before both.
	if dryRunFlag {
		if _, err := core.NewOrchestrator(config); err != nil {
			log.Error("%s", err)
			os.Exit(1)
		}
		log.Info("rules syntax OK")
		os.Exit(0)
	}

	if config.Get("debug") == "true" {
		log.Debug("Configurations:")
		for k, v := range config.GetConfig() {
			log.Debug(" %s -> %s", k, v)
		}
	}

	mainSupervisor = core.NewSupervisor(config)

	// web.Start must run here, before any other goroutine that logs is
	// started (go Signal() and the optional go Update() below): it installs
	// log.Callback by assigning a package-level variable in islazy/log that
	// is not synchronised against the logger's own internal lock, so doing
	// it while another goroutine could be logging concurrently would be a
	// data race. At this point in main() the only goroutine running is
	// main() itself, so the assignment is safe. It must still come after
	// log.Open() above, so the startup line with the URL and token goes to
	// the configured log destination.
	webServer, err := web.Start(mainSupervisor)
	if err != nil {
		log.Fatal("web interface: %s", err)
	}
	if webServer != nil {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := webServer.Shutdown(ctx); err != nil {
				log.Debug("web shutdown: %s", err)
			}
		}()
	}

	go Signal()

	if config.Get("update.enable") == "true" {
		log.Debug("Auto-update enabled")
		go Update(config)
	} else {
		log.Debug("Auto-update disabled")
	}

	if err := mainSupervisor.Run(); err != nil {
		log.Error("%s", err)
		os.Exit(1)
	}
}
