/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package exec

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wangtengda0310/gobee/lvan/cmd/dumper/cmd"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

var command *exec.Cmd

// execCmd represents the execCmd command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		log.Println("PersistentPreRunE exec")

		c := dbParams(cmd)
		log.Println("config", c)

		Accept(func(visitor Visitor, where string, args ...string) {
			var arguments []string
			arguments = append(arguments, "--skip-ssl")
			if c.Host != "" {
				arguments = append(arguments, "-h", c.Host)
			}
			if c.Port != 3306 {
				arguments = append(arguments, "-P", fmt.Sprintf("%d", c.Port))
			}
			arguments = append(arguments, "-u", c.User)
			if c.Password != "" {
				arguments = append(arguments, "-p", c.Password)
			}
			arguments = append(arguments, c.Database)

			command = exec.Command("mariadb", arguments...)

			visitor(command)

		})
		fmt.Println("exec called", c)
		return nil
	},
}

func init() {
	cmd.RootCmd.AddCommand(execCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// execCmd.PersistentFlags().String("foo", "", "A help for foo")

	// 将cmd和args标志设置为持久标志，这样它们就可以应用于exec命令及其所有子命令
	// 这样用户就可以在命令行中使用 `exec --cmd echo --args "1 2 3" import` 这样的顺序

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// execCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	execCmd.PersistentFlags().BoolP("help", "?", false, "Help message for exec")
	flags(execCmd)
}

func flags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("host", "h", "localhost", "MySQL 主机名 (默认: localhost)")
	cmd.PersistentFlags().Uint16P("port", "P", 3306, "MySQL 端口号 (默认: 3306)")
	cmd.PersistentFlags().StringP("user", "u", "root", "MySQL 用户名 (默认: root)")
	cmd.PersistentFlags().StringP("password", "p", "", "MySQL 密码 (空密码时可不提供)")
	cmd.PersistentFlags().StringP("database", "d", "gforge", "数据库名 (默认: gforge)")
	cmd.PersistentFlags().StringP("table", "t", "user", "表名 (默认: user)")

	if viper.BindPFlag("host", cmd.PersistentFlags().Lookup("host")) != nil {
		return
	}
	if viper.BindPFlag("port", cmd.PersistentFlags().Lookup("port")) != nil {
		return
	}
	if viper.BindPFlag("user", cmd.PersistentFlags().Lookup("user")) != nil {
		return
	}
	if viper.BindPFlag("password", cmd.PersistentFlags().Lookup("password")) != nil {
		return
	}
	if viper.BindPFlag("database", cmd.PersistentFlags().Lookup("database")) != nil {
		return
	}
	if viper.BindPFlag("table", cmd.PersistentFlags().Lookup("table")) != nil {
		return
	}
}

func dbParams(cmd *cobra.Command) (c dump.Config) {

	err := viper.Unmarshal(&c)
	if err != nil {
		log.Panic(err)
	}
	return
}
