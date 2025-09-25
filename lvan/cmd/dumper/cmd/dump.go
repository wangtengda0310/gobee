/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
	write "github.com/wangtengda0310/gobee/lvan/pkg/dump/write/dir"

	"github.com/spf13/cobra"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		host, port, user, password, database, table := funcName(cmd)
		conn := dump.Conn(host, database, port, user, password)
		conn(func(db *sql.DB) {
			log.Println("ping error", db.Ping())
			where, err := cmd.Flags().GetString("where")
			if err != nil {
				log.Panic(err)
			}

			columns := dump.Columns(db, database, table)
			records := dump.Dump(db, database, table, columns, where, args...)
			log.Println(len(records), "records")

			pkColumns, err := dump.PetPrimaryKeyColumns(db, database, table)
			if err != nil {
				log.Printf("获取主键失败: %v\n", err)
			}
			log.Println("主键", pkColumns)

			write.Dir(records, fmt.Sprintf("%s.%s", database, table), pkColumns...)
			log.Println("done")
		})
		fmt.Println("dump called", host, user, password, database, table)
	},
}

func funcName(cmd *cobra.Command) (host string, port uint16, user, password, database, table string) {
	host = viper.GetString("host")
	port = viper.GetUint16("port")
	user = viper.GetString("user")
	password = viper.GetString("password")
	database = viper.GetString("database")
	table = viper.GetString("table")
	return
}

func init() {
	rootCmd.AddCommand(dumpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dumpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dumpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	dumpCmd.Flags().BoolP("help", "?", false, "Help message for dump")
	dumpCmd.Flags().StringP("where", "w", "uid", "表名 (默认: user)")
	dumpCmd.Flags().Bool("simulate", false, "模拟模式，不实际连接数据库")
}
