package jiegouceng

import (
	"bufio"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ---------- 基础点结构 ----------

// Point2D 二维点（左右侧共用）
type Point2D struct {
	X float64
	Y float64
}

// NewPoint2D 构造点
func NewPoint2D(x, y float64) Point2D {
	return Point2D{X: x, Y: y}
}

// ---------- 结构层配置 ----------

// JiegoucengConfig 结构层配置（左右通用）
type JiegoucengConfig struct {
	StartStation   float64 // 起始桩号
	EndStation     float64 // 结束桩号
	LayerIndex     int     // 层序号
	LayerName      string  // 层名
	Thickness      float64 // 厚度
	InnerStepWidth float64 // 内台阶宽
	InnerSlope     float64 // 内坡率（绝对值）
	OuterStepWidth float64 // 外台阶宽
	OuterSlope     float64 // 外坡率（绝对值）
}

// ParseConfigFile 解析结构层配置文件（CSV，跳过注释行和空行）
func ParseConfigFile(filePath string) ([]JiegoucengConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var configs []JiegoucengConfig

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 9 {
			continue
		}
		cfg := JiegoucengConfig{}
		cfg.StartStation, _ = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		cfg.EndStation, _ = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		cfg.LayerIndex, _ = strconv.Atoi(strings.TrimSpace(parts[2]))
		cfg.LayerName = strings.TrimSpace(parts[3])
		cfg.Thickness, _ = strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)
		cfg.InnerStepWidth, _ = strconv.ParseFloat(strings.TrimSpace(parts[5]), 64)
		cfg.InnerSlope, _ = strconv.ParseFloat(strings.TrimSpace(parts[6]), 64)
		cfg.OuterStepWidth, _ = strconv.ParseFloat(strings.TrimSpace(parts[7]), 64)
		cfg.OuterSlope, _ = strconv.ParseFloat(strings.TrimSpace(parts[8]), 64)
		configs = append(configs, cfg)
	}
	return configs, scanner.Err()
}

// ---------- 辅助：筛选与稳定排序 ----------

func filterAndSortConfigs(station float64, configs []JiegoucengConfig) []JiegoucengConfig {
	var active []JiegoucengConfig
	for _, cfg := range configs {
		if station >= cfg.StartStation && station <= cfg.EndStation {
			active = append(active, cfg)
		}
	}
	// 使用稳定排序，保持相同 LayerIndex 的原始顺序（与 C# OrderBy 一致）
	sort.SliceStable(active, func(i, j int) bool {
		return active[i].LayerIndex < active[j].LayerIndex
	})
	return active
}

// ---------- 左侧结构层计算 ----------

// ComputeLeftCoordinates 左侧结构层计算（与 C# LeftJiegoucengManager.ComputeCoordinates 完全一致）
func ComputeLeftCoordinates(
	station float64,
	centerPoint Point2D,
	crossfall float64, // 左幅横坡百分比（如 -2.5）
	configs []JiegoucengConfig,
) (polygons [][][]float64, subgradeLine [][]float64) {
	active := filterAndSortConfigs(station, configs)
	if len(active) == 0 {
		return nil, nil
	}

	k := -crossfall / 100.0 // 左幅横坡斜率

	layerCount := len(active)
	topOut := make([]Point2D, layerCount)
	botOut := make([]Point2D, layerCount)
	topIn := make([]Point2D, layerCount)
	botIn := make([]Point2D, layerCount)

	curTopInX := centerPoint.X + active[0].InnerStepWidth
	curTopInY := centerPoint.Y
	curTopOutX := curTopInX + active[0].OuterStepWidth
	curTopOutY := curTopInY + (curTopOutX-curTopInX)*k

	for i := 0; i < layerCount; i++ {
		cfg := active[i]
		if i > 0 {
			prevBotOut := botOut[i-1]
			prevBotIn := botIn[i-1]

			curTopOutX = prevBotOut.X + cfg.OuterStepWidth
			curTopOutY = prevBotOut.Y + (curTopOutX-prevBotOut.X)*k

			curTopInX = prevBotIn.X + cfg.InnerStepWidth
			curTopInY = prevBotIn.Y + (curTopInX-prevBotIn.X)*k
		}

		topOut[i] = Point2D{X: curTopOutX, Y: curTopOutY}
		topIn[i] = Point2D{X: curTopInX, Y: curTopInY}

		outN := cfg.OuterSlope
		if outN < 0 {
			outN = -outN
		}
		dx := -(outN * cfg.Thickness) / (1.0 - outN*k)
		botOutX := curTopOutX + dx
		botOutY := curTopOutY - cfg.Thickness + dx*k
		botOut[i] = Point2D{X: botOutX, Y: botOutY}

		botIn[i] = Point2D{X: curTopInX, Y: curTopInY - cfg.Thickness}

		poly := [][]float64{
			{topOut[i].X, topOut[i].Y},
			{botOut[i].X, botOut[i].Y},
			{botIn[i].X, botIn[i].Y},
			{topIn[i].X, topIn[i].Y},
		}
		polygons = append(polygons, poly)
	}

	// 组合设计线（左侧：外点先，内点反转后）
	outerList := make([]Point2D, 0, layerCount*2)
	innerList := make([]Point2D, 0, layerCount*2)

	for i := 0; i < layerCount; i++ {
		outerList = append(outerList, topOut[i], botOut[i])
	}
	for i := 0; i < layerCount; i++ {
		innerList = append(innerList, topIn[i], botIn[i])
	}
	innerList = append([]Point2D{{X: topIn[0].X, Y: centerPoint.Y}}, innerList...)
	// 反转 innerList
	for i, j := 0, len(innerList)-1; i < j; i, j = i+1, j-1 {
		innerList[i], innerList[j] = innerList[j], innerList[i]
	}

	combined := make([][]float64, 0, len(outerList)+len(innerList))
	for _, p := range outerList {
		combined = append(combined, []float64{p.X, p.Y})
	}
	for _, p := range innerList {
		combined = append(combined, []float64{p.X, p.Y})
	}
	subgradeLine = combined
	return polygons, subgradeLine
}

