package api

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"CodeAutoGo/appcontext"
	"CodeAutoGo/cmdclient"
	"CodeAutoGo/models"
	"CodeAutoGo/repository"

	"github.com/gin-gonic/gin"
)

func finished(project string, taskStatus *sync.Map) {
	repository.SaveTaskStatus(models.Task{
		ProjectName: project,
		Status:      "finished",
		Content:     "任务已完成",
	})
	taskStatus.Delete(project)
}

func failed(project string, taskStatus *sync.Map, err error) {
	repository.SaveTaskStatus(models.Task{
		ProjectName: project,
		Status:      "failed",
		Content:     "任务失败: " + err.Error(),
	})
	taskStatus.Delete(project)
}

func AuthMiddleware(ac *appcontext.AppContext) gin.HandlerFunc {
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
		if token != ac.Config.Server.Token {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, invalid token"})
			c.Abort()
			return
		}

		// 如果验证通过，则继续处理请求
		c.Next()
	}
}

func ScanHandler(c *gin.Context, ac *appcontext.AppContext) {
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
		project, err := ac.GitClient.CloneRepo(request.RepoURL, request.Branch)
		if err != nil {
			log.Printf("[ERROR] 克隆项目代码失败: %v", err)
			failed(project, &ac.TaskStatus, err)
			return
		}

		// Step2: Get Language
		log.Printf("[INFO] 准备获取项目主要语言")
		language, err := ac.GitClient.GetProjectLanguage(request.RepoURL)
		if err != nil {
			log.Printf("[ERROR] 获取项目主要语言失败: %v", err)
			failed(project, &ac.TaskStatus, err)
			return
		}

		// Step3: Create CodeQL Database
		log.Printf("[INFO] 准备创建 CodeQL 数据库")
		ac.TaskStatus.Store(project, cmdclient.Task{
			Status:   "building",
			Progress: 0,
		})
		if err := ac.CodeQLClient.CreateCodeQLDatabase(project, language); err != nil {
			log.Printf("[ERROR] 创建 CodeQL 数据库失败: %v", err)
			failed(project, &ac.TaskStatus, err)
			return
		}

		// Step4: Run CodeQL Analyze
		log.Printf("[INFO] 准备运行 CodeQL 分析")
		if err := ac.CodeQLClient.AnalyzeCodeQLDatabase(project, &ac.TaskStatus); err != nil {
			log.Printf("[ERROR] 运行 CodeQL 分析失败: %v", err)
			failed(project, &ac.TaskStatus, err)
			return
		}

		// Step5: 完成任务
		finished(project, &ac.TaskStatus)
	}()

	// 立即响应客户端
	c.JSON(202, gin.H{"message": "Scan request accepted, processing in background"})
}

func BuildHandler(c *gin.Context, ac *appcontext.AppContext) {
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
		ac.TaskStatus.Store(request.Project, cmdclient.Task{
			Status:   "building",
			Progress: 0,
		})
		if err := ac.CodeQLClient.CreateCodeQLDatabase(request.Project, request.Language); err != nil {
			log.Printf("[ERROR] 创建 CodeQL 数据库失败: %v", err)
			failed(request.Project, &ac.TaskStatus, err)
			return
		}
		finished(request.Project, &ac.TaskStatus)
	}()

	// 立即响应客户端
	c.JSON(202, gin.H{"message": "Build request accepted, processing in background"})

}

func AnalyzeHandler(c *gin.Context, ac *appcontext.AppContext) {
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
		if err := ac.CodeQLClient.AnalyzeCodeQLDatabase(request.Project, &ac.TaskStatus); err != nil {
			log.Printf("[ERROR] 运行 CodeQL 分析失败: %v", err)
			failed(request.Project, &ac.TaskStatus, err)
			return
		}
		finished(request.Project, &ac.TaskStatus)
	}()

	// 立即响应客户端
	c.JSON(202, gin.H{"message": "Analyze request accepted, processing in background"})
}

func CloneHandler(c *gin.Context, ac *appcontext.AppContext) {
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
		project, err := ac.GitClient.CloneRepo(request.RepoURL, request.Branch)
		if err != nil {
			log.Printf("[ERROR] 克隆项目代码失败: %v", err)
			failed(project, &ac.TaskStatus, err)
			return
		}

		finished(project, &ac.TaskStatus)
	}()

	// 立即响应客户端
	c.JSON(202, gin.H{"message": "Clone request accepted, processing in background"})
}

func StatusHandler(c *gin.Context, ac *appcontext.AppContext) {
	repoURL := c.Query("repo")
	repoInfo := strings.SplitN(repoURL, "/", 4)
	if len(repoInfo) != 4 {
		c.JSON(400, gin.H{"error": "Invalid repo URL"})
	}
	project := repoInfo[3]
	if status, ok := ac.TaskStatus.Load(project); ok {
		c.JSON(200, gin.H{"status": status})
	} else {
		c.JSON(404, gin.H{"error": "Task not found"})
	}
}
