package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Matrix86/driplane/core"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
)

var (
	helpFlag   bool
	rulePath   string
	configFile string
)

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
	flag.StringVar(&configFile, "c", "", "Set configuration file.")
	flag.StringVar(&rulePath, "r", "", "Path of the rules' directory.")
	flag.BoolVar(&helpFlag, "h", false, "This help.")
	flag.Parse()

	if helpFlag {
		flag.Usage()
		return
	}

	if configFile == "" {
		log.Error("you need to set a configuration file")
		flag.Usage()
		return
	}

	if rulePath == "" {
		log.Error("you need to set up a directory containing the *.rule files")
		flag.Usage()
		return
	}

	log.Output = ""
	log.Level = log.INFO
	log.OnFatal = log.ExitOnFatal
	log.Format = "[{datetime}] {level:color}{level:name}{reset} {message}"

	config, err := core.LoadConfiguration(configFile)
	if err != nil {
		log.Fatal("configuration file not found: %s", err)
	}

	log.Level = config.GetLogLevel()
	log.Output = config.LogPath
	if err := log.Open(); err != nil {
		panic(err)
	}
	defer log.Close()

	if _, err := os.Stat(rulePath); os.IsNotExist(err) {
		log.Fatal("rule directory '%s' doesn't exists", rulePath)
	}

	parser, _ := core.NewParser()

	ruleAsts := make(map[string]*core.AST)
	err = fs.Glob(rulePath, "*.rule", func(file string) error {
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
