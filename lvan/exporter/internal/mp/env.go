package mp

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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
		envlogfile := "lvan_mp_logfile"
		_, e := fmt.Fprintf(os.Stderr, "可通过环境变量 %s 设置日志文件：%s\n", envlogfile, "error.log")
		if e != nil {
			log.Fatal(fmt.Errorf("%v, %w", err, e))
			return
		}
		logfile := GetEnvStr(envlogfile, "error.log")
		file, e := os.Create(logfile)
		if e != nil {
			_, e1 := fmt.Fprintf(os.Stderr, "%v\n", e)
			if e1 != nil {
				log.Fatal(fmt.Errorf("%v, %w %w", err, e, e1))
				return
			}
		}
		defer func(file *os.File) {
			e2 := file.Close()
			if e2 != nil {
				log.Fatal(e)
			}
		}(file)
		abs, e := filepath.Abs(file.Name())
		if e != nil {
			log.Fatal(fmt.Errorf("%v, %w", err, e))
			return
		}
		_, e = fmt.Fprintf(os.Stderr, "可通过日志文件查看错误：%s\n", abs)
		if e != nil {
			log.Fatal(fmt.Errorf("%v, %w", err, e))
			return
		}
		writer := io.MultiWriter(file, os.Stderr)
		_, e = fmt.Fprintf(writer, "%v\n", err)
		if e != nil {
			_, e1 := fmt.Fprintf(os.Stderr, "%v\n", e)
			if e1 != nil {
				log.Fatal(fmt.Errorf("%v, %w %w", err, e, e1))
				return
			}
		}
		envsleepseconds := "lvan_mp_error_sleep"
		seconds := GetEnvInt(envsleepseconds, 5)
		_, e = fmt.Fprintf(os.Stderr, "可通过环境变量 %s 控制当前停留时间：%d 秒\n", envsleepseconds, seconds)
		if e != nil {
			log.Fatal(fmt.Errorf("%v, %w", err, e))
			return
		}
		time.Sleep(time.Second * time.Duration(seconds))
		os.Exit(1)
	}
}
