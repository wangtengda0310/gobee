/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/go-spring/spring-core/gs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var gsShutdown func()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dumper",
	Short: "外网数据克隆工具",
	Long: `外网数据克隆工具 支持数据库的dump和import操作:

	dumper dump	导出玩家数据
	dumper import	导入玩家数据`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		gsShutdown()
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		gs.EnableServers(false)
		gs.EnableJobs(false)
		gs.EnableSimpleHttpServer(false)
		gs.EnableSimplePProfServer(false)
		gs.Web(false)
		async, err := gs.RunAsync()
		gsShutdown = async
		log.Println("start go-spring")
		return err
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dumper.yaml)")
	rootCmd.PersistentFlags().StringP("host", "h", "localhost", "MySQL 主机名 (默认: localhost)")
	rootCmd.PersistentFlags().Uint16P("port", "P", 3306, "MySQL 端口号 (默认: 3306)")
	rootCmd.PersistentFlags().StringP("user", "u", "root", "MySQL 用户名 (默认: root)")
	rootCmd.PersistentFlags().StringP("password", "p", "", "MySQL 密码 (空密码时可不提供)")
	rootCmd.PersistentFlags().StringP("database", "d", "gforge", "数据库名 (默认: gforge)")
	rootCmd.PersistentFlags().StringP("table", "t", "user", "表名 (默认: user)")

	if viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host")) != nil {
		return
	}
	if viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port")) != nil {
		return
	}
	if viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("user")) != nil {
		return
	}
	if viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password")) != nil {
		return
	}
	if viper.BindPFlag("database", rootCmd.PersistentFlags().Lookup("database")) != nil {
		return
	}
	if viper.BindPFlag("table", rootCmd.PersistentFlags().Lookup("table")) != nil {
		return
	}

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("help", "?", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".dumper" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".dumper")
	}

	viper.SetEnvPrefix("dump")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
