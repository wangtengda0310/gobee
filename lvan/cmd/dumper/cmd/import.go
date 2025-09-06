/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/wangtengda0310/gobee/lvan/cmd/dumper/cmd/internal"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/write"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "导入数据",
	Long:  `connect mysql import`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Println("PersistentPreRun import")
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		log.Println("PersistentPreRunE import")
		parent := cmd.Parent()
		err := parent.PersistentPreRunE(parent, args)
		if err != nil {
			log.Panic(err)
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("import called")

		c := dbParams(cmd)
		database := c.Database
		table := c.Table
		log.Println("config", c)

		internal.VisitImport(func(db dump.Datasource) {

			for _, input := range args {

				// 检查input是否为空
				if input == "" {
					log.Panic("import-dir is required")
				}

				var records []dump.Record
				records = internal.TransImport(input)

				write.Console(records)

				log.Println(database, table)
				dump.Import(db.DB, database, table, records)
			}

		}, "", "")
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
