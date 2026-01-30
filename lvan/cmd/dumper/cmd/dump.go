/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wangtengda0310/gobee/lvan/cmd/dumper/cmd/internal"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
	_type "github.com/wangtengda0310/gobee/lvan/pkg/dump/type"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/write"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "导出数据",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		internal.TransExport = transExportFormat()
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("dump command called")
		where, err := cmd.Flags().GetString("where")
		if err != nil {
			log.Panic(err)
		}
		internal.VisitExport(func(db dump.Datasource) {

			columns := dump.Columns(db.DB, db.Database, db.Table)
			columnTypes := dump.GetColumnTypes(db.DB, db.Database, db.Table)
			records := dump.Dump(db.DB, db.Database, db.Table, columns, columnTypes, where, args...)
			log.Println(len(records), "records")

			pkColumns, err := dump.GetPrimaryKeyColumns(db.DB, db.Database, db.Table)
			if err != nil {
				log.Printf("获取主键失败: %v\n", err)
			}
			log.Println("主键", pkColumns)
			export := internal.TransExport(records, pkColumns...)

			log.Println("exported format", export)
		}, where, args...)

	},
}

func transExportFormat() func(records []dump.Record, pks ...string) string {
	output := viper.GetString("output")
	in := viper.GetString(in)
	switch in {
	case tpzip:
		// 从 viper 获取 database 和 table 名称
		database := viper.GetString("database")
		table := viper.GetString("table")
		// 如果用户指定了输出路径，使用用户指定的路径；否则使用默认的 database.table.zip
		filename := output
		if filename == "" {
			filename = fmt.Sprintf("%s.%s.zip", database, table)
		}
		return _type.Zip(filename)
	//case tpdir:
	//	database := viper.GetString("database")
	//	table := viper.GetString("table")
	//	dirname := output
	//	if dirname == "" {
	//		dirname = fmt.Sprintf("%s.%s", database, table)
	//	}
	//	return _type.Dir(dirname)
	case tpsql:
		return func(records []dump.Record, pks ...string) string {
			return fmt.Sprintf("%v", records)
		}
	case tpconsole:
		return write.Console
	default:
		return write.Console
	}
}

func init() {
	//redisCmd.AddCommand(dumpCmd)
	mysqlCmd.AddCommand(dumpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dumpCmd.PersistentFlags().String("foo", "", "A help for foo")
	dumpCmd.PersistentFlags().String("to", "", "--output的别名")
	dumpCmd.PersistentFlags().StringP("output", "o", "", "数据输出到哪里 (默认: zip文件)")

	// 将output标志绑定到viper，支持通过配置文件和环境变量设置
	if err := viper.BindPFlag("output", dumpCmd.PersistentFlags().Lookup("output")); err != nil {
		if err := viper.BindPFlag("output", dumpCmd.Flags().Lookup("to")); err != nil {
			log.Printf("Error binding output flag to viper: %v", err)
		}
	}

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dumpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	dumpCmd.Flags().BoolP("help", "?", false, "Help message for dump")
	dumpCmd.Flags().Bool("simulate", false, "模拟模式，不实际连接数据库")
	dumpCmd.Flags().StringP("where", "w", "uid", "查询列 (默认: uid)")
}
