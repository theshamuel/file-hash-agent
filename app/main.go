package main

import (
	"fmt"
	"github.com/hashicorp/logutils"
	"github.com/jessevdk/go-flags"
	"github.com/theshamuel/file-hash-agent/app/fileagent"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var opts struct {
	Path        string        `long:"path" env:"FILEPATH" required:"true" description:"URL for pulling when checksum update is detected"`
	URL         string        `long:"url" env:"URL" required:"true" description:"URL for pulling when checksum update is detected"`
	Interval    time.Duration `long:"interval" env:"INTERVAL" default:"600s" description:"Interval calculating checksum (default 600sec)"`
	Delay       time.Duration `long:"delay" env:"DELAY"  default:"60s" description:"startup delay for calculating checksum (default 60sec)"`
	Debug       bool          `long:"debug" env:"DEBUG" description:"debug mode"`
	Credentials struct {
		Login    string `long:"login" env:"LOGIN" description:"login to access to remote url"`
		Password string `long:"password" env:"PASSWORD" description:"password to access to remote url"`
		Token    string `long:"token" env:"TOKEN" description:"token to access to remote url"`
	} `group:"credentials" namespace:"credentials" env-namespace:"CREDENTIALS"`
	Config struct {
		Enabled  bool   `long:"enabled" env:"ENABLED" description:"enable getting parameters from config. In that case all parameters will be read only form config"`
		FileName string `long:"file-name" env:"FILE_NAME" default:"file-hash-agent.yml" description:"config file name"`
	} `group:"config" namespace:"config" env-namespace:"CONFIG"`
}

var version = "unknown"

func main() {
	fmt.Printf("[INFO]: File Hash Agent version %s\n", version)
	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	p.SubcommandsOptional = true
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			log.Printf("[ERROR] cli error: %v", err)
		}
		os.Exit(2)
	}

	setupLogLevel(opts.Debug)

	log.Printf("[DEBUG] options: %+v", opts)

	fa := fileagent.FileAgent{
		Path:     opts.Path,
		Interval: opts.Interval,
		Delay:    opts.Delay,
	}

	if err := fa.Run(); err != nil {
		log.Fatalf("[ERROR] proxy server failed, %v", err)
	}

}
func setupLogLevel(debug bool) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   os.Stdout,
	}
	log.SetFlags(log.Ldate | log.Ltime)

	if debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
		filter.MinLevel = logutils.LogLevel("DEBUG")
	}
	log.SetOutput(filter)
}

func getStackTrace() string {
	maxSize := 7 * 1024 * 1024
	stacktrace := make([]byte, maxSize)
	length := runtime.Stack(stacktrace, true)
	if length > maxSize {
		length = maxSize
	}
	return string(stacktrace[:length])
}

func init() {
	sigChan := make(chan os.Signal)
	go func() {
		for range sigChan {
			log.Printf("[INFO] Singal QUITE is cought , stacktrace [\n%s", getStackTrace())
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)
}
