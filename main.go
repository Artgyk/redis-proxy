package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)



func main() {
	httpServer := NewHttpServer(HttpServerConfigFromEnv())
	srv := listenAndServer(httpServer.Router)
	srv.RegisterOnShutdown(httpServer.Close)

	tcpServer:= NewTcpServer(httpServer.RedisClient, httpServer.Cache)
	tcpListener:=tcpServer.Serve(env("TCP_LISTEN_PORT", ""))

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown Server ...")

	tcpListener.Close()

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("Server Shutdown: %q", err.Error())
	}

	log.Println("Server exiting")
}


func listenAndServer(router *gin.Engine) *http.Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", env("HTTP_LISTEN_PORT", "")),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return srv
}