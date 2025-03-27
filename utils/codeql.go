package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Task struct {
	Status   string  `json:"status"`
	Progress float32 `json:"progress"`
	Error    string  `json:"error,omitempty"`
}

func CreateCodeQLDatabase(project, language string) error {
	dbPath := fmt.Sprintf("%s/%s", Config.Storage.CodeqlDBPath, project)
	// 检查文件夹是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		os.MkdirAll(dbPath, 0755)
		log.Printf("[INFO] 创建 CodeQL 数据库文件夹 %s", dbPath)
	}

	db := fmt.Sprintf("%s/codeql_db", dbPath)
	repoPath := fmt.Sprintf("%s/%s", Config.Storage.RepoPath, project)
	log.Printf("[INFO] 创建 CodeQL 数据库")
	output, err := exec.Command(Config.CodeQL.Path, "database", "create", "--overwrite", "--language", strings.ToLower(language), "--source-root", repoPath, db).CombinedOutput()
	if err != nil {
		return fmt.Errorf("创建 CodeQL 数据库失败: %v\n%s", err, string(output))
	}
	return nil
}

func parseProgress(line, status string, re *regexp.Regexp, project string, taskStatus *sync.Map) {
	if matches := re.FindStringSubmatch(line); len(matches) == 3 {
		current, _ := strconv.Atoi(matches[1])
		total, _ := strconv.Atoi(matches[2])
		if total == 0 {
			total = 1 // 避免除零错误
		}

		taskStatus.Store(project, Task{
			Status:   status,
			Progress: float32(current) / float32(total) * 100,
		})
	}
}

func AnalyzeCodeQLDatabase(project string, taskStatus *sync.Map) error {
	db := fmt.Sprintf("%s/%s/codeql_db", Config.Storage.CodeqlDBPath, project)
	result := fmt.Sprintf("%s/%s/codeql_result.sarif", Config.Storage.RepoPath, project)

	cmd := exec.Command(Config.CodeQL.Path, "database", "analyze", db, "--format=sarifv2.1.0", "--output", result)

	cmd.Stdout = cmd.Stderr
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("无法获取 CodeQL 错误管道: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("CodeQL 启动失败: %v", err)
	}

	// 处理输出
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			// 提取形如 "[数字/数字]" 的数字作为进度信息
			loadRe := regexp.MustCompile(`\[(\d+)/(\d+)\] Loaded`)
			evalRe := regexp.MustCompile(`\[(\d+)/(\d+) eval .*\]`)

			parseProgress(line, "loading", loadRe, project, taskStatus)
			parseProgress(line, "evaluating", evalRe, project, taskStatus)
		}
		if err := scanner.Err(); err != nil {
			log.Printf("读取 CodeQL 标准输出错误: %v", err)
		}
	}()
	log.Printf("[INFO] CodeQL 分析开始")
	// 等待 CodeQL 运行结束
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("CodeQL 执行失败: %v", err)
	}

	log.Printf("[INFO] CodeQL 分析完成")
	return nil
}
