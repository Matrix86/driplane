package main

import (
	"flag"
	"fmt"
	"github.com/Matrix86/driplane/utils"
	"github.com/evilsocket/islazy/tui"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/Matrix86/driplane/core"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
)

var (
	helpFlag   bool
	debugFlag  bool
	rulePath   string
	jsPath     string
	configFile string
)

// Signal stops feeders on SIGINT or SIGTERM signal interception
func Signal(o *core.Orchestrator) {

	sChan := make(chan os.Signal, 1)
	signal.Notify(sChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		s := <-sChan
		switch s {
		case os.Interrupt, syscall.SIGTERM:
			log.Debug("CTRL-C detected")
			o.StopFeeders()
			return
		}
	}
}

func main() {
	flag.StringVar(&configFile, "config", "", "Set configuration file.")
	flag.StringVar(&rulePath, "rules", "", "Path of the rules' directory.")
	flag.StringVar(&jsPath, "js", "", "Path of the js plugins.")
	flag.BoolVar(&helpFlag, "help", false, "This help.")
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug logs.")
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
		if utils.DirExists(rulePath) == false {
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
			log.Fatal("log file opening: %v", err)
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

	parser, _ := core.NewParser()

	ruleAsts := make(map[string]*core.AST)
	err = fs.Glob(config.Get("general.rules_path"), "*.rule", func(file string) error {
		ast, err := parser.ParseFile(file)
		if err != nil {
			log.Fatal("rule parsing: %s", err)
		}
		ruleAsts[file] = ast
		return nil
	})
	if err != nil {
		log.Fatal("rule directory enumeration: %s", err)
	}

	o, err := core.NewOrchestrator(ruleAsts, config)
	if err != nil {
		log.Fatal("Error %s", err)
	}

	go Signal(&o)

	log.Debug("Trying to start orchestrator")
	o.StartFeeders()

	//c := make(chan os.Signal)
	//signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	//fmt.Println(<-c)

	o.WaitFeeders()

	log.Debug("Stopping")
	o.StopFeeders()

	return
}
