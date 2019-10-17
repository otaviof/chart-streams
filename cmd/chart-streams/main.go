package main

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const appName = "chart-streams"

// rootCmd main command.
var rootCmd = &cobra.Command{
	Use: appName,
}

// init initialize the command-line flags and interpolation with environment.
func init() {
	flags := rootCmd.PersistentFlags()

	viper.SetEnvPrefix(appName)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	bindViperFlags(flags)
}

// bindViperFlags based on flag-set, creating a environment variable equivalent with Viper.
func bindViperFlags(flags *pflag.FlagSet) {
	if err := viper.BindPFlags(flags); err != nil {
		log.Fatal(err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
