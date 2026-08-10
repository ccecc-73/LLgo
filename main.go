package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"LLhdm/processor"
)

func main() {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("获取当前目录失败: %v", err)
	}

	// 读取所有子目录
	entries, err := os.ReadDir(cwd)
	if err != nil {
		log.Fatalf("读取目录失败: %v", err)
	}

	// 收集有效项目目录（排除隐藏目录、系统目录和结果目录）
	var projects []string
	exclude := map[string]bool{
		"bin":     true,
		"obj":     true,
		"result":  true,
		".git":    true,
		".idea":   true,
		"vendor":  true,
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// 跳过隐藏目录（以 "." 开头）
		if strings.HasPrefix(name, ".") {
			continue
		}
		if exclude[name] {
			continue
		}
		projects = append(projects, name)
	}

	if len(projects) == 0 {
		log.Fatal("当前目录下没有找到任何项目文件夹。")
	}

	// 显示项目列表
	fmt.Println("\n📁 可用的项目：")
	for i, name := range projects {
		fmt.Printf("  %d. %s\n", i+1, name)
	}

	// 用户选择
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n请输入项目序号 (1~", len(projects), ")，或输入 0 退出: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("读取输入失败: %v", err)
	}
	input = strings.TrimSpace(input)
	idx, err := strconv.Atoi(input)
	if err != nil || idx < 0 || idx > len(projects) {
		log.Fatal("无效的序号，请输入 0~", len(projects))
	}
	if idx == 0 {
		fmt.Println("退出程序。")
		return
	}

	// 处理选中的项目
	selected := projects[idx-1]
	log.Printf("开始处理项目: %s", selected)

	startAll := time.Now()
	err = processor.Process(selected)
	elapsed := time.Since(startAll)

	if err != nil {
		log.Fatalf("处理失败: %v", err)
	}
	log.Printf("项目 %s 处理完成，总耗时: %v", selected, elapsed)
}