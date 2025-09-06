/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wangtengda0310/gobee/lvan/cmd/dumper/cmd/internal"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

// mysqlCmd represents the mysql command
var mysqlCmd = &cobra.Command{
	Use:   "mysql",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		parent := cmd.Parent()
		err := parent.PersistentPreRunE(parent, args)
		if err != nil {
			log.Panic(err)
			return err
		}
		log.Println("PersistentPreRunE mysql")

		c := dbParams(cmd)
		log.Println("config", c)

		internal.Accept(func(visitor internal.Visitor, where string, args ...string) {
			conn := dump.ConnC(c)
			conn(func(db dump.Datasource) {

				visitor(db)

				log.Println("done")

			})
		})
		fmt.Println("dump called", c)

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func dbParams(cmd *cobra.Command) (c dump.Config) {

	err := viper.Unmarshal(&c)
	if err != nil {
		log.Panic(err)
	}
	return
}

func init() {
	rootCmd.AddCommand(mysqlCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mysqlCmd.PersistentFlags().String("foo", "", "A help for foo")
	persistentFlags(mysqlCmd)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mysqlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	mysqlCmd.Flags().BoolP("help", "?", false, "Help message for mysql")
	flags(mysqlCmd)
}

func persistentFlags(mysqlCmd *cobra.Command) {

}

func flags(mysqlCmd *cobra.Command) {
	mysqlCmd.Flags().StringP("host", "h", "localhost", "MySQL 主机名 (默认: localhost)")
	mysqlCmd.Flags().Uint16P("port", "P", 3306, "MySQL 端口号 (默认: 3306)")
	mysqlCmd.Flags().StringP("user", "u", "root", "MySQL 用户名 (默认: root)")
	mysqlCmd.Flags().StringP("password", "p", "", "MySQL 密码 (空密码时可不提供)")
	mysqlCmd.Flags().StringP("database", "d", "gforge", "数据库名 (默认: gforge)")
	mysqlCmd.Flags().StringP("table", "t", "user", "表名 (默认: user)")

	if viper.BindPFlag("host", mysqlCmd.Flags().Lookup("host")) != nil {
		return
	}
	if viper.BindPFlag("port", mysqlCmd.Flags().Lookup("port")) != nil {
		return
	}
	if viper.BindPFlag("user", mysqlCmd.Flags().Lookup("user")) != nil {
		return
	}
	if viper.BindPFlag("password", mysqlCmd.Flags().Lookup("password")) != nil {
		return
	}
	if viper.BindPFlag("database", mysqlCmd.Flags().Lookup("database")) != nil {
		return
	}
	if viper.BindPFlag("table", mysqlCmd.Flags().Lookup("table")) != nil {
		return
	}
}
