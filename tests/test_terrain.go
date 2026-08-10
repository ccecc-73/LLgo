package main

import (
	"fmt"
	"log"

	"LLhdm/terrain"
)

func main() {
	mesh, err := terrain.FromTextFile("lot1/lot1.xyz", 50, 0)
	if err != nil {
		log.Fatalf("加载地形失败: %v", err)
	}
	fmt.Printf("✅ 三角形数量: %d\n", mesh.TriangleCount())

	// 一条切割线段（例如横断面）
	segments := [][]float64{{ 9595756.593,441032.536, 9595729.169,  441003.417 }}
	results := mesh.ExtractSections(segments)
	fmt.Printf("切割得到 %d 个交点\n", len(results))
	if len(results) > 0 {
		fmt.Println("前5个交点 (距离, 高程):")
		for i := 0;  i < len(results); i++ {
			fmt.Printf("  (%10.3f, %10.3f)\n", results[i][0], results[i][1])
		}
	}
}