package config

import (
	"fmt"
	// "log"
	"os"

	"github.com/namsral/flag"
	"github.com/outbrain/golib/log"
)

//CLIOptions defines the base configuration that can be passed to the WASTE system via CLI
type CLIOptions struct {
	DatabaseName      string
	OriginalTableName string
	DBUser            string
	DBPasswd          string
	Artifact          string
	AlterStatement    string
	OrcUsername       string
	OrcPasswd         string
	GithubToken       string
	GithubOwner       string
	GithubRepo        string
	WebAddress        string
	Execute           bool
	KVStoreAddress    string
	KVStoreUser       string
	KVStorePassword   string
}

var Config CLIOptions

func (clio *CLIOptions) ReadArgs() {
	flag.String(flag.DefaultConfigFlagname, "", "path to config file")
	flag.StringVar(&clio.AlterStatement, "artifact", "", "full artifact (mandatory)")
	flag.StringVar(&clio.DBUser, "DBUser", "", "MySQL user")
	flag.StringVar(&clio.DBPasswd, "DBPasswd", "", "MySQL password")
	flag.StringVar(&clio.DatabaseName, "database", "", "database name (mandatory)")
	flag.StringVar(&clio.OriginalTableName, "table", "", "table name (mandatory)")
	flag.StringVar(&clio.AlterStatement, "alter", "", "alter statement (mandatory)")
	flag.StringVar(&clio.OrcUsername, "OrcUsername", "", "Orchestrator username")
	flag.StringVar(&clio.OrcPasswd, "OrcPasswd", "", "Orchestrator password")
	flag.StringVar(&clio.GithubToken, "GithubToken", "", "Github Token")
	flag.StringVar(&clio.GithubOwner, "GithubOwner", "", "Github Owner")
	flag.StringVar(&clio.GithubRepo, "GithubRepo", "", "Github Repo")
	flag.StringVar(&clio.WebAddress, "WebAddress", "", "address for the web API")
	flag.StringVar(&clio.KVStoreAddress, "KVStoreAddress", "", "Github Repo")
	flag.StringVar(&clio.KVStoreUser, "KVStoreUser", "", "Github Repo")
	flag.StringVar(&clio.KVStorePassword, "KVStorePassword", "", "Github Repo")
	execute := flag.Bool("execute", false, "should we execute")
	quiet := flag.Bool("quiet", false, "quiet")
	verbose := flag.Bool("verbose", false, "verbose")
	debug := flag.Bool("debug", false, "debug mode (very verbose)")
	stack := flag.Bool("stack", false, "add stack trace upon error")
	help := flag.Bool("help", false, "Display usage")
	flag.CommandLine.SetOutput(os.Stdout)
	clio.Execute = *execute

	flag.Parse()

	if *help {
		fmt.Println("W.A.S.T.E stands for What Artifact Schema Transformer Etc")
		fmt.Fprintf(os.Stdout, "Usage of waste:\n")
		flag.PrintDefaults()
		return
	}

	log.SetLevel(log.INFO)
	if *verbose {
		log.SetLevel(log.INFO)
	}
	if *debug {
		log.SetLevel(log.DEBUG)
	}
	if *stack {
		log.SetPrintStackTrace(*stack)
	}
	if *quiet {
		// Override!!
		log.SetLevel(log.ERROR)
	}
}
