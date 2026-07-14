package protocol

import (
	"context"
	"log"
	"net"
	"time"
)

// PooledAuthenticator 优先从连接池获取连接，失败则回退到内部 authenticator 新建连接。
// 它把连接池复用逻辑纳入 Authenticator IoC，使上层调用方只需依赖 Authenticator 接口。
type PooledAuthenticator struct {
	inner      Authenticator
	connPool   *AccountConnectionPool
	serverAddr string
	httpAddr   string
}

// NewPooledAuthenticator 创建 PooledAuthenticator。
func NewPooledAuthenticator(inner Authenticator, connPool *AccountConnectionPool, serverAddr, httpAddr string) *PooledAuthenticator {
	return &PooledAuthenticator{
		inner:      inner,
		connPool:   connPool,
		serverAddr: serverAddr,
		httpAddr:   httpAddr,
	}
}

// Authenticate 实现 Authenticator 接口。
// 优先尝试从连接池借出；borrow 失败或池不可用时回退到 inner.Authenticate。
// skipDrain=true 时跳过 DrainConn，把帧排空职责交给调用方（如 FrameMux.DrainAndStart）。
func (a *PooledAuthenticator) Authenticate(ctx context.Context, accountID string, skipDrain bool) (net.Conn, bool, uint32, error) {
	if a.connPool != nil {
		pc, poolErr := a.connPool.GetOrCreate(accountID, a.serverAddr, a.httpAddr)
		if poolErr != nil {
			log.Printf("[重放] [account=%s] 连接池获取失败: %v，回退到新建连接", accountID, poolErr)
		} else {
			borrowedConn, seqID := pc.Borrow()
			if borrowedConn != nil {
				if !skipDrain {
					if drainErr := DrainConn(borrowedConn, 100*time.Millisecond); drainErr != nil {
						log.Printf("[重放] [account=%s] DrainConn 失败: %v，关闭脏连接并回退到新建连接", accountID, drainErr)
						a.connPool.Close(accountID)
						// 继续执行下方的新建连接逻辑
					} else {
						log.Printf("[重放] [account=%s] 复用池中连接 (续接 seqID=%d, skipDrain=%v)", accountID, seqID, skipDrain)
						return borrowedConn, true, seqID, nil
					}
				} else {
					log.Printf("[重放] [account=%s] 复用池中连接 (续接 seqID=%d, skipDrain=%v)", accountID, seqID, skipDrain)
					return borrowedConn, true, seqID, nil
				}
			}
		}
	}

	conn, _, seqID, err := a.inner.Authenticate(ctx, accountID, skipDrain)
	return conn, false, seqID, err
}

// Return 归还连接池连接或关闭新建连接。
func (a *PooledAuthenticator) Return(accountID string, conn net.Conn, lastSeqID uint32) {
	if a.connPool != nil {
		a.connPool.Return(accountID, lastSeqID)
		log.Printf("[重放] [account=%s] 连接已归还连接池 (seqID=%d)", accountID, lastSeqID)
	} else {
		a.inner.Return(accountID, conn, lastSeqID)
	}
}
