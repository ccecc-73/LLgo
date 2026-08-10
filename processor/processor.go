package processor

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"LLhdm/bianpo"
	"LLhdm/config"
	"LLhdm/jiegouceng"
	"LLhdm/llutil"
	"LLhdm/section"
	"LLhdm/terrain"
	"LLhdm/topslope"
)

// Process 处理项目，生成 CSV、DXF 和 HTML
func Process(projectName string) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}
	workingDir = filepath.Join(workingDir, projectName)
	resultDir := filepath.Join(workingDir, "result")
	if err := os.MkdirAll(resultDir, 0755); err != nil {
		return err
	}
	config.LoadConfig(projectName, workingDir)  // ← 这一行

	// 构建文件路径
	xyzPath := filepath.Join(workingDir, projectName+".xyz")
	lt := filepath.Join(workingDir, projectName+".leftslope")
	rt := filepath.Join(workingDir, projectName+".rightslope")
	bp := filepath.Join(workingDir, projectName+".bianpo")
	lSt := filepath.Join(workingDir, projectName+".leftstructure")
	rSt := filepath.Join(workingDir, projectName+".rightstructure")
	lCf := filepath.Join(workingDir, projectName+".leftcrossfall")
	rCf := filepath.Join(workingDir, projectName+".rightcrossfall")
	pqxPath := filepath.Join(workingDir, projectName+".pqx")
	sqxPath := filepath.Join(workingDir, projectName+".sqx")
	kzbPath := filepath.Join(workingDir, projectName+".k")
	csvPath := filepath.Join(resultDir, projectName+".csv")
	dxfPath := filepath.Join(resultDir, projectName+".dxf")
	htmlPath := filepath.Join(resultDir, projectName+".html")

	// 加载地形
	mesh, err := terrain.FromTextFile(xyzPath, float64(config.XyzRows), float64(config.XyzCols))
	if err != nil {
		return fmt.Errorf("加载地形失败: %v", err)
	}

	// 加载平曲线、竖曲线、桩号表
	pqx, err := llutil.ReadDataFromFile(pqxPath, 8, "")
	if err != nil {
		return fmt.Errorf("加载平曲线失败: %v", err)
	}
	sqx, err := llutil.ReadDataFromFile(sqxPath, 3, "")
	if err != nil {
		return fmt.Errorf("加载竖曲线失败: %v", err)
	}
	kzb, err := llutil.ReadDataFromFile(kzbPath, 3, "")
	if err != nil {
		return fmt.Errorf("加载桩号表失败: %v", err)
	}

	// 加载各模块数据
	leftWidths, err := topslope.ParseWidthFile(lt)
	if err != nil {
		return fmt.Errorf("加载左侧宽度失败: %v", err)
	}
	rightWidths, err := topslope.ParseWidthFile(rt)
	if err != nil {
		return fmt.Errorf("加载右侧宽度失败: %v", err)
	}
	slopeData, err := bianpo.ParseBianPo(bp)
	if err != nil {
		return fmt.Errorf("加载边坡失败: %v", err)
	}
	leftCrossfalls, err := jiegouceng.ParseCrossfallFile(lCf)
	if err != nil {
		return fmt.Errorf("加载左侧横坡失败: %v", err)
	}
	rightCrossfalls, err := jiegouceng.ParseCrossfallFile(rCf)
	if err != nil {
		return fmt.Errorf("加载右侧横坡失败: %v", err)
	}
	leftStructures, err := jiegouceng.ParseConfigFile(lSt)
	if err != nil {
		return fmt.Errorf("加载左侧结构层失败: %v", err)
	}
	rightStructures, err := jiegouceng.ParseConfigFile(rSt)
	if err != nil {
		return fmt.Errorf("加载右侧结构层失败: %v", err)
	}

	// 创建计算器
	calculator := section.NewSectionCalculator(
		mesh, pqx, sqx, leftWidths, rightWidths,
		slopeData, leftCrossfalls, rightCrossfalls,
		leftStructures, rightStructures,
	)

	// CSV 表头
	csvHeader := config.BuildHeader(leftStructures, rightStructures)
	var csvSb strings.Builder
	csvSb.WriteString(csvHeader)

	// DXF 实体
	var dxfEntities strings.Builder

	// HTML 数据收集
	var htmlSections []config.SectionData

	currentOffsetX := 0.0

	// 遍历桩号
	for i := 0; i < len(kzb); i++ {
		st := kzb[i][0]
		kLeft := kzb[i][1]
		kRight := kzb[i][2]

		result := calculator.Compute(st, kLeft, kRight)
		if result == nil {
			continue
		}

		appendSectionToDxf(&dxfEntities, result, &currentOffsetX)

		// 构建设置层面积 CSV
		var layerCsv strings.Builder
		for _, areaText := range result.LayerAreaTexts {
			parts := strings.Split(areaText, ":")
			if len(parts) >= 3 {
				area, _ := strconv.ParseFloat(parts[2], 64)
				layerCsv.WriteString(fmt.Sprintf(",%.3f", area))
			}
		}

		stationStr := llutil.Hua_Num2K(st)
		row := config.BuildDataRow(
			stationStr,
			result.FillArea,
			result.CutArea,
			result.ClearArea,
			result.MinX,
			result.MaxX,
			layerCsv.String(),
		)
		csvSb.WriteString(row + "\n")

		// 收集 HTML 断面数据
		htmlSections = append(htmlSections, config.SectionData{
	Station:          stationStr,
	FillArea:         result.FillArea,
	CutArea:          result.CutArea,
	ClearArea:        result.ClearArea,
	FinalDesign:      result.FinalDesign,
	Ground:           result.Ground,
	Cleared:          result.Cleared,
	LeftSubgrade:     result.LeftSubgrade,
	RightSubgrade:    result.RightSubgrade,
	LayerPolygons:    result.LayerPolygons,
	LayerAreaTexts:   result.LayerAreaTexts,
	LeftSlopePoints:  result.LeftSlopePoints,
	RightSlopePoints: result.RightSlopePoints,
	CenterY:          result.CenterY,
	LOuterX:          result.LOuterX,
	LOuterY:          result.LOuterY,
	ROuterX:          result.ROuterX,
	ROuterY:          result.ROuterY,
	LeftCrossfall:    result.LeftCrossfall,
	RightCrossfall:   result.RightCrossfall,
	LeftToeX:         result.LeftToeX,
	LeftToeY:         result.LeftToeY,
	RightToeX:        result.RightToeX,
	RightToeY:        result.RightToeY,
})
	}

	// 写入 CSV
	if err := os.WriteFile(csvPath, []byte(csvSb.String()), 0644); err != nil {
		return fmt.Errorf("写入 CSV 失败: %v", err)
	}

	// 写入 DXF
	dxfContent := config.BuildDxfFromBuilder(&dxfEntities)
	if err := os.WriteFile(dxfPath, []byte(dxfContent), 0644); err != nil {
		return fmt.Errorf("写入 DXF 失败: %v", err)
	}

	// 写入 HTML
	htmlContent := config.BuildHTML(projectName, htmlSections)
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("写入 HTML 失败: %v", err)
	}

	fmt.Printf("%s: 处理完成\n", projectName)
	return nil
}

