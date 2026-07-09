package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"yourtube/internal/controllers"

	"yourtube/internal/repositories"
	"yourtube/internal/util"

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

	repositories.InitDb(dsn)

	// Healthcheck
	router.GET("/_healthcheck")

	// Auth
	router.POST("/authenticate")

	// Administrative actions
	router.POST("/admin")

	// Global channel management
	router.GET("/channel", controllers.GetChannels)
	router.POST("/channel", controllers.AddChannel)
	router.GET("/channel/:channel_id", controllers.GetChannel)
	router.DELETE("/channel/:channel_id", controllers.DeleteChannel)

	// Global video management
	router.GET("/video", controllers.GetVideos)
	router.GET("/video/:video_id", controllers.GetVideo)
	router.PUT("/video/:video_id", controllers.PutVideo)

	// Family management
	router.POST("/family")
	router.GET("/family/:family_id")
	router.PATCH("/family/:family_id")

	// Family-level channel management
	router.GET("/family/:family_id/allowed-channel")
	router.POST("/family/:family_id/allowed-channel")
	router.DELETE("/family/:family_id/allowed-channel/:channel_id")

	// Family-level video exception management
	router.GET("/family/:family_id/allowed-video")
	router.PUT("/family/:family_id/allowed-video/:video_id")
	router.DELETE("/family/:family_id/allowed-video/:video_id")
	router.GET("/family/:family_id/blocked-video")
	router.PUT("/family/:family_id/blocked-video/:video_id")
	router.DELETE("/family/:family_id/blocked-video/:video_id")

	// User management
	router.POST("/user", controllers.AddUser)
	router.GET("/user/:user_id")
	router.PUT("/user/:user_id")

	// User-level channel management
	router.GET("/user/:user_id/allowed-channel")
	router.POST("/user/:user_id/allowed-channel", controllers.AllowUserChannel)
	router.GET("/user/:user_id/allowed-channel/:channel_id")
	router.DELETE("/user/:user_id/allowed-channel/:channel_id", controllers.DeleteAllowedUserChannel)
	router.GET("/user/:user_id/blocked-channel")
	router.POST("/user/:user_id/blocked-channel")
	router.DELETE("/user/:user_id/blocked-channel/:channel_id")

	// User-level video exception management
	router.GET("/user/:user_id/allowed-video")
	router.PUT("/user/:user_id/allowed-video/:video_id")
	router.DELETE("/user/:user_id/allowed-video/:video_id")
	router.GET("/user/:user_id/blocked-video")
	router.PUT("/user/:user_id/blocked-video/:video_id")
	router.DELETE("/user/:user_id/blocked-video/:video_id")

	// User-level available videos
	router.GET("/user/:user_id/video", controllers.GetUserVideos)
	router.GET("/user/:user_id/video/:video_id", controllers.GetUserVideo)

	// User-level video view management
	router.GET("/user/:user_id/view", controllers.GetViewedVideos)
	router.POST("/user/:user_id/view/:video_id", controllers.UpdateProgress)

	listener, err := util.GetListener(port, socket)
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
