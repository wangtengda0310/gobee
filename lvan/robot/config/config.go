package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"time"
)

type RobotMode uint

const (
	_           = iota // 常量初始值
	SimpleRobot        // 普通机器人，运行一遍回放
	PressRobot         // 压测机器人，无限运行回放
)

var (
	configFile string               // 配置文件名
	conf       = newDefaultConfig() // 创建默认配置对象
)

func initParse() {
	// 替换config层flag
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.Usage = flag.Usage

	flag.UintVar(&conf.ZoneID, "zone-id", conf.ZoneID, "区服ID")
	flag.StringVar(&configFile, "conf", "./robot.yaml", "配置文件路径")
	flag.UintVar(&conf.RobotCount, "count", conf.RobotCount, "机器人数量")
	flag.UintVar(&conf.Interval, "interval", conf.Interval, "消息调用间隔, 单位: ms")
	flag.StringVar(&conf.AccountFile, "accounts", "", "配置登录账号文件")
}

// GetZoneID 获取区服ID
func GetZoneID() uint32 { return uint32(conf.ZoneID) }

// GetLoginUrl login服务器地址
func GetLoginUrl() string { return GetConfig().LoginUrl }

// GetGateUrl gate服地址
func GetGateUrl() string { return GetConfig().GateUrl }

func GetInterval() time.Duration {
	n := GetConfig().Interval
	if n < 10 {
		n = 10
	}
	diff := float32(n) * 0.2
	i := float32(n) - diff + rand.Float32()*2*diff
	return time.Duration(i) * time.Millisecond
}

func GetConfig() *Config { return conf }

type NamePWD struct {
	Name string // 用户名
	PWD  string // 用户密码
}
type Config struct {
	ZoneID        uint       // 区服ID
	LoginUrl      string     // login地址
	GateUrl       string     // gate地址
	RobotCount    uint       // 机器人数量,如果数量为0则根据登录账号数量创建机器人
	PerRobotCount uint       // 每秒钟创建机器人数量
	Interval      uint       // 消息发送间隔(单位: ms)
	ClientVer     string     // 客户端版本号,默认:"1.0.0"
	AccountFile   string     // 登录账号文件,如果此配置为空,注册新账号
	Accounts      []*NamePWD // 登录账号密码列表
	Mode          RobotMode  // 机器人模式 1:普通机器人，2:压测机器人
}

func newDefaultConfig() *Config {
	return &Config{
		ZoneID:        1,
		LoginUrl:      "",
		GateUrl:       "",
		RobotCount:    1,
		PerRobotCount: 100,
		Interval:      100,
		ClientVer:     "1.0.0",
		Mode:          SimpleRobot,
	}
}

func (c *Config) Process() error {
	if err := c.parseUserPWD(); err != nil {
		return err
	}
	return nil
}

// parseUserPWD 加载登录用户名密码
func (c *Config) parseUserPWD() error {
	if len(c.AccountFile) == 0 {
		return nil
	}

	f, err := os.Open(c.AccountFile)
	if err != nil {
		return fmt.Errorf("open account failed, path:%s error: %w", c.AccountFile, err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	accounts := make([]*NamePWD, 0)

	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " ")

		if len(line) == 0 {
			continue
		}

		sp := strings.Split(line, ":")

		if len(sp) != 2 {
			logger.Error("read illegal account", "content", line, "path", c.AccountFile, "error", err)
			continue
		}

		accounts = append(accounts, &NamePWD{
			Name: sp[0],
			PWD:  sp[1],
		})
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("scan account failed, path:%s error: %w", c.AccountFile, err)
	}

	rand.Shuffle(len(accounts), func(i, j int) {
		accounts[i], accounts[j] = accounts[j], accounts[i]
	})

	c.Accounts = accounts

	return nil
}

func (c *Config) GetAccount(index uint) *NamePWD {
	if int(index) >= len(c.Accounts) {
		return nil
	}
	return c.Accounts[index]
}
