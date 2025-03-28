package cmdclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type GitClient struct {
	gitPath            string
	repoPath           string
	gitlabToken        string
	supportedLanguages []string
}

func NewGitClient(gitPath, repoPath, gitlabToken string, supportLanguages []string) *GitClient {
	return &GitClient{
		gitPath:            gitPath,
		repoPath:           repoPath,
		gitlabToken:        gitlabToken,
		supportedLanguages: supportLanguages,
	}
}

func (c *GitClient) CloneRepo(repoURL, branch string) (string, error) {
	repoInfo := strings.SplitN(repoURL, "/", 4)
	if len(repoInfo) != 4 {
		return "", fmt.Errorf("仓库地址无效: %s", repoURL)
	}
	scheme, _, domain, project := repoInfo[0], repoInfo[1], repoInfo[2], repoInfo[3]
	localRepoPath := fmt.Sprintf("%s/%s", c.repoPath, project)
	// 临时切换工作目录到localRepoPath检查是否存在仓库
	cmd := exec.Command(c.gitPath, "rev-parse", "--is-inside-work-tree")
	cmd.Dir = localRepoPath
	if err := cmd.Run(); err == nil {
		// 本地仓库存在，切换到对应分支
		log.Printf("[INFO] 本地仓库存在，切换到对应分支")
		cmd = exec.Command(c.gitPath, "checkout", branch)
		cmd.Dir = localRepoPath
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("切换分支失败: %v", err)
		}
		// 更新代码
		log.Printf("[INFO] 更新代码")
		cmd = exec.Command(c.gitPath, "pull")
		cmd.Dir = localRepoPath
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("更新代码失败: %v", err)
		}
		return project, nil
	}

	// 本地仓库不存在，克隆代码
	log.Printf("[INFO] 本地仓库不存在，克隆代码")
	repoURL = fmt.Sprintf("%s//oauth2:%s@%s/%s.git", scheme, c.gitlabToken, domain, project)
	output, err := exec.Command("git", "clone", "--branch", branch, "--single-branch", repoURL, localRepoPath).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("克隆代码失败: %v, %s", err, string(output))
	}

	return project, nil
}

// GetProjectLanguage 获取指定 GitLab 项目的主要语言。
// 参数:
// - repoURL: GitLab 项目的仓库 URL，格式通常为 "https://domain.com/namespace/project"。
// 返回值:
// - string: 项目的主要语言名称，如果未找到支持的语言则返回空字符串。
// - error: 如果发生错误（如 URL 格式无效、请求失败、解析失败等），返回相应的错误信息。
func (c *GitClient) GetProjectLanguage(repoURL string) (string, error) {
	// 将仓库 URL 按 "/" 分割，提取协议、域名和项目路径。
	repoInfo := strings.SplitN(repoURL, "/", 4)
	if len(repoInfo) != 4 {
		return "", fmt.Errorf("invalid repo url: %s", repoURL)
	}
	scheme, _, domain, project := repoInfo[0], repoInfo[1], repoInfo[2], repoInfo[3]

	// 构造 GitLab API 的语言统计接口 URL。
	apiURL := fmt.Sprintf("%s//%s/api/v4/projects/%s/languages", scheme, domain, strings.ReplaceAll(project, "/", "%2F"))
	log.Printf("[INFO] 获取项目主要语言")

	// 创建 HTTP GET 请求以获取项目语言统计信息。
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头中的私有令牌用于身份验证。
	req.Header.Set("PRIVATE-TOKEN", c.gitlabToken)

	// 发送 HTTP 请求并获取响应。
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码是否为 200 OK，否则返回错误。
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应体内容。
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应体为语言统计的键值对（语言名称 -> 使用比例）。
	var languages map[string]float64
	if err := json.Unmarshal(body, &languages); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 遍历语言统计结果，检查是否有支持的语言。
	for lang := range languages {
		for _, supported := range c.supportedLanguages {
			if lang == supported {
				return lang, nil
			}
		}
	}

	// 如果未找到支持的语言，返回错误。
	return "", fmt.Errorf("未获取到支持的语言: %v", languages)
}
