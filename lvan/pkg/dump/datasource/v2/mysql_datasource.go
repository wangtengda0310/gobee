package v2

import (
    "database/sql"
    "fmt"
    "log"
    "runtime"
    "sync"
    "time"

    _ "github.com/go-sql-driver/mysql"
)

// MySQLDatasourceImpl MySQL数据源实现
type MySQLDatasourceImpl struct {
    config       MySQLConfig   // 使用类型安全的配置接口
    metadata     Metadata      // 数据源元数据
    conn         *sql.DB       // 数据库连接
    connMutex    sync.RWMutex  // 连接读写锁
    isConnected  bool          // 连接状态
    lastPingTime time.Time     // 最后一次ping时间
    pingInterval time.Duration // ping间隔
}

// NewMySQLDatasource 创建MySQL数据源实例
func NewMySQLDatasource(config MySQLConfig) MySQLDatasource {
    // 验证配置
    if err := config.Validate(); err != nil {
        panic(fmt.Sprintf("MySQL配置验证失败: %v", err))
    }

    ds := &MySQLDatasourceImpl{
        config:       config,
        metadata:     &TestMetadata{data: "mysql"},
        pingInterval: 30 * time.Second, // 默认30秒ping一次
    }

    // 初始化数据库连接
    if err := ds.initConnection(); err != nil {
        panic(fmt.Sprintf("初始化MySQL连接失败: %v", err))
    }

    // 设置终结器，当对象被垃圾回收时自动关闭资源
    runtime.SetFinalizer(ds, func(ds *MySQLDatasourceImpl) {
        ds.cleanup()
    })

    return ds
}

// initConnection 初始化数据库连接
func (ds *MySQLDatasourceImpl) initConnection() error {
    // 构建DSN
    dsn := ds.config.GetDSN()

    // 打开数据库连接
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return fmt.Errorf("打开MySQL连接失败: %v", err)
    }

    // 设置连接池参数
    db.SetMaxOpenConns(25)                 // 最大开放连接数
    db.SetMaxIdleConns(5)                  // 最大空闲连接数
    db.SetConnMaxLifetime(5 * time.Minute) // 连接最大生存时间
    db.SetConnMaxIdleTime(1 * time.Minute) // 空闲连接最大生存时间

    // 验证连接
    if err := db.Ping(); err != nil {
        db.Close()
        return fmt.Errorf("MySQL连接验证失败: %v", err)
    }

    ds.connMutex.Lock()
    ds.conn = db
    ds.isConnected = true
    ds.lastPingTime = time.Now()
    ds.connMutex.Unlock()

    log.Printf("成功连接到MySQL: %s:%d/%s", ds.config.GetHost(), ds.config.GetPort(), ds.config.GetDatabase())
    return nil
}

// cleanup 清理资源
func (ds *MySQLDatasourceImpl) cleanup() {
    ds.connMutex.Lock()
    defer ds.connMutex.Unlock()

    if ds.conn != nil {
        if err := ds.conn.Close(); err != nil {
            log.Printf("关闭MySQL连接时出错: %v", err)
        } else {
            log.Printf("MySQL连接已关闭: %s:%d/%s", ds.config.GetHost(), ds.config.GetPort(), ds.config.GetDatabase())
        }
        ds.conn = nil
        ds.isConnected = false
    }
}

// checkConnection 检查连接状态并在需要时重新连接
func (ds *MySQLDatasourceImpl) checkConnection() error {
    ds.connMutex.RLock()

    // 如果连接不存在或未连接，尝试重新连接
    if ds.conn == nil || !ds.isConnected {
        ds.connMutex.RUnlock()
        return ds.reconnect()
    }

    // 检查是否需要ping
    if time.Since(ds.lastPingTime) > ds.pingInterval {
        ds.connMutex.RUnlock()

        // 执行ping检查
        if err := ds.ping(); err != nil {
            log.Printf("MySQL ping失败，尝试重新连接: %v", err)
            return ds.reconnect()
        }
        return nil
    }

    ds.connMutex.RUnlock()
    return nil
}

// ping 执行数据库ping检查
func (ds *MySQLDatasourceImpl) ping() error {
    ds.connMutex.Lock()
    defer ds.connMutex.Unlock()

    if ds.conn == nil {
        return fmt.Errorf("数据库连接为空")
    }

    if err := ds.conn.Ping(); err != nil {
        ds.isConnected = false
        return fmt.Errorf("数据库ping失败: %v", err)
    }

    ds.isConnected = true
    ds.lastPingTime = time.Now()
    return nil
}