// ---------- 右侧结构层计算 ----------

// ComputeRightCoordinates 右侧结构层计算（与 C# RightJiegoucengManager.ComputeCoordinates 完全一致）
func ComputeRightCoordinates(
	station float64,
	centerPoint Point2D,
	crossfall float64, // 右幅横坡百分比（如 -2.5）
	configs []JiegoucengConfig,
) (polygons [][][]float64, subgradeLine [][]float64) {
	active := filterAndSortConfigs(station, configs)
	if len(active) == 0 {
		return nil, nil
	}

	k := crossfall / 100.0 // 右幅横坡斜率

	layerCount := len(active)
	topOut := make([]Point2D, layerCount)
	botOut := make([]Point2D, layerCount)
	topIn := make([]Point2D, layerCount)
	botIn := make([]Point2D, layerCount)

	curTopInX := centerPoint.X + active[0].InnerStepWidth
	curTopInY := centerPoint.Y
	curTopOutX := curTopInX + active[0].OuterStepWidth
	curTopOutY := curTopInY + (curTopOutX-curTopInX)*k

	for i := 0; i < layerCount; i++ {
		cfg := active[i]
		if i > 0 {
			prevBotOut := botOut[i-1]
			prevBotIn := botIn[i-1]

			curTopOutX = prevBotOut.X + cfg.OuterStepWidth
			curTopOutY = prevBotOut.Y + (curTopOutX-prevBotOut.X)*k

			curTopInX = prevBotIn.X + cfg.InnerStepWidth
			curTopInY = prevBotIn.Y + (curTopInX-prevBotIn.X)*k
		}

		topOut[i] = Point2D{X: curTopOutX, Y: curTopOutY}
		topIn[i] = Point2D{X: curTopInX, Y: curTopInY}

		outN := cfg.OuterSlope
		if outN < 0 {
			outN = -outN
		}
		dx := (outN * cfg.Thickness) / (1.0 + outN*k)
		botOutX := curTopOutX + dx
		botOutY := curTopOutY - cfg.Thickness + (botOutX-curTopOutX)*k
		botOut[i] = Point2D{X: botOutX, Y: botOutY}

		botIn[i] = Point2D{X: curTopInX, Y: curTopInY - cfg.Thickness}

		poly := [][]float64{
			{topOut[i].X, topOut[i].Y},
			{botOut[i].X, botOut[i].Y},
			{botIn[i].X, botIn[i].Y},
			{topIn[i].X, topIn[i].Y},
		}
		polygons = append(polygons, poly)
	}

	// 组合设计线（右侧：内点先，外点后反转）
	innerList := make([]Point2D, 0, layerCount*2+1)
	outerList := make([]Point2D, 0, layerCount*2)

	innerList = append(innerList, Point2D{X: topIn[0].X, Y: centerPoint.Y})
	for i := 0; i < layerCount; i++ {
		innerList = append(innerList, topIn[i], botIn[i])
	}
	for i := 0; i < layerCount; i++ {
		outerList = append(outerList, topOut[i], botOut[i])
	}
	// 反转 outerList
	for i, j := 0, len(outerList)-1; i < j; i, j = i+1, j-1 {
		outerList[i], outerList[j] = outerList[j], outerList[i]
	}

	combined := make([][]float64, 0, len(innerList)+len(outerList))
	for _, p := range innerList {
		combined = append(combined, []float64{p.X, p.Y})
	}
	for _, p := range outerList {
		combined = append(combined, []float64{p.X, p.Y})
	}
	subgradeLine = combined
	return polygons, subgradeLine
}

// ---------- 横坡管理 ----------

// CrossfallRecord 横坡记录
type CrossfallRecord struct {
	Station float64
	Slope   float64 // 绝对横坡(%)，例如 -2.5 表示 -2.5%
}

// ParseCrossfallFile 解析横坡文件，跳过首行说明头和注释，按站号升序排列
func ParseCrossfallFile(filePath string) ([]CrossfallRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var records []CrossfallRecord

	// 跳过首行（说明头）
	if !scanner.Scan() {
		return nil, scanner.Err()
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}
		station, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		slope, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err1 != nil || err2 != nil {
			continue
		}
		records = append(records, CrossfallRecord{Station: station, Slope: slope})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	// 按站号升序排序
	sort.Slice(records, func(i, j int) bool {
		return records[i].Station < records[j].Station
	})
	return records, nil
}

// InterpolateCrossfall 线性插值横坡
func InterpolateCrossfall(records []CrossfallRecord, currentStation float64) float64 {
	if len(records) == 0 {
		return 0.0
	}
	if len(records) == 1 {
		return records[0].Slope
	}
	if currentStation <= records[0].Station {
		return records[0].Slope
	}
	if currentStation >= records[len(records)-1].Station {
		return records[len(records)-1].Slope
	}

	// 二分查找
	low, high := 0, len(records)-1
	for low <= high {
		mid := low + ((high - low) >> 1)
		midStation := records[mid].Station
		if math.Abs(midStation-currentStation) < 0.00001 {
			return records[mid].Slope
		}
		if midStation < currentStation {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	left := records[high]
	right := records[low]
	stationDelta := right.Station - left.Station
	if math.Abs(stationDelta) < 0.00001 {
		return left.Slope
	}
	ratio := (currentStation - left.Station) / stationDelta
	return left.Slope + ratio*(right.Slope-left.Slope)
}