// ========== DXF 绘制辅助 ==========

func appendSectionToDxf(dxf *strings.Builder, r *section.SectionResult, currentOffsetX *float64) {
	width := r.MaxX - r.MinX
	if width <= 0 {
		width = 40.0
	}
	offX := *currentOffsetX - r.MinX
	offY := 0.0

	// 绘制各条线
	llutil.AppendDxfLwPolyline(dxf, r.FinalDesign, offX, offY, "DESIGN_ROAD_TRIMMED", 1)
	llutil.AppendDxfLwPolyline(dxf, r.Ground, offX, offY, "NATURAL_GROUND", 3)
	llutil.AppendDxfLwPolyline(dxf, r.Cleared, offX, offY, "CLEAR_GROUND", 8)

	if len(r.FinalFinished) > 0 {
		llutil.AppendDxfLwPolyline(dxf, r.FinalFinished, offX, offY, "FINISHED_ROAD", 2)
	}
	if len(r.LeftSubgrade) > 0 {
		llutil.AppendDxfLwPolyline(dxf, r.LeftSubgrade, offX, offY, "STRUCTURE_LEFT", 4)
	}
	if len(r.RightSubgrade) > 0 {
		llutil.AppendDxfLwPolyline(dxf, r.RightSubgrade, offX, offY, "STRUCTURE_RIGHT", 6)
	}

	// 结构层多边形
	for i, poly := range r.LayerPolygons {
		color := 4
		if i >= len(r.LayerPolygons)/2 {
			color = 6
		}
		llutil.AppendDxfLwPolyline(dxf, poly, offX, offY, fmt.Sprintf("LAYER_%d", i), color)
	}

	// 信息文本
	station := llutil.Hua_Num2K(r.Station)
	infoTexts := []string{
		fmt.Sprintf("AT:%.3f", r.FillArea),
		fmt.Sprintf("AW:%.3f", r.CutArea),
		fmt.Sprintf("Topsoil:%.3f", r.ClearArea),
	}
	infoTexts = append(infoTexts, r.LayerAreaTexts...)

	rawW := width * 1.25
	estH := float64(len(infoTexts))*(rawW*0.015*config.LineSpacingFactor) + rawW*0.015*5.0
	rawH := (r.CenterY - r.MinY + estH) * 1.35

	scale := rawW / config.A4WidthMM
	if hScale := rawH / config.A4HeightMM; hScale > scale {
		scale = hScale
	}
	if scale < 0.1 {
		scale = 0.1
	}
	textScale := scale * config.TextScaleFactor
	lineSpacing := textScale * config.LineSpacingFactor

	boxW := config.A4WidthMM * scale
	boxH := config.A4HeightMM * scale
	boxMinX := offX - boxW/2
	boxMinY := r.MinY - textScale*4 - float64(len(infoTexts))*lineSpacing - textScale*3

	// A4 边框
	frame := [][]float64{
		{boxMinX, boxMinY},
		{boxMinX + boxW, boxMinY},
		{boxMinX + boxW, boxMinY + boxH},
		{boxMinX, boxMinY + boxH},
		{boxMinX, boxMinY},
	}
	llutil.AppendDxfLwPolyline(dxf, frame, 0, 0, "A4_BORDER", 7)

	// 桩号标题
	llutil.AppendDxfTextCenter(dxf, station, offX, boxMinY+textScale*25, textScale*1.2, "STATION_TITLE", 0)

	// 中桩高程
	llutil.AppendDxfTextCenter(dxf, fmt.Sprintf("%.3f", r.CenterY), offX, r.CenterY+textScale*1.2, textScale, "POINT_TEXT_H", 90)

	// 边桩高程
	llutil.AppendDxfTextCenter(dxf, fmt.Sprintf("%.3f", r.LOuterY), offX+r.LOuterX, r.LOuterY+textScale*1.2, textScale, "LEFT_TEXT_H", 90)
	llutil.AppendDxfTextCenter(dxf, fmt.Sprintf("%.3f", r.ROuterY), offX+r.ROuterX, r.ROuterY+textScale*1.2, textScale, "RIGHT_TEXT_H", 90)

	// 横坡标注
	lcf := r.LeftCrossfall
	if lcf < 0.2 && lcf > -0.2 {
		lcf = lcf * 100
	}
	rcf := r.RightCrossfall
	if rcf < 0.2 && rcf > -0.2 {
		rcf = rcf * 100
	}
	llutil.AppendDxfTextCenter(dxf, fmt.Sprintf("%.2f%%", lcf), offX+r.LOuterX/2, r.CenterY+absFloat(r.LOuterX/2)*r.LeftCrossfall/100+textScale*0.8, textScale, "SLOPE_TEXT", 0)
	llutil.AppendDxfTextCenter(dxf, fmt.Sprintf("%.2f%%", rcf), offX+r.ROuterX/2, r.CenterY+absFloat(r.ROuterX/2)*r.RightCrossfall/100+textScale*0.8, textScale, "SLOPE_TEXT", 0)

	// 边坡坡率标注
	annotateSlope(dxf, r.LeftSlopePoints, r.LeftToeX, r.LeftToeY, r.LOuterX, r.LOuterY, offX, textScale, true)
	annotateSlope(dxf, r.RightSlopePoints, r.ROuterX, r.ROuterY, r.RightToeX, r.RightToeY, offX, textScale, false)

	// 信息文本列表
	tx := offX - textScale*4.5
	baseY := boxMinY + textScale*25 - textScale*2
	for i, txt := range infoTexts {
		llutil.AppendDxfText(dxf, txt, tx, baseY-float64(i)*lineSpacing, textScale, "INFO_TEXT")
	}

	*currentOffsetX += boxW + config.A4Gap
}

