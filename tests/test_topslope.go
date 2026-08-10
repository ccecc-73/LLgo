package main

import (
	"fmt"
	"log"

	"LLhdm/topslope"
)

func main() {
	// 解析左侧宽度文件
	leftData, err := topslope.ParseWidthFile("lot1/lot1.leftslope")
	if err != nil {
		log.Fatalf("解析左侧宽度失败: %v", err)
	}
	fmt.Printf("✅ 左侧加载了 %d 个断面\n", len(leftData))

	// 解析右侧宽度文件
	rightData, err := topslope.ParseWidthFile("lot1/lot1.rightslope")
	if err != nil {
		log.Fatalf("解析右侧宽度失败: %v", err)
	}
	fmt.Printf("✅ 右侧加载了 %d 个断面\n\n", len(rightData))

	// 测试桩号和中桩高程
	testStation := 100600.0
	centerElev := 100.0 // 示例中桩高程（实际应由竖曲线提供）

	result := topslope.GetCrossSectionXY(testStation, centerElev, leftData, rightData)

	fmt.Printf("--- 桩号 %.3f 车道坐标 ---\n", testStation)
	fmt.Printf("左侧点数: %d\n", len(result.ZuoCeCheDaoJueDui))
	fmt.Printf("右侧点数: %d\n", len(result.YouCeCheDaoJueDui))

	if len(result.ZuoCeCheDaoJueDui) > 0 {
		fmt.Println("\n左侧所有点 (从外侧到中桩):")
		for i, p := range result.ZuoCeCheDaoJueDui {
			fmt.Printf("  %d: (%.3f, %.3f)\n", i+1, p.WidthX, p.GaoChengY)
		}
	}
	if len(result.YouCeCheDaoJueDui) > 0 {
		fmt.Println("\n右侧所有点 (从中桩向外):")
		for i, p := range result.YouCeCheDaoJueDui {
			fmt.Printf("  %d: (%.3f, %.3f)\n", i+1, p.WidthX, p.GaoChengY)
		}
	}

	fmt.Println("\n✅ 测试完成。")
}
