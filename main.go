package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"yourtube/controllers"
	"yourtube/sqldb"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dsnRef := flag.String("dsn", "", "Required. The Data Sournce Name of the PostgreSQL instance the server will connect to")
	portRef := flag.String("port", "", "Optional. The port the server will listen on. Defaults to '8080'. Cannot be used in conjunction with the 'socket' flag")
	socketRef := flag.String("socket", "", "Optional. The path to the Unix socket the server will listen on. Falls back to 'port' if not provided. Cannot be used in conjunction with the 'port' flag")
	releaseRef := flag.Bool("release", false, "Optional. If provided, runs the server in 'release' mode. Defaults to 'false'")
	flag.Parse()

	dsn := *dsnRef
	port := *portRef
	socket := *socketRef
	release := *releaseRef

	if dsn == "" {
		panic("Must provide a value for 'dsn'")
	}

	if port != "" && socket != "" {
		panic("Cannot set both 'port' and 'socket' arguments")
	}

	if port == "" && socket == "" {
		port = "8080"
	}

	if release {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	router.SetTrustedProxies(nil)

	sqldb.InitDb(dsn)

	// Auth
	router.POST("/authenticate")

	// Global video management
	router.GET("/videos", controllers.GetVideos)
	router.GET("/video/:video_id", controllers.GetVideo)
	router.PUT("/video/:video_id", controllers.PutVideo)

	// Global channel management
	router.GET("/channels", controllers.GetChannels)
	router.GET("/channel/:channel_id", controllers.GetChannel)
	router.PUT("/channel/:channel_id", controllers.PutChannel)

	// User management
	router.POST("/user")
	router.GET("/user/:user_id")
	router.PUT("/user/:user_id")

	// User parent/child management
	router.GET("/user/:user_id/parent")
	router.PUT("/user/:user_id/parent/:parent_id")
	router.DELETE("/user/:user_id/parent")
	router.GET("/user/:user_id/children")

	// User-level channel management
	router.GET("/user/:user_id/channels")
	router.GET("/user/:user_id/channel/:channel_id")
	router.PUT("/user/:user_id/channel/:channel_id")
	router.DELETE("/user/:user_id/channel/:channel_id")

	// User-level available videos
	router.GET("/user/:user_id/videos", controllers.GetUserVideos)
	router.GET("/user/:user_id/video/:video_id")
	router.GET("/user/:user_id/video/:video_id/watch", controllers.GetVideoPlayer)

	// User-level video exception management
	router.PUT("/user/:user_id/allow/:video_id")
	router.DELETE("/user/:user_id/allow/:video_id")
	router.PUT("/user/:user_id/block/:video_id")
	router.DELETE("/user/:user_id/block/:video_id")

	// User-level video view management
	router.GET("/user/:user_id/views")
	router.GET("/user/:user_id/view/:video_id")
	router.PUT("/user/:user_id/view/:video_id")

	listener, err := GetListener(port, socket)
	if err != nil {
		log.Fatalf("Error creating HTTP listener: %v", err)
	}

	server := &http.Server{
		Handler: router,
	}

	go server.Serve(listener)

	<-ctx.Done()

	log.Println("Attempting graceful shutdown of server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Println("Graceful shutdown failed: ", err)
	}

	listener.Close()
}