func annotateSlope(dxf *strings.Builder, slopePoints [][]float64, startX, startY, endX, endY, offX, textScale float64, isLeft bool) {
	if len(slopePoints) < 2 {
		return
	}
	pts := make([][]float64, 0)
	pts = append(pts, []float64{startX, startY})
	for _, p := range slopePoints {
		if isLeft && p[0] > startX {
			pts = append(pts, p)
		} else if !isLeft && p[0] < startX {
			pts = append(pts, p)
		}
	}
	pts = append(pts, []float64{endX, endY})

	for i := 0; i < len(pts)-1; i++ {
		x1, y1 := pts[i][0], pts[i][1]
		x2, y2 := pts[i+1][0], pts[i+1][1]
		dx := x2 - x1
		dy := y2 - y1
		if absFloat(dy) > 0.01 && absFloat(dx) > 0.01 {
			ratio := absFloat(dx) / absFloat(dy)
			midX := offX + (x1+x2)/2.0
			midY := (y1 + y2) / 2.0
			angleRad := math.Atan2(dy, dx)
			angleDeg := angleRad * 180.0 / math.Pi
			llutil.AppendDxfTextSlope(dxf, fmt.Sprintf("1:%.2f", ratio), midX, midY, textScale, angleDeg, "SLOPE_TEXT")
		}
	}
}

// ========== 辅助函数 ==========

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}