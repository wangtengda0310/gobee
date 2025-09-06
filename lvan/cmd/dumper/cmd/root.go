/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wangtengda0310/gobee/lvan/cmd/dumper/cmd/internal"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/load"
)

const (
	in        = "in"
	tpzip     = "zip"
	tpsql     = "sql-tpl"
	tpdir     = "dir"
	tpconsole = "-"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "connect",
	Short: "外网数据克隆工具",
	Long: `外网数据克隆工具 支持数据库的dump和import操作:

	connect mysql dump	导出玩家数据
	connect mysql import	导入玩家数据`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Println("PersistentPreRun root")
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		log.Println("PersistentPreRunE root")
		internal.TransImport = transImportFormat()
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Run root")
	},
}

func transImportFormat() func(string) []dump.Record {
	in := viper.GetString(in)
	switch in {
	case tpzip:
		return load.Zip
	case tpdir:
		return load.Dir
	case tpsql:
		return func(sqlTemplateFile string) []dump.Record {
			// 解析模板
			// 将新生成的sql写入文件
			sqlFile := "sqlTemplateFile" + ".sql"
			return []dump.Record{map[string][]byte{"sql": []byte(sqlFile)}}
		}
	case tpconsole:
		return load.Console
	default:
		return load.Zip
	}
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
	rootCmd.PersistentFlags().String(in, tpzip, fmt.Sprintf("输出format:%s/%s/%s (默认: %s)", tpzip, tpdir, tpsql, tpzip))
	if viper.BindPFlag(in, rootCmd.PersistentFlags().Lookup(in)) != nil {
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
