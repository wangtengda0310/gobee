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
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/redis"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
)

// redisCmd represents the redis command
var redisCmd = &cobra.Command{
	Use:   "redis",
	Short: "Redis 数据源操作",
	Long:  `Redis 数据源的导入导出操作`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		parent := cmd.Parent()
		err := parent.PersistentPreRunE(parent, args)
		if err != nil {
			log.Panic(err)
			return err
		}
		log.Println("PersistentPreRunE redis")

		// 解析 Redis 配置
		c := redisParams(cmd)
		log.Println("config", c)

		// 转换为 service.Config
		svcCfg := service.Config{
			Host:     c.Host,
			Port:     c.Port,
			Password: c.Password,
		}

		// 创建 Redis 管理器
		mgr, err := service.NewRedisManager(context.Background(), svcCfg)
		if err != nil {
			return fmt.Errorf("创建 Redis 连接失败: %w", err)
		}

		// 获取当前 context，如果为 nil 则使用 Background
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		// 存储到 context
		ctx = cmdcontext.SetManager(ctx, mgr)
		cmd.SetContext(ctx)

		log.Println("Redis 数据源已初始化")
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// 清理：关闭连接
		if mgr := cmdcontext.GetManager(cmd.Context()); mgr != nil {
			if err := mgr.Close(); err != nil {
				log.Printf("关闭 Redis 连接失败: %v", err)
				return err
			}
			log.Println("Redis 连接已关闭")
		}
		return nil
	},
}

// RedisParams Redis 连接参数
type RedisParams struct {
	Host     string
	Port     uint16
	Password string
	DB       int
}

func redisParams(cmd *cobra.Command) RedisParams {
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetUint16("port")
	password, _ := cmd.Flags().GetString("auth")
	db, _ := cmd.Flags().GetInt("db")

	return RedisParams{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
	}
}

// redisDumpCmd Redis dump 子命令
var redisDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "导出 Redis 数据",
	Long:  `导出 Redis 数据到 ZIP 文件`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("redis dump command called")

		// 获取参数
		pattern, _ := cmd.Flags().GetString("pattern")
		output, _ := cmd.Flags().GetString("output")

		log.Printf("Pattern: %s, Output: %s", pattern, output)

		// 从 context 获取 manager
		mgr := cmdcontext.GetManager(cmd.Context())
		if mgr == nil {
			log.Panic("Redis 数据源未初始化")
		}

		// 类型断言获取 RedisManager
		redisMgr, ok := mgr.(*service.RedisManager)
		if !ok {
			log.Panic("不是 Redis 管理器")
		}

		// 创建导出器
		dumper := redis.NewDumper(redisMgr)

		// 执行导出
		ctx := context.Background()
		result, err := dumper.Dump(ctx, pattern, output)
		if err != nil {
			log.Fatalf("导出失败: %v", err)
		}

		log.Printf("导出成功: %d 个 key -> %s", result.KeysExported, result.FileName)
	},
}

// redisImportCmd Redis import 子命令
var redisImportCmd = &cobra.Command{
	Use:   "import",
	Short: "导入 Redis 数据",
	Long:  `从 ZIP 文件导入 Redis 数据`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("redis import command called")
		inputFile := args[0]
		log.Printf("Input file: %s", inputFile)

		// 获取目标 DB 参数
		db, _ := cmd.Flags().GetInt("db")
		log.Printf("Target DB: %d", db)

		// 从 context 获取 manager
		mgr := cmdcontext.GetManager(cmd.Context())
		if mgr == nil {
			log.Panic("Redis 数据源未初始化")
		}

		// 类型断言获取 RedisManager
		redisMgr, ok := mgr.(*service.RedisManager)
		if !ok {
			log.Panic("不是 Redis 管理器")
		}

		// 创建导入器
		importer := redis.NewImporter(redisMgr)

		// 执行导入
		ctx := context.Background()
		result, err := importer.Import(ctx, inputFile, db)
		if err != nil {
			log.Fatalf("导入失败: %v", err)
		}

		log.Printf("导入成功: %d 个 key <- %s", result.KeysImported, result.FileName)
	},
}

func init() {
	rootCmd.AddCommand(redisCmd)

	// Redis Persistent Flags - 对所有子命令可用
	// 注意：不使用 -h 作为 host 的简写，因为 -h 被 help 占用
	redisCmd.PersistentFlags().String("host", "127.0.0.1", "Redis 主机名 (默认: 127.0.0.1)")
	redisCmd.PersistentFlags().Uint16P("port", "p", 6379, "Redis 端口号 (默认: 6379)")
	redisCmd.PersistentFlags().StringP("auth", "a", "", "Redis 密码")
	redisCmd.PersistentFlags().IntP("db", "n", 0, "Redis 数据库索引 (默认: 0)")

	// 绑定到 viper
	viper.BindPFlag("redis.host", redisCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("redis.port", redisCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("redis.auth", redisCmd.PersistentFlags().Lookup("auth"))
	viper.BindPFlag("redis.db", redisCmd.PersistentFlags().Lookup("db"))

	// dump 命令的 flags
	redisDumpCmd.Flags().StringP("pattern", "P", "*", "Redis Key 匹配模式 (默认: *)")
	redisDumpCmd.Flags().StringP("output", "o", "", "输出 ZIP 文件路径")

	// 绑定到 viper
	viper.BindPFlag("redis.dump.pattern", redisDumpCmd.Flags().Lookup("pattern"))
	viper.BindPFlag("redis.dump.output", redisDumpCmd.Flags().Lookup("output"))

	// import 命令的 flags
	redisImportCmd.Flags().IntP("db", "n", 0, "目标数据库索引 (默认: 0)")

	// 绑定到 viper
	viper.BindPFlag("redis.import.db", redisImportCmd.Flags().Lookup("db"))

	// 添加子命令
	redisCmd.AddCommand(redisDumpCmd)
	redisCmd.AddCommand(redisImportCmd)
}
