package main

import (
	"fmt"
	"log"

	"LLhdm/jiegouceng"
)

func main() {
	// 解析左侧结构层配置文件
	leftConfigs, err := jiegouceng.ParseConfigFile("lot1/lot1.leftstructure")
	if err != nil {
		log.Fatalf("解析左侧结构层失败: %v", err)
	}
	fmt.Printf("✅ 左侧加载了 %d 个结构层配置\n", len(leftConfigs))

	// 解析右侧结构层配置文件
	rightConfigs, err := jiegouceng.ParseConfigFile("lot1/lot1.rightstructure")
	if err != nil {
		log.Fatalf("解析右侧结构层失败: %v", err)
	}
	fmt.Printf("✅ 右侧加载了 %d 个结构层配置\n\n", len(rightConfigs))

	// 测试桩号，假设中心点
	center := jiegouceng.NewPoint2D(0, 100)
	crossfall := -2.5 // 横坡百分比

	// 左侧计算
	fmt.Println("--- 左侧结构层 ---")
	leftPoly, leftSub := jiegouceng.ComputeLeftCoordinates(100600, center, crossfall, leftConfigs)
	fmt.Printf("生成 %d 个层多边形\n", len(leftPoly))
	fmt.Printf("设计线包含 %d 个点\n", len(leftSub))
	if len(leftPoly) > 0 {
		fmt.Println("第一层多边形顶点 (顶外, 底外, 底内, 顶内):")
		for i, pt := range leftPoly[0] {
			fmt.Printf("  %d: (%.3f, %.3f)\n", i+1, pt[0], pt[1])
		}
	}

	// 右侧计算
	fmt.Println("\n--- 右侧结构层 ---")
	rightPoly, rightSub := jiegouceng.ComputeRightCoordinates(100600, center, crossfall, rightConfigs)
	fmt.Printf("生成 %d 个层多边形\n", len(rightPoly))
	fmt.Printf("设计线包含 %d 个点\n", len(rightSub))
	if len(rightPoly) > 0 {
		fmt.Println("第一层多边形顶点 (顶外, 底外, 底内, 顶内):")
		for i, pt := range rightPoly[0] {
			fmt.Printf("  %d: (%.3f, %.3f)\n", i+1, pt[0], pt[1])
		}
	}

	fmt.Println("\n✅ 测试完成。")
}