package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func CloneRepo(repoURL, branch string) (string, error) {
	repoInfo := strings.SplitN(repoURL, "/", 4)
	if len(repoInfo) != 4 {
		return "", fmt.Errorf("仓库地址无效: %s", repoURL)
	}
	scheme, _, domain, project := repoInfo[0], repoInfo[1], repoInfo[2], repoInfo[3]
	localRepoPath := fmt.Sprintf("%s/%s", Config.Storage.RepoPath, project)
	// 临时切换工作目录到localRepoPath检查是否存在仓库
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = localRepoPath
	if err := cmd.Run(); err == nil {
		// 本地仓库存在，切换到对应分支
		log.Printf("[INFO] 本地仓库存在，切换到对应分支")
		cmd = exec.Command("git", "checkout", branch)
		cmd.Dir = localRepoPath
		if err := cmd.Run(); err != nil {
			return "", err
		}
		// 更新代码
		log.Printf("[INFO] 更新代码")
		cmd = exec.Command("git", "pull")
		cmd.Dir = localRepoPath
		if err := cmd.Run(); err != nil {
			return "", err
		}
		return project, nil
	}

	// 本地仓库不存在，克隆代码
	log.Printf("[INFO] 本地仓库不存在，克隆代码")
	repoURL = fmt.Sprintf("%s//oauth2:%s@%s/%s.git", scheme, Config.GitLab.Token, domain, project)
	output, err := exec.Command("git", "clone", "--branch", branch, "--single-branch", repoURL, localRepoPath).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s, %s", err, string(output))
	}

	return project, nil
}

func GetProjectLanguage(repoURL string) (string, error) {
	repoInfo := strings.SplitN(repoURL, "/", 4)
	if len(repoInfo) != 4 {
		return "", fmt.Errorf("invalid repo url: %s", repoURL)
	}
	scheme, _, domain, project := repoInfo[0], repoInfo[1], repoInfo[2], repoInfo[3]
	apiURL := fmt.Sprintf("%s//%s/api/v4/projects/%s/languages", scheme, domain, strings.ReplaceAll(project, "/", "%2F"))
	log.Printf("[INFO] 获取项目主要语言")
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("PRIVATE-TOKEN", Config.GitLab.Token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var languages map[string]float64
	if err := json.Unmarshal(body, &languages); err != nil {
		return "", err
	}

	for lang := range languages {
		for _, supported := range Config.SupportedLanguages {
			if lang == supported {
				return lang, nil
			}
		}
	}

	return "", fmt.Errorf("未获取到支持的语言: %v", languages)
}
