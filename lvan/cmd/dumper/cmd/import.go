/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/wangtengda0310/gobee/lvan/cmd/dumper/cmd/cmdcontext"
	"github.com/wangtengda0310/gobee/lvan/cmd/dumper/cmd/internal"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/write"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "导入数据",
	Long:  `connect mysql import`,
	// 移除 PersistentPreRunE，避免重复初始化
	// mysql 命令已经处理了连接初始化
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("import called")

		// 从 context 获取 manager
		mgr := cmdcontext.GetManager(cmd.Context())
		if mgr == nil {
			log.Panic("数据源未初始化")
		}

		db := mgr.GetDB()
		cfg := mgr.GetConfig()

		log.Println("config", cfg)

		for _, input := range args {
			// 检查input是否为空
			if input == "" {
				log.Panic("import-dir is required")
			}

			var records []dump.Record
			records = internal.TransImport(input)

			write.Console(records)

			log.Println(cfg.Database, cfg.Table)
			dump.Import(db, cfg.Database, cfg.Table, records)
		}
	},
}

func init() {
	//redisCmd.AddCommand(importCmd)
	mysqlCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	importCmd.Flags().BoolP("help", "?", false, "Help message for import")

}
