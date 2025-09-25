/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
	load "github.com/wangtengda0310/gobee/lvan/pkg/dump/load/dir"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("import called")
		host, port, user, password, database, table := funcName(cmd)

		// 获取import-dir参数值，优先级：命令行标志 > 配置文件 > 环境变量 > 默认值
		importDir := viper.GetString("import.dir")

		// 检查importDir是否为空
		if importDir == "" {
			log.Panic("import-dir is required")
		}

		records := load.Dir(importDir)

		conn := dump.Conn(host, database, port, user, password)
		conn(func(db *sql.DB) {
			log.Println(database, table)
			dump.Import(db, database, table, records)
		})
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	importCmd.Flags().Bool("help", false, "Help message for import")
	importCmd.Flags().StringP("import-dir", "i", "", "dir will be imported")

	// 将import-dir标志绑定到viper，支持通过配置文件和环境变量设置
	if err := viper.BindPFlag("import.dir", importCmd.Flags().Lookup("import-dir")); err != nil {
		log.Printf("Error binding import-dir flag to viper: %v", err)
	}
}
