package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// 默认常量
const (
	DefaultClearDepth         = -0.3
	DefaultA4WidthMM          = 297.0
	DefaultA4HeightMM         = 210.0
	DefaultA4Gap              = 30.0
	DefaultXyzRows            = 35
	DefaultXyzCols            = 60
	DefaultTextScaleFactor    = 3.0
	DefaultLineSpacingFactor  = 1.5
)

// 运行时变量（默认值）
var (
	ClearDepth         float64 = DefaultClearDepth
	A4WidthMM          float64 = DefaultA4WidthMM
	A4HeightMM         float64 = DefaultA4HeightMM
	A4Gap              float64 = DefaultA4Gap
	XyzRows            int     = DefaultXyzRows
	XyzCols            int     = DefaultXyzCols
	TextScaleFactor    float64 = DefaultTextScaleFactor
	LineSpacingFactor  float64 = DefaultLineSpacingFactor
)

// LoadConfig 从项目文件夹加载配置文件
// 格式：每行一个配置项，等号分隔，分号后面为注释
// 示例: ClearDepth = -0.3    ; 清表深度（负值表示向下）
func LoadConfig(projectName, projectDir string) {
	// 重置为默认值
	ClearDepth = DefaultClearDepth
	A4WidthMM = DefaultA4WidthMM
	A4HeightMM = DefaultA4HeightMM
	A4Gap = DefaultA4Gap
	XyzRows = DefaultXyzRows
	XyzCols = DefaultXyzCols
	TextScaleFactor = DefaultTextScaleFactor
	LineSpacingFactor = DefaultLineSpacingFactor

	configPath := filepath.Join(projectDir, projectName+".config")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return
	}

	file, err := os.Open(configPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// 去掉行内注释（分号后面的内容）
		if idx := strings.Index(line, ";"); idx != -1 {
			line = line[:idx]
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "ClearDepth":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				ClearDepth = v
			}
		case "A4WidthMM":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				A4WidthMM = v
			}
		case "A4HeightMM":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				A4HeightMM = v
			}
		case "A4Gap":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				A4Gap = v
			}
		case "XyzRows":
			if v, err := strconv.Atoi(val); err == nil {
				XyzRows = v
			}
		case "XyzCols":
			if v, err := strconv.Atoi(val); err == nil {
				XyzCols = v
			}
		case "TextScaleFactor":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				TextScaleFactor = v
			}
		case "LineSpacingFactor":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				LineSpacingFactor = v
			}
		}
	}
}





/**
package config

// 全局配置常量（与 C# AppConfig 完全一致）
const (
	// ClearDepth 清表深度（负值表示向下）
	ClearDepth = -0.3

	// A4 图纸尺寸（毫米）
	A4WidthMM  = 297.0
	A4HeightMM = 210.0

	// A4Gap 图纸边距（毫米）
	A4Gap = 30.0

	// XyzRows / XyzCols 地形网格行列数（用于生成横断面地形矩阵）
	XyzRows = 35
	XyzCols = 60

	// TextScaleFactor 文字缩放因子
	TextScaleFactor = 3.0

	// LineSpacingFactor 行距因子（用于多行文本）
	LineSpacingFactor = 1.5
)

**/