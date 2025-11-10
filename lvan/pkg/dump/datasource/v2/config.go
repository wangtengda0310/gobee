package v2

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Config 基础配置接口
// 定义所有数据源配置必须实现的基本方法
type Config interface {
	// GetType 返回数据源类型
	GetType() string

	// Validate 验证配置的有效性
	Validate() error

	// Clone 创建配置的深拷贝
	Clone() Config

	// ToMap 转换为字典格式（用于序列化）
	ToMap() map[string]interface{}
}

// MySQLConfig MySQL数据源配置接口
// 定义MySQL特有的配置方法和参数
type MySQLConfig interface {
	Config

	// GetHost 获取MySQL主机地址
	GetHost() string

	// GetPort 获取MySQL端口号
	GetPort() int

	// GetUser 获取MySQL用户名
	GetUser() string

	// GetPassword 获取MySQL密码
	GetPassword() string

	// GetDatabase 获取数据库名
	GetDatabase() string

	// GetTable 获取表名
	GetTable() string

	// GetCharset 获取字符集（可选）
	GetCharset() string

	// GetTimeout 获取连接超时时间（秒）
	GetTimeout() int

	// GetDSN 生成数据源名称（连接字符串）
	GetDSN() string
}

// RedisConfig Redis数据源配置接口
// 定义Redis特有的配置方法和参数
type RedisConfig interface {
	Config

	// GetHost 获取Redis主机地址
	GetHost() string

	// GetPort 获取Redis端口号
	GetPort() int

	// GetPassword 获取Redis密码
	GetPassword() string

	// GetDatabase 获取Redis数据库索引
	GetDatabase() int

	// GetKeyPattern 获取键匹配模式
	GetKeyPattern() string

	// GetPoolSize 获取连接池大小
	GetPoolSize() int

	// GetTimeout 获取连接超时时间（秒）
	GetTimeout() int
}

// mysqlConfigImpl MySQL配置的具体实现
type mysqlConfigImpl struct {
	host     string
	port     int
	user     string
	password string
	database string
	table    string
	charset  string
	timeout  int
}

// NewMySQLConfig 创建MySQL配置实例
func NewMySQLConfig(host string, port int, user, password, database, table string) MySQLConfig {
	return &mysqlConfigImpl{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		database: database,
		table:    table,
		charset:  "utf8mb4",
		timeout:  30,
	}
}

// GetType 返回数据源类型
func (c *mysqlConfigImpl) GetType() string {
	return "mysql"
}

// Validate 验证配置的有效性
func (c *mysqlConfigImpl) Validate() error {
	var errs []error

	// 验证主机名
	if c.host == "" {
		errs = append(errs, errors.New("主机名不能为空"))
	}

	// 验证端口范围
	if c.port <= 0 || c.port > 65535 {
		errs = append(errs, fmt.Errorf("端口号无效: %d，必须在1-65535之间", c.port))
	}

	// 验证用户名
	if c.user == "" {
		errs = append(errs, errors.New("用户名不能为空"))
	}

	// 验证数据库名
	if c.database == "" {
		errs = append(errs, errors.New("数据库名不能为空"))
	}

	// 验证表名
	if c.table == "" {
		errs = append(errs, errors.New("表名不能为空"))
	}

	// 验证超时时间
	if c.timeout <= 0 {
		errs = append(errs, fmt.Errorf("超时时间无效: %d，必须大于0", c.timeout))
	}

	if len(errs) > 0 {
		return fmt.Errorf("配置验证失败: %v", errs)
	}

	return nil
}

// Clone 创建配置的深拷贝
func (c *mysqlConfigImpl) Clone() Config {
	return &mysqlConfigImpl{
		host:     c.host,
		port:     c.port,
		user:     c.user,
		password: c.password,
		database: c.database,
		table:    c.table,
		charset:  c.charset,
		timeout:  c.timeout,
	}
}

// ToMap 转换为字典格式
func (c *mysqlConfigImpl) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"type":     c.GetType(),
		"host":     c.host,
		"port":     c.port,
		"user":     c.user,
		"password": c.maskPassword(),
		"database": c.database,
		"table":    c.table,
		"charset":  c.charset,
		"timeout":  c.timeout,
	}
}

// maskPassword 密码脱敏处理
func (c *mysqlConfigImpl) maskPassword() string {
	if c.password == "" {
		return ""
	}
	return strings.Repeat("*", min(len(c.password), 8))
}

// GetHost 获取MySQL主机地址
func (c *mysqlConfigImpl) GetHost() string {
	return c.host
}

// GetPort 获取MySQL端口号
func (c *mysqlConfigImpl) GetPort() int {
	return c.port
}

// GetUser 获取MySQL用户名
func (c *mysqlConfigImpl) GetUser() string {
	return c.user
}

