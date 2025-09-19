package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-spring/spring-core/gs"
)

func main() { gs.Run() }
func init() {
	gs.Object(&Service{})

	gs.Provide(func(s *Service) http.Handler {
		http.HandleFunc("/echo", s.Echo)
		http.HandleFunc("/refresh", s.Refresh)
		return http.DefaultServeMux
	})

	gs.Property("start-time", time.Now().Format(timeLayout))
	gs.Property("refresh-time", time.Now().Format(timeLayout))
}

const timeLayout = "2006-01-02 15:04:05.999 -0700 MST"

type Service struct {
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
