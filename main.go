package main

import (
	"CodeAutoGo/utils"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	r.Use(AuthMiddleware())

	r.POST("/api/scan", ScanHandler)
	r.GET("/api/status", StatusHandler)
	r.POST("/api/build", BuildHandler)
	r.POST("/api/analyze", AnalyzeHandler)
	r.POST("/api/clone", CloneHandler)
	r.Static("/static", "./static")

	r.Run(utils.Config.Server.ListenOn)
}
