package main

import (
	"CodeAutoGo/utils"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var taskStatus sync.Map

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在这里实现你的权限校验逻辑
		// 例如，从请求头中获取 token 并验证
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, missing token"})
			c.Abort()
			return
		}

		// 假设 validateToken 是一个验证 token 的函数
		if token != utils.Config.Server.Token {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, invalid token"})
			c.Abort()
			return
		}

		// 如果验证通过，则继续处理请求
		c.Next()
	}
}

func ScanHandler(c *gin.Context) {
	var request struct {
		RepoURL string `json:"repo_url"`
		Branch  string `json:"branch"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Step1: 异步处理
	go func() {
		// Step1: Clone Repo
		log.Printf("[INFO] 准备克隆项目代码")
		project, err := utils.CloneRepo(request.RepoURL, request.Branch)
		if err != nil {
			log.Printf("[ERROR] 克隆项目代码失败: %v", err)
			taskStatus.Store(project, utils.Task{
				Status:   "failed",
				Progress: -1,
				Error:    err.Error(),
			})
			return
		}

		// Step2: Get Language
		log.Printf("[INFO] 准备获取项目主要语言")
		language, err := utils.GetProjectLanguage(request.RepoURL)
		if err != nil {
			log.Printf("[ERROR] 获取项目主要语言失败: %v", err)
			taskStatus.Store(project, utils.Task{
				Status:   "failed",
				Progress: -1,
				Error:    err.Error(),
			})
			return
		}

		// Step3: Create CodeQL Database
		log.Printf("[INFO] 准备创建 CodeQL 数据库")
		taskStatus.Store(project, utils.Task{
			Status:   "building",
			Progress: 0,
		})
		if err := utils.CreateCodeQLDatabase(project, language); err != nil {
			log.Printf("[ERROR] 创建 CodeQL 数据库失败: %v", err)
			taskStatus.Store(project, utils.Task{
				Status:   "failed",
				Progress: -1,
				Error:    err.Error(),
			})
			return
		}

		// Step4: Run CodeQL Analyze
		log.Printf("[INFO] 准备运行 CodeQL 分析")
		if err := utils.AnalyzeCodeQLDatabase(project, &taskStatus); err != nil {
			log.Printf("[ERROR] 运行 CodeQL 分析失败: %v", err)
			taskStatus.Store(project, utils.Task{
				Status:   "failed",
				Progress: -1,
				Error:    err.Error(),
			})
			return
		}

		// Step5: 完成任务
		taskStatus.Store(project, utils.Task{
			Status:   "finished",
			Progress: 100,
		})
	}()

	// 立即响应客户端
	c.JSON(202, gin.H{"message": "Scan request accepted, processing in background"})
}

func BuildHandler(c *gin.Context) {
	var request struct {
		Project  string `json:"project"`
		Language string `json:"language"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Step1: 异步处理
	go func() {

		log.Printf("[INFO] 准备创建 CodeQL 数据库")
		taskStatus.Store(request.Project, utils.Task{
			Status:   "building",
			Progress: 0,
		})
		if err := utils.CreateCodeQLDatabase(request.Project, request.Language); err != nil {
			log.Printf("[ERROR] 创建 CodeQL 数据库失败: %v", err)
			taskStatus.Store(request.Project, utils.Task{
				Status:   "failed",
				Progress: -1,
				Error:    err.Error(),
			})
			return
		}
		taskStatus.Store(request.Project, utils.Task{
			Status:   "finished",
			Progress: 100,
		})
	}()

	// 立即响应客户端
	c.JSON(202, gin.H{"message": "Build request accepted, processing in background"})

}

func AnalyzeHandler(c *gin.Context) {
	var request struct {
		Project string `json:"project"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Step1: 异步处理
	go func() {
		log.Printf("[INFO] 准备运行 CodeQL 分析")
		if err := utils.AnalyzeCodeQLDatabase(request.Project, &taskStatus); err != nil {
			log.Printf("[ERROR] 运行 CodeQL 分析失败: %v", err)
			taskStatus.Store(request.Project, utils.Task{
				Status:   "failed",
				Progress: -1,
				Error:    err.Error(),
			})
			return
		}
		taskStatus.Store(request.Project, utils.Task{
			Status:   "finished",
			Progress: 100,
		})
	}()

	// 立即响应客户端
	c.JSON(202, gin.H{"message": "Analyze request accepted, processing in background"})
}

func CloneHandler(c *gin.Context) {
	var request struct {
		RepoURL string `json:"repo_url"`
		Branch  string `json:"branch"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Step1: 异步处理
	go func() {
		// Step1: Clone Repo
		log.Printf("[INFO] 准备克隆项目代码")
		project, err := utils.CloneRepo(request.RepoURL, request.Branch)
		if err != nil {
			log.Printf("[ERROR] 克隆项目代码失败: %v", err)
			taskStatus.Store(project, utils.Task{
				Status:   "failed",
				Progress: -1,
				Error:    err.Error(),
			})
			return
		}

		taskStatus.Store(project, utils.Task{
			Status:   "finished",
			Progress: 100,
		})
	}()

	// 立即响应客户端
	c.JSON(202, gin.H{"message": "Clone request accepted, processing in background"})
}

func StatusHandler(c *gin.Context) {
	repoURL := c.Query("repo")
	repoInfo := strings.SplitN(repoURL, "/", 4)
	if len(repoInfo) != 4 {
		c.JSON(400, gin.H{"error": "Invalid repo URL"})
	}
	project := repoInfo[3]
	if status, ok := taskStatus.Load(project); ok {
		c.JSON(200, gin.H{"status": status})
	} else {
		c.JSON(404, gin.H{"error": "Task not found"})
	}
}
