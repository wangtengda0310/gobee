package mp

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func GetEnvStr(key string, defaultValue string) string {
	getenv := os.Getenv(key)
	if getenv == "" {
		return defaultValue
	}
	return getenv
}
func GetEnvInt(key string, defaultValue int) int {
	getenv := os.Getenv(key)
	if getenv == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(getenv)
	if err != nil {
		return defaultValue
	}
	return intValue
}
func Recover() {
	if err := recover(); err != nil {
		logfile := GetEnvStr("lvan_mp_logfile", "error.log")
		fmt.Fprintf(os.Stderr, "查看日志文件：%s\n", logfile)
		file, e := os.Create(logfile)
		if e != nil {
			fmt.Fprintf(os.Stderr, "%v\n", e)
		}
		defer file.Close()
		_, e = fmt.Fprintf(file, "%v\n", err)
		if e != nil {
			fmt.Fprintf(os.Stderr, "%v\n", e)
		}
		seconds := GetEnvInt("lvan_mp_error_sleep", 5)
		time.Sleep(time.Second * time.Duration(seconds))
		os.Exit(1)
	}
}
