package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// 定义检查结果的结构体
type CheckResult struct {
	Host       string
	Available  bool
	Time       time.Duration
	StatusCode int
	IsTimeout  bool
}

// 从GitHub下载docker.txt
func downloadFromGithub() error {
	url := "https://raw.githubusercontent.com/YMingPro/docker-register-check/main/docker.txt"

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	file, err := os.Create("docker.txt")
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	return nil
}

// 定义worker池来处理检查任务
func worker(id int, jobs <-chan string, results chan<- CheckResult, timeout time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	for host := range jobs {
		start := time.Now()
		result := CheckResult{
			Host: host,
		}

		url := fmt.Sprintf("https://%s/v2/", host)
		resp, err := client.Get(url)

		if err != nil {
			result.Available = false
			if os.IsTimeout(err) || strings.Contains(err.Error(), "timeout") {
				result.IsTimeout = true
			}
			results <- result
			continue
		}

		result.StatusCode = resp.StatusCode
		result.Time = time.Since(start)
		result.Available = (resp.StatusCode >= 200 && resp.StatusCode < 400) || resp.StatusCode == 401

		resp.Body.Close()
		results <- result
	}
}

// 等待用户按键
func waitForKeyPress() {
	fmt.Println("\n按回车键退出...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// 显示进度条
func showProgress(current, total int) {
	width := 40 // 进度条宽度
	percentage := float64(current) / float64(total)
	filled := int(float64(width) * percentage)

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Printf("\r检测进度: [%s] %d/%d (%.1f%%)", bar, current, total, percentage*100)
}

func main() {
	// 定义命令行参数
	timeoutPtr := flag.Float64("timeout", 10.0, "请求超时时间（秒）")
	workersPtr := flag.Int("workers", runtime.NumCPU()*2, "并发worker数量")
	updatePtr := flag.Bool("update", false, "强制从GitHub更新docker.txt")
	listSuccessPtr := flag.Bool("l", false, "只显示成功的结果")
	flag.Parse()

	timeout := time.Duration(*timeoutPtr * float64(time.Second))
	numWorkers := *workersPtr

	fmt.Printf("启动检测 (并发数: %d, 超时: %.1fs)\n", numWorkers, timeout.Seconds())

	// 处理文件更新逻辑
	if *updatePtr {
		fmt.Println("正在从GitHub更新docker.txt...")
		if err := downloadFromGithub(); err != nil {
			fmt.Printf("更新失败: %v\n", err)
			waitForKeyPress()
			return
		}
		fmt.Println("更新成功!")
	} else if _, err := os.Stat("docker.txt"); os.IsNotExist(err) {
		fmt.Println("本地未找到docker.txt，正在从GitHub下载...")
		if err := downloadFromGithub(); err != nil {
			fmt.Printf("下载失败: %v\n", err)
			waitForKeyPress()
			return
		}
		fmt.Println("下载成功!")
	}

	// 打开docker.txt文件
	file, err := os.Open("docker.txt")
	if err != nil {
		fmt.Printf("无法打开docker.txt文件: %v\n", err)
		waitForKeyPress()
		return
	}
	defer file.Close()

	// 读取所有hosts
	var hosts []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		host := strings.TrimSpace(scanner.Text())
		if host != "" && !strings.HasPrefix(host, "#") {
			hosts = append(hosts, host)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("读取文件出错: %v\n", err)
		waitForKeyPress()
		return
	}

	if len(hosts) == 0 {
		fmt.Println("docker.txt 文件为空或没有有效的主机地址")
		waitForKeyPress()
		return
	}

	// 创建任务和结果通道
	jobs := make(chan string, len(hosts))
	results := make(chan CheckResult, len(hosts))

	// 启动worker池
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, results, timeout, &wg)
	}

	// 发送所有任务
	for _, host := range hosts {
		jobs <- host
	}
	close(jobs)

	// 收集结果
	allResults := make([]CheckResult, 0, len(hosts))
	resultCount := 0

	// 在后台等待所有worker完成并关闭results通道
	go func() {
		wg.Wait()
		close(results)
	}()

	// 显示进度并收集结果
	fmt.Println() // 为进度条留出空行
	for result := range results {
		resultCount++
		allResults = append(allResults, result)
		showProgress(resultCount, len(hosts))
	}

	// 根据-l参数过滤结果
	var displayResults []CheckResult
	if *listSuccessPtr {
		for _, result := range allResults {
			if result.Available && !result.IsTimeout {
				displayResults = append(displayResults, result)
			}
		}
	} else {
		displayResults = allResults
	}

	// 按主机名排序结果
	sort.Slice(displayResults, func(i, j int) bool {
		return displayResults[i].Host < displayResults[j].Host
	})

	// 清除进度条并显示结果
	fmt.Println("\n\nRegistry                        状态       状态码     响应时间")
	fmt.Println(strings.Repeat("-", 65))

	for _, result := range displayResults {
		status := "✓"
		if !result.Available {
			status = "✗"
		}

		statusCode := fmt.Sprintf("%d", result.StatusCode)
		if result.StatusCode == 0 {
			statusCode = "-"
		}

		timeStr := "超时"
		if !result.IsTimeout {
			timeStr = fmt.Sprintf("%.2fs", result.Time.Seconds())
		}

		fmt.Printf("%-30s %-10s %-10s %-15s\n",
			result.Host,
			status,
			statusCode,
			timeStr,
		)
	}

	// 显示统计信息
	totalCount := len(allResults)
	successCount := 0
	for _, result := range allResults {
		if result.Available && !result.IsTimeout {
			successCount++
		}
	}
	fmt.Printf("\n检测完成! (成功: %d, 总计: %d)\n", successCount, totalCount)
	waitForKeyPress()
}
