package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//Configuration defines the base configuration that can be passed to the WASTE system
type Configuration struct {
	Debug       bool
	Execute     bool
	DBUser      string
	DBPasswd    string
	GithubToken string
	GithubOwner string
	GithubRepo  string
	WebAddress  string
}

// Config is the global configuration variable
var Config = loadConfiguration()

// loadConfiguration loads configuration using viper
func loadConfiguration() *Configuration {

	viper.SetDefault("Debug", true)
	viper.SetDefault("Execute", false)
	viper.SetDefault("WebAddress", "localhost:4000")

	viper.SetConfigName("waste.conf")   // name of config file (without extension)
	viper.AddConfigPath("/etc/waste/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.waste") // call multiple times to add many search paths
	viper.AddConfigPath("./conf")       // optionally look for config in the working directory
	err := viper.ReadInConfig()         // Find and read the config file
	if err != nil {                     // Handle errors reading the config file
		log.Error().Err(err).Msg("Fatal error config file")
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Info().Msgf("Config file changed: %v", e.Name)
	})
	var cfg Configuration
	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Error().Err(err).Msg("unable to decode into struct")
	}

	log.Info().Msgf("configuration loaded: %+v", cfg)

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return &cfg
}
