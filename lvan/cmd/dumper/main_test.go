package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func getUserTimestamps(host, username, password string, uid int) (time.Time, time.Time, error) {
	// 构建数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/gforge_u01_alpha2", username, password, host)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// 查询语句
	query := "SELECT `ctime`, `mtime` FROM `user` WHERE `uid`=?"

	// 执行查询
	var ctime, mtime time.Time
	err = db.QueryRow(query, uid).Scan(&ctime, mtime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, time.Time{}, fmt.Errorf("no user found with uid %d", uid)
		}
		return time.Time{}, time.Time{}, fmt.Errorf("query failed: %w", err)
	}

	return ctime, mtime, nil
}

func TestSql(t *testing.T) {
	host := "101.34.211.79:32533"
	username := "root"
	password := "p_mysql"
	uid := 14

	ctime, mtime, err := getUserTimestamps(host, username, password, uid)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("ctime: %v, mtime: %v\n", ctime, mtime)
}
