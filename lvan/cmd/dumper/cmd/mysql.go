/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wangtengda0310/gobee/lvan/cmd/dumper/cmd/cmdcontext"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
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

		// 解析配置
		c := dbParams(cmd)
		log.Println("config", c)

		// 转换为 service.Config
		svcCfg := service.Config{
			Host:     c.Host,
			Port:     c.Port,
			User:     c.User,
			Password: c.Password,
			Database: c.Database,
			Table:    c.Table,
		}

		// 创建 MySQL 管理器
		mgr, err := service.NewMySQLManager(context.Background(), svcCfg)
		if err != nil {
			return fmt.Errorf("创建数据源失败: %w", err)
		}

		// 获取当前 context，如果为 nil 则使用 Background
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		// 存储到 context
		ctx = cmdcontext.SetManager(ctx, mgr)
		cmd.SetContext(ctx)

		log.Println("数据源已初始化")
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// 清理：关闭连接
		if mgr := cmdcontext.GetManager(cmd.Context()); mgr != nil {
			if err := mgr.Close(); err != nil {
				log.Printf("关闭数据源失败: %v", err)
				return err
			}
			log.Println("数据源已关闭")
		}
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
	// 将数据库连接参数定义为 Persistent Flags，使其对所有子命令可用
	mysqlCmd.PersistentFlags().StringP("host", "h", "localhost", "MySQL 主机名 (默认: localhost)")
	mysqlCmd.PersistentFlags().Uint16P("port", "P", 3306, "MySQL 端口号 (默认: 3306)")
	mysqlCmd.PersistentFlags().StringP("user", "u", "root", "MySQL 用户名 (默认: root)")
	mysqlCmd.PersistentFlags().StringP("password", "p", "", "MySQL 密码 (空密码时可不提供)")
	mysqlCmd.PersistentFlags().StringP("database", "d", "gforge", "数据库名 (默认: gforge)")
	mysqlCmd.PersistentFlags().StringP("table", "t", "user", "表名 (默认: user)")
}

func flags(mysqlCmd *cobra.Command) {
	// 本地 flag 可以在这里定义
	// 注意：数据库连接参数已移至 persistentFlags()

	// 绑定 persistent flags 到 viper
	if viper.BindPFlag("host", mysqlCmd.PersistentFlags().Lookup("host")) != nil {
		return
	}
	if viper.BindPFlag("port", mysqlCmd.PersistentFlags().Lookup("port")) != nil {
		return
	}
	if viper.BindPFlag("user", mysqlCmd.PersistentFlags().Lookup("user")) != nil {
		return
	}
	if viper.BindPFlag("password", mysqlCmd.PersistentFlags().Lookup("password")) != nil {
		return
	}
	if viper.BindPFlag("database", mysqlCmd.PersistentFlags().Lookup("database")) != nil {
		return
	}
	if viper.BindPFlag("table", mysqlCmd.PersistentFlags().Lookup("table")) != nil {
		return
	}
}
