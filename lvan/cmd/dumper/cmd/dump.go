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
	"github.com/wangtengda0310/gobee/lvan/cmd/dumper/cmd/cmdcontext"
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

		// 从 context 获取 manager
		mgr := cmdcontext.GetManager(cmd.Context())
		if mgr == nil {
			log.Panic("数据源未初始化")
		}

		db := mgr.GetDB()
		cfg := mgr.GetConfig()

		columns := dump.Columns(db, cfg.Database, cfg.Table)
		columnTypes := dump.GetColumnTypes(db, cfg.Database, cfg.Table)
		records := dump.Dump(db, cfg.Database, cfg.Table, columns, columnTypes, where, args...)
		log.Println(len(records), "records")

		pkColumns, err := dump.GetPrimaryKeyColumns(db, cfg.Database, cfg.Table)
		if err != nil {
			log.Printf("获取主键失败: %v\n", err)
		}
		log.Println("主键", pkColumns)
		export := internal.TransExport(records, pkColumns...)

		log.Println("exported format", export)
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
	case tprredis:
		// Redis 输出格式 (MySQL → Redis 迁移)
		return _type.Redis(_type.RedisConfig{
			Host:      viper.GetString("redis-host"),
			Port:      viper.GetInt("redis-port"),
			Password:  viper.GetString("redis-password"),
			DB:        viper.GetInt("redis-db"),
			KeyPrefix: viper.GetString("redis-key-prefix"),
			TTL:       viper.GetInt64("redis-ttl"),
		})
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

	// Redis 输出相关 flags (MySQL → Redis 迁移)
	dumpCmd.PersistentFlags().String("redis-host", "127.0.0.1", "Redis 主机地址")
	dumpCmd.PersistentFlags().Int("redis-port", 6379, "Redis 端口")
	dumpCmd.PersistentFlags().String("redis-password", "", "Redis 密码")
	dumpCmd.PersistentFlags().Int("redis-db", 0, "Redis 数据库编号")
	dumpCmd.PersistentFlags().String("redis-key-prefix", "export:", "Redis key 前缀")
	dumpCmd.PersistentFlags().Int64("redis-ttl", 0, "Redis key 过期时间（秒），0 表示永不过期")

	// 绑定 Redis flags 到 viper
	viper.BindPFlag("redis-host", dumpCmd.PersistentFlags().Lookup("redis-host"))
	viper.BindPFlag("redis-port", dumpCmd.PersistentFlags().Lookup("redis-port"))
	viper.BindPFlag("redis-password", dumpCmd.PersistentFlags().Lookup("redis-password"))
	viper.BindPFlag("redis-db", dumpCmd.PersistentFlags().Lookup("redis-db"))
	viper.BindPFlag("redis-key-prefix", dumpCmd.PersistentFlags().Lookup("redis-key-prefix"))
	viper.BindPFlag("redis-ttl", dumpCmd.PersistentFlags().Lookup("redis-ttl"))

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dumpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	dumpCmd.Flags().BoolP("help", "?", false, "Help message for dump")
	dumpCmd.Flags().Bool("simulate", false, "模拟模式，不实际连接数据库")
	dumpCmd.Flags().StringP("where", "w", "uid", "查询列 (默认: uid)")
}
