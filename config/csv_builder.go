package config

import (
	"fmt"
	"strings"

	"LLhdm/jiegouceng"
)

// BuildHeader 生成 CSV 表头
// leftStructures / rightStructures: 左右结构层配置列表
// 返回: CSV 头字符串（不含换行）
func BuildHeader(leftStructures, rightStructures []jiegouceng.JiegoucengConfig) string {
	var sb strings.Builder
	sb.WriteString("桩号,填方面积(㎡),挖方面积(㎡),清表面积(㎡),清表左边界X,清表右边界X")

	if leftStructures != nil {
		for i := 0; i < len(leftStructures); i++ {
			name := leftStructures[i].LayerName
			if name == "" {
				name = fmt.Sprintf("left%d", i+1)
			}
			sb.WriteString(",")
			sb.WriteString(name)
		}
	}

	if rightStructures != nil {
		for i := 0; i < len(rightStructures); i++ {
			name := rightStructures[i].LayerName
			if name == "" {
				name = fmt.Sprintf("right%d", i+1)
			}
			sb.WriteString(",")
			sb.WriteString(name)
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

// BuildDataRow 生成 CSV 数据行
// station: 桩号（字符串或数值）
// fill, cut, clearArea: 填方、挖方、清表面积
// minX, maxX: 清表左右边界 X 坐标
// layerAreasCsv: 各结构层面积（已格式化为 CSV 字符串，如 ",1.234,5.678"）
// 返回: CSV 数据行（不含换行）
func BuildDataRow(station string, fill, cut, clearArea, minX, maxX float64, layerAreasCsv string) string {
	return fmt.Sprintf("%s,%.3f,%.3f,%.3f,%.3f,%.3f%s",
		station, fill, cut, clearArea, minX, maxX, layerAreasCsv)
}