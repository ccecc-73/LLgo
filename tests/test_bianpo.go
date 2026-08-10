package main

import (
	"fmt"
	"log"

	"LLhdm/bianpo"
)

func main() {
	// 1. 解析边坡文件
	duanLuoList, err := bianpo.ParseBianPo("lot1/lot1.bianpo")
	if err != nil {
		log.Fatalf("解析失败: %v", err)
	}
	fmt.Printf("✅ 成功加载 %d 个边坡段落\n\n", len(duanLuoList))

	// 2. 打印段落概要
	for idx, d := range duanLuoList {
		fmt.Printf("--- 段落 %d ---\n", idx+1)
		fmt.Printf("  桩号范围: %.3f ~ %.3f\n", d.QiShiZhuangHao, d.JieShuZhuangHao)
		fmt.Printf("  左填点数: %d\n", len(d.ZuoTian))
		fmt.Printf("  左挖点数: %d\n", len(d.ZuoWa))
		fmt.Printf("  右填点数: %d\n", len(d.YouTian))
		fmt.Printf("  右挖点数: %d\n", len(d.YouWa))
		fmt.Println()
	}

	// 3. 测试 GetAbsolute（取第一个段落）
	if len(duanLuoList) == 0 {
		log.Fatal("没有段落可测试")
	}
	duanLuo := duanLuoList[0]

	// 假设左原点 (0,0)，右原点 (10,0)（单位：米）
	zuoYuanX, zuoYuanY := 0.0, 0.0
	youYuanX, youYuanY := 10.0, 0.0

	fmt.Println("--- GetAbsolute 测试（基于第一个段落）---")
	fmt.Printf("左原点: (%.2f, %.2f), 右原点: (%.2f, %.2f)\n", zuoYuanX, zuoYuanY, youYuanX, youYuanY)

	houxuan := bianpo.GetAbsolute(duanLuo, zuoYuanX, zuoYuanY, youYuanX, youYuanY)

	// 打印绝对坐标序列
	printPoints("左填绝对坐标", houxuan.ZuoTianJueDui)
	printPoints("左挖绝对坐标", houxuan.ZuoWaJueDui)
	printPoints("右填绝对坐标", houxuan.YouTianJueDui)
	printPoints("右挖绝对坐标", houxuan.YouWaJueDui)

	fmt.Println("\n✅ GetAbsolute 测试完成。")
}

func printPoints(label string, pts [][]float64) {
	fmt.Printf("%s (数量: %d):\n", label, len(pts))
	for i, p := range pts {
		fmt.Printf("  %d: (%.3f, %.3f)\n", i+1, p[0], p[1])
	}
	fmt.Println()
}
