package main

import (
	"CodeAutoGo/api"
	"CodeAutoGo/appcontext"
	"CodeAutoGo/cmdclient"
	"CodeAutoGo/config"
	"CodeAutoGo/database"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
)

func main() {
	appConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[ERROR] 配置加载失败: %v", err)
	}

	gitClient := cmdclient.NewGitClient(appConfig.CmdClient.GitPath, appConfig.Storage.RepoPath, appConfig.GitLab.Token, appConfig.SupportedLanguages)
	codeqlClient := cmdclient.NewCodeQLClient(appConfig.CmdClient.CodeQLPath, appConfig.Storage.CodeQLDBPath, appConfig.Storage.RepoPath)

	appContext := &appcontext.AppContext{
		GitClient:    gitClient,
		CodeQLClient: codeqlClient,
		Config:       appConfig,
		TaskStatus:   sync.Map{},
	}

	database.ConnectDB(appConfig.Database.MongoURI, appConfig.Database.DBName, appConfig.Database.Username, appConfig.Database.Password)
	defer database.DisconnectDB()

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		api.AuthMiddleware(appContext)
	})

	r.POST("/api/scan", func(c *gin.Context) {
		api.ScanHandler(c, appContext)
	})
	r.GET("/api/status", func(c *gin.Context) {
		api.StatusHandler(c, appContext)
	})
	r.POST("/api/build", func(c *gin.Context) {
		api.BuildHandler(c, appContext)
	})
	r.POST("/api/analyze", func(c *gin.Context) {
		api.AnalyzeHandler(c, appContext)
	})
	r.POST("/api/clone", func(c *gin.Context) {
		api.CloneHandler(c, appContext)
	})
	r.Static("/static", "./static")
	r.Static("/static/codeql_dbs", "./codeql_dbs")

	r.Run() // 运行服务
}
