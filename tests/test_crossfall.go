package main

import (
	"fmt"
	"log"

	"LLhdm/jiegouceng"
)

func main() {
	// 解析左侧横坡文件
	leftRecords, err := jiegouceng.ParseCrossfallFile("lot1/lot1.leftcrossfall")
	if err != nil {
		log.Fatalf("解析左侧横坡失败: %v", err)
	}
	fmt.Printf("✅ 左侧加载了 %d 条横坡记录\n", len(leftRecords))

	// 解析右侧横坡文件
	rightRecords, err := jiegouceng.ParseCrossfallFile("lot1/lot1.rightcrossfall")
	if err != nil {
		log.Fatalf("解析右侧横坡失败: %v", err)
	}
	fmt.Printf("✅ 右侧加载了 %d 条横坡记录\n\n", len(rightRecords))

	// 测试桩号
	testStation := 100600.0
	fmt.Printf("--- 插值测试 (桩号 %.3f) ---\n", testStation)

	leftSlope := jiegouceng.InterpolateCrossfall(leftRecords, testStation)
	rightSlope := jiegouceng.InterpolateCrossfall(rightRecords, testStation)

	fmt.Printf("  左侧横坡: %.4f%%\n", leftSlope)
	fmt.Printf("  右侧横坡: %.4f%%\n", rightSlope)

	// 再测一个桩号
	testStation2 := 105000.0
	fmt.Printf("\n--- 插值测试 (桩号 %.3f) ---\n", testStation2)
	leftSlope2 := jiegouceng.InterpolateCrossfall(leftRecords, testStation2)
	rightSlope2 := jiegouceng.InterpolateCrossfall(rightRecords, testStation2)
	fmt.Printf("  左侧横坡: %.4f%%\n", leftSlope2)
	fmt.Printf("  右侧横坡: %.4f%%\n", rightSlope2)

	fmt.Println("\n✅ 测试完成。")
}