// reconnect 重新连接数据库
func (ds *MySQLDatasourceImpl) reconnect() error {
    ds.connMutex.Lock()
    defer ds.connMutex.Unlock()

    // 关闭现有连接
    if ds.conn != nil {
        ds.conn.Close()
        ds.conn = nil
    }

    // 重新初始化连接
    return ds.initConnection()
}

// getConnection 获取数据库连接（带健康检查）
func (ds *MySQLDatasourceImpl) getConnection() (*sql.DB, error) {
    if err := ds.checkConnection(); err != nil {
        return nil, fmt.Errorf("获取数据库连接失败: %v", err)
    }

    ds.connMutex.RLock()
    defer ds.connMutex.RUnlock()

    if !ds.isConnected {
        return nil, fmt.Errorf("数据库未连接")
    }

    return ds.conn, nil
}

// Accept 实现访问者模式的Accept方法
func (ds *MySQLDatasourceImpl) Accept(visitor Visitor) {
    // 首先尝试数据特定访问者
    if dataVisitor, ok := visitor.(DataVisitor); ok {
        dataVisitor.VisitMySQL(ds)
        return
    }

    // 回退到基础访问者
    visitor.VisitDatasource(ds)
}

// GetMetadata 获取数据源元数据
func (ds *MySQLDatasourceImpl) GetMetadata() Metadata {
    return ds.metadata
}

// GetConnection 获取数据库连接
func (ds *MySQLDatasourceImpl) GetConnection() interface{} {
    conn, err := ds.getConnection()
    if err != nil {
        panic(fmt.Sprintf("获取数据库连接失败: %v", err))
    }
    return conn
}

// GetDatabase 获取数据库名
func (ds *MySQLDatasourceImpl) GetDatabase() string {
    return ds.config.GetDatabase()
}

// GetTable 获取表名
func (ds *MySQLDatasourceImpl) GetTable() string {
    return ds.config.GetTable()
}

// GetHost 获取主机名
func (ds *MySQLDatasourceImpl) GetHost() string {
    return ds.config.GetHost()
}

// GetPort 获取端口号
func (ds *MySQLDatasourceImpl) GetPort() int {
    return ds.config.GetPort()
}

// GetUser 获取用户名
func (ds *MySQLDatasourceImpl) GetUser() string {
    return ds.config.GetUser()
}

// GetPassword 获取密码
func (ds *MySQLDatasourceImpl) GetPassword() string {
    return ds.config.GetPassword()
}

// IsConnected 检查连接状态
func (ds *MySQLDatasourceImpl) IsConnected() bool {
    ds.connMutex.RLock()
    defer ds.connMutex.RUnlock()
    return ds.isConnected && ds.conn != nil
}

// Close 关闭数据库连接
func (ds *MySQLDatasourceImpl) Close() error {
    ds.cleanup()
    return nil
}

// GetConnectionStats 获取连接池统计信息
func (ds *MySQLDatasourceImpl) GetConnectionStats() map[string]interface{} {
    ds.connMutex.RLock()
    defer ds.connMutex.RUnlock()

    if ds.conn == nil {
        return map[string]interface{}{
            "connected": false,
            "message":   "连接未初始化",
        }
    }

    stats := ds.conn.Stats()
    return map[string]interface{}{
        "connected":           ds.isConnected,
        "open_connections":    stats.OpenConnections,
        "in_use":             stats.InUse,
        "idle":               stats.Idle,
        "wait_count":         stats.WaitCount,
        "wait_duration":      stats.WaitDuration.String(),
        "max_idle_closed":    stats.MaxIdleClosed,
        "max_lifetime_closed": stats.MaxLifetimeClosed,
        "last_ping":          ds.lastPingTime.Format(time.RFC3339),
        "ping_interval":      ds.pingInterval.String(),
    }
}

// SetPingInterval 设置ping间隔
func (ds *MySQLDatasourceImpl) SetPingInterval(interval time.Duration) {
    ds.connMutex.Lock()
    defer ds.connMutex.Unlock()
    ds.pingInterval = interval
}

// GetPingInterval 获取ping间隔
func (ds *MySQLDatasourceImpl) GetPingInterval() time.Duration {
    ds.connMutex.RLock()
    defer ds.connMutex.RUnlock()
    return ds.pingInterval
}


// 测试用的元数据实现
type TestMetadata struct {
    data string
}

func (m *TestMetadata) GetType() string {
    return m.data
}

func (m *TestMetadata) GetTables() []string {
    // 临时实现，后续完善
    return []string{"user"}
}