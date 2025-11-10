package v2

// Datasource 数据源基础接口
type Datasource interface {
    Accept(visitor Visitor)
    GetMetadata() Metadata
}

// Visitor 访问者基础接口
type Visitor interface {
    VisitDatasource(ds Datasource)
}

// DataVisitor 数据特定访问者接口，支持特定数据源类型
type DataVisitor interface {
    Visitor
    VisitMySQL(ds MySQLDatasource)
    VisitRedis(ds RedisDatasource)
}

// MySQLDatasource MySQL数据源接口
type MySQLDatasource interface {
    Datasource
    GetConnection() interface{} // 现在返回真实的*sql.DB
    GetDatabase() string
    GetTable() string
    GetHost() string
    GetPort() int
    GetUser() string
    GetPassword() string
    IsConnected() bool // 检查连接状态
    Close() error     // 关闭连接
}

// RedisDatasource Redis数据源接口
type RedisDatasource interface {
    Datasource
    GetClient() interface{} // 临时返回interface{}，后续完善
    GetKeyPattern() string
}

// Metadata 元数据接口
type Metadata interface {
    GetType() string
    GetTables() []string
}