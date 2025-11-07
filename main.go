/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-spring/spring-core/gs"
)

var test = Test{}

func init() {
	gs.Object(&Service{}).AsServer()
	gs.Object(&test)
	gs.Object(&Config{Host: "localhost", Port: 8080, User: "root", Password: "password", Database: "test"})

	gs.Property("start-time", time.Now().Format(timeLayout))
	gs.Property("refresh-time", time.Now().Format(timeLayout))
}

func (s *Service) ListenAndServe(sig gs.ReadySignal) error {
	log.Println(s.test)
	return nil
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}

func main() {
	gs.EnableServers(false)
	gs.EnableJobs(false)
	gs.EnableSimpleHttpServer(false)
	gs.EnableSimplePProfServer(false)
	gs.Web(false)
	async, err := gs.RunAsync()
	if err != nil {
		log.Println(err)
	}
	log.Println(test)
	//gs.Run()
	defer async()
}

const timeLayout = "2006-01-02 15:04:05.999 -0700 MST"

type Service struct {
	gs.Server
	test        *Test              `autowire:""`
	StartTime   time.Time          `value:"${start-time}"`
	RefreshTime gs.Dync[time.Time] `value:"${refresh-time}"`
}

func (s *Service) Echo(w http.ResponseWriter, r *http.Request) {
	str := fmt.Sprintf("start-time: %s refresh-time: %s",
		s.StartTime.Format(timeLayout),
		s.RefreshTime.Value().Format(timeLayout))
	w.Write([]byte(str))
}

func (s *Service) Refresh(w http.ResponseWriter, r *http.Request) {
	gs.Property("refresh-time", time.Now().Format(timeLayout))
	gs.RefreshProperties()
	w.Write([]byte("OK!"))
}

type Config struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Database string
	Table    string
}

type Test struct {
	c *Config `autowire:""`
}

var T = Test{}

const versionKey = "app.version"

type App struct {
	Version gs.Dync[string] `value:"${app.version:=v0.0.1}"`
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Version:", a.Version.Value())
}

func RefreshVersion(w http.ResponseWriter, r *http.Request) {
	gs.Property(versionKey, "v2.0.1")
	gs.RefreshProperties()
	fmt.Fprintln(w, "Version updated!")
}