// GetPassword 获取MySQL密码
func (c *mysqlConfigImpl) GetPassword() string {
	return c.password
}

// GetDatabase 获取数据库名
func (c *mysqlConfigImpl) GetDatabase() string {
	return c.database
}

// GetTable 获取表名
func (c *mysqlConfigImpl) GetTable() string {
	return c.table
}

// GetCharset 获取字符集
func (c *mysqlConfigImpl) GetCharset() string {
	return c.charset
}

// GetTimeout 获取连接超时时间
func (c *mysqlConfigImpl) GetTimeout() int {
	return c.timeout
}

// GetDSN 生成数据源名称
func (c *mysqlConfigImpl) GetDSN() string {
	// 构建连接参数
	params := url.Values{}

	// 添加字符集参数
	if c.charset != "" {
		params.Add("charset", c.charset)
	}

	// 添加超时参数
	if c.timeout > 0 {
		params.Add("timeout", strconv.Itoa(c.timeout)+"s")
	}

	// 添加解析时间参数
	params.Add("parseTime", "true")

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		c.user,
		url.QueryEscape(c.password),
		c.host,
		c.port,
		c.database)

	// 添加参数
	if len(params) > 0 {
		dsn += "?" + params.Encode()
	}

	return dsn
}

// redisConfigImpl Redis配置的具体实现
type redisConfigImpl struct {
	host       string
	port       int
	password   string
	database   int
	keyPattern string
	poolSize   int
	timeout    int
}

// NewRedisConfig 创建Redis配置实例
func NewRedisConfig(host string, port int, password string, database int, keyPattern string) RedisConfig {
	return &redisConfigImpl{
		host:       host,
		port:       port,
		password:   password,
		database:   database,
		keyPattern: keyPattern,
		poolSize:   10,
		timeout:    30,
	}
}

// GetType 返回数据源类型
func (c *redisConfigImpl) GetType() string {
	return "redis"
}

// Validate 验证配置的有效性
func (c *redisConfigImpl) Validate() error {
	var errs []error

	// 验证主机名
	if c.host == "" {
		errs = append(errs, errors.New("主机名不能为空"))
	}

	// 验证端口范围
	if c.port <= 0 || c.port > 65535 {
		errs = append(errs, fmt.Errorf("端口号无效: %d，必须在1-65535之间", c.port))
	}

	// 验证数据库索引
	if c.database < 0 || c.database > 15 {
		errs = append(errs, fmt.Errorf("Redis数据库索引无效: %d，必须在0-15之间", c.database))
	}

	// 验证键模式
	if c.keyPattern == "" {
		errs = append(errs, errors.New("键匹配模式不能为空"))
	}

	// 验证连接池大小
	if c.poolSize <= 0 {
		errs = append(errs, fmt.Errorf("连接池大小无效: %d，必须大于0", c.poolSize))
	}

	// 验证超时时间
	if c.timeout <= 0 {
		errs = append(errs, fmt.Errorf("超时时间无效: %d，必须大于0", c.timeout))
	}

	if len(errs) > 0 {
		return fmt.Errorf("配置验证失败: %v", errs)
	}

	return nil
}

// Clone 创建配置的深拷贝
func (c *redisConfigImpl) Clone() Config {
	return &redisConfigImpl{
		host:       c.host,
		port:       c.port,
		password:   c.password,
		database:   c.database,
		keyPattern: c.keyPattern,
		poolSize:   c.poolSize,
		timeout:    c.timeout,
	}
}

// ToMap 转换为字典格式
func (c *redisConfigImpl) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"type":        c.GetType(),
		"host":        c.host,
		"port":        c.port,
		"password":    c.maskPassword(),
		"database":    c.database,
		"keyPattern":  c.keyPattern,
		"poolSize":    c.poolSize,
		"timeout":     c.timeout,
	}
}

// maskPassword 密码脱敏处理
func (c *redisConfigImpl) maskPassword() string {
	if c.password == "" {
		return ""
	}
	return strings.Repeat("*", min(len(c.password), 8))
}

// GetHost 获取Redis主机地址
func (c *redisConfigImpl) GetHost() string {
	return c.host
}

// GetPort 获取Redis端口号
func (c *redisConfigImpl) GetPort() int {
	return c.port
}

// GetPassword 获取Redis密码
func (c *redisConfigImpl) GetPassword() string {
	return c.password
}

// GetDatabase 获取Redis数据库索引
func (c *redisConfigImpl) GetDatabase() int {
	return c.database
}

// GetKeyPattern 获取键匹配模式
func (c *redisConfigImpl) GetKeyPattern() string {
	return c.keyPattern
}

// GetPoolSize 获取连接池大小
func (c *redisConfigImpl) GetPoolSize() int {
	return c.poolSize
}

// GetTimeout 获取连接超时时间
func (c *redisConfigImpl) GetTimeout() int {
	return c.timeout
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}