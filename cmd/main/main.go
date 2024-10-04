package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	resource "demo/resource"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

// golang Http 服务仅能使用绝对路径
func main() {
	staticDir := http.FileServer(http.Dir("E:\\work\\go\\demo\\resource"))
	embedFs := http.FileServer(http.FS(resource.StaticResouces))
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(loggingMiddleware(staticDir))
	router.PathPrefix("/version").HandlerFunc(GetEnvVersion)
	router.PathPrefix("/healthz").HandlerFunc(Healthz)
	router.PathPrefix("/").Handler(loggingMiddleware(embedFs))

	var errChan chan (error)
	var server http.Server
	go func() {
		server := http.Server{Addr: ":8080", Handler: router}
		err := server.ListenAndServe()
		errChan <- err
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Main goroutine is waiting for signal...")

	var err error
	var sig os.Signal
	select {
	case err = <-errChan:
		log.Printf("Received starting web signal: %v\n", err)
	case sig = <-c:
		log.Printf("Received signal: %v\n", sig)
		log.Printf("%s web service is exiting... \n", "demo APP")
		log.Println("Cleaning up...")
		// 这里可以执行一些清理工作，比如关闭文件、释放资源等
		ctx, cf := context.WithTimeout(context.Background(), time.Second)
		defer cf()
		server.Shutdown(ctx) // 优雅关闭http服务实例
		log.Printf("%s program exit ok\n", "demo APP")
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func GetEnvVersion(resp http.ResponseWriter, req *http.Request) {
	processHeader(resp, req)
	str := fmt.Sprintf("Server VERSION: %s \nServer GOPATH: %s", os.Getenv("VERSION"), os.Getenv("GOPATH"))
	io.WriteString(resp, str)
}

func Healthz(resp http.ResponseWriter, req *http.Request) {
	processHeader(resp, req)
	str := "200"
	resp.Write([]byte(str))
}

func processHeader(resp http.ResponseWriter, req *http.Request) {
	defer glog.Flush()
	header := req.Header
	remoteAddr := req.RemoteAddr
	// status := req.Response.Status
	if header != nil {
		for k, v := range header {
			for _, vv := range v {
				resp.Header().Set(k, vv)
			}
		}
	}

	data, err := json.Marshal(header)
	if err == nil {
		defer log.Printf("ClientIP: %s ResponseStatus: %s header json : %s", remoteAddr, "200", string(data))
	}
}
