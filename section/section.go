package section

import (
	"LLhdm/bianpo"
	"LLhdm/config"
	"LLhdm/jiegouceng"
	"LLhdm/llutil"
	"LLhdm/terrain"
	"LLhdm/topslope"
	"strconv"
)

// SectionResult 横断面计算结果
type SectionResult struct {
	Station         float64
	CenterY         float64
	LOuterX, LOuterY float64
	ROuterX, ROuterY float64
	LeftCrossfall   float64
	RightCrossfall  float64
	FinalDesign     [][]float64
	FinalFinished   [][]float64
	Ground          [][]float64
	Cleared         [][]float64
	LeftSubgrade    [][]float64
	RightSubgrade   [][]float64
	LayerPolygons   [][][]float64
	LayerAreaTexts  []string
	FillArea        float64
	CutArea         float64
	ClearArea       float64
	MinX, MaxX, MinY float64
	LeftSlopePoints  [][]float64
	RightSlopePoints [][]float64
	LeftToeX, LeftToeY, RightToeX, RightToeY float64
}

// SectionCalculator 横断面计算器
type SectionCalculator struct {
	mesh            *terrain.TerrainMesh
	pqx             [][]float64
	sqx             [][]float64
	leftWidths      []topslope.DuanMianShuJu
	rightWidths     []topslope.DuanMianShuJu
	slopeData       []*bianpo.BianPoDuanLuo
	leftCrossfalls  []jiegouceng.CrossfallRecord
	rightCrossfalls []jiegouceng.CrossfallRecord
	leftStructures  []jiegouceng.JiegoucengConfig
	rightStructures []jiegouceng.JiegoucengConfig
}

// NewSectionCalculator 构造函数
func NewSectionCalculator(
	mesh *terrain.TerrainMesh,
	pqx [][]float64,
	sqx [][]float64,
	leftWidths []topslope.DuanMianShuJu,
	rightWidths []topslope.DuanMianShuJu,
	slopeData []*bianpo.BianPoDuanLuo,
	leftCrossfalls []jiegouceng.CrossfallRecord,
	rightCrossfalls []jiegouceng.CrossfallRecord,
	leftStructures []jiegouceng.JiegoucengConfig,
	rightStructures []jiegouceng.JiegoucengConfig,
) *SectionCalculator {
	return &SectionCalculator{
		mesh:            mesh,
		pqx:             pqx,
		sqx:             sqx,
		leftWidths:      leftWidths,
		rightWidths:     rightWidths,
		slopeData:       slopeData,
		leftCrossfalls:  leftCrossfalls,
		rightCrossfalls: rightCrossfalls,
		leftStructures:  leftStructures,
		rightStructures: rightStructures,
	}
}

// Compute 计算指定桩号的横断面
func (sc *SectionCalculator) Compute(station, kLeft, kRight float64) *SectionResult {
	// 1. 获取中桩高程
	centerY := llutil.Hua_H(sc.sqx, station)

	// 2. 计算车道坐标
	dao := topslope.GetCrossSectionXY(station, centerY, sc.leftWidths, sc.rightWidths)
	if len(dao.ZuoCeCheDaoJueDui) == 0 || len(dao.YouCeCheDaoJueDui) == 0 {
		return nil
	}

	// 3. 获取左右外边缘点
	lOuterX := dao.ZuoCeCheDaoJueDui[0].WidthX
	lOuterY := dao.ZuoCeCheDaoJueDui[0].GaoChengY
	rOuterX := dao.YouCeCheDaoJueDui[len(dao.YouCeCheDaoJueDui)-1].WidthX
	rOuterY := dao.YouCeCheDaoJueDui[len(dao.YouCeCheDaoJueDui)-1].GaoChengY

	// 4. 计算中线左右偏移后的端点（用于地形切割）
	lseg := llutil.Hua_Dantiaoxianludange(sc.pqx, station, kLeft, 90)
	rseg := llutil.Hua_Dantiaoxianludange(sc.pqx, station, kRight, 90)
	segment := [][]float64{{lseg[0], lseg[1], rseg[0], rseg[1]}}

	// 5. 切割地形获取地面线
	ground := sc.mesh.ExtractSections(segment)
	if len(ground) == 0 {
		return nil
	}

	// 6. 清表线
	cleared := llutil.Hua_OffsetPolyline(ground, config.ClearDepth)

	// 7. 获取边坡参数
	bpCfg := bianpo.K2BianPo(sc.slopeData, station)
	if bpCfg == nil {
		return nil
	}
	candidates := bianpo.GetAbsolute(bpCfg, lOuterX, lOuterY, rOuterX, rOuterY)

	// 8. 判断填挖并选择对应的边坡点
	lGroundY := llutil.FromXgetY(ground, lOuterX)
	var finalLeft [][]float64
	if lOuterY > lGroundY {
		finalLeft = candidates.ZuoTianJueDui
	} else {
		finalLeft = candidates.ZuoWaJueDui
	}

	rGroundY := llutil.FromXgetY(ground, rOuterX)
	var finalRight [][]float64
	if rOuterY > rGroundY {
		finalRight = candidates.YouTianJueDui
	} else {
		finalRight = candidates.YouWaJueDui
	}

	// 9. 构建设计线（原始设计线，不含结构层）
	rawDesign := make([][]float64, 0)
	if finalLeft != nil {
		rawDesign = append(rawDesign, finalLeft...)
	}
	for _, p := range dao.ZuoCeCheDaoJueDui {
		rawDesign = append(rawDesign, []float64{p.WidthX, p.GaoChengY})
	}
	rawDesign = append(rawDesign, []float64{0.0, centerY})
	for _, p := range dao.YouCeCheDaoJueDui {
		rawDesign = append(rawDesign, []float64{p.WidthX, p.GaoChengY})
	}
	if finalRight != nil {
		rawDesign = append(rawDesign, finalRight...)
	}
	designTop := toMat(rawDesign)

	// 10. 左侧结构层计算
	leftCrossfall := jiegouceng.InterpolateCrossfall(sc.leftCrossfalls, station)
	leftCenter := jiegouceng.Point2D{X: 0.0, Y: centerY}
	leftPolygons, leftSubgrade := jiegouceng.ComputeLeftCoordinates(station, leftCenter, leftCrossfall, sc.leftStructures)

	if len(leftSubgrade) == 0 {
		return nil
	}
	leftInnerX := leftSubgrade[len(leftSubgrade)-1][0]
	newH := llutil.FromXgetY(designTop, leftInnerX)
	leftCenter = jiegouceng.Point2D{X: 0.0, Y: newH}
	leftPolygons, leftSubgrade = jiegouceng.ComputeLeftCoordinates(station, leftCenter, leftCrossfall, sc.leftStructures)

	// 11. 右侧结构层计算
	rightCrossfall := jiegouceng.InterpolateCrossfall(sc.rightCrossfalls, station)
	rightCenter := jiegouceng.Point2D{X: 0.0, Y: centerY}
	rightPolygons, rightSubgrade := jiegouceng.ComputeRightCoordinates(station, rightCenter, rightCrossfall, sc.rightStructures)

	if len(rightSubgrade) == 0 {
		return nil
	}
	rightInnerX := rightSubgrade[0][0]
	newH = llutil.FromXgetY(designTop, rightInnerX)
	rightCenter = jiegouceng.Point2D{X: 0.0, Y: newH}
	rightPolygons, rightSubgrade = jiegouceng.ComputeRightCoordinates(station, rightCenter, rightCrossfall, sc.rightStructures)

	// 12. 获取结构层内外锚点
	var leftOuterAnchor, rightOuterAnchor float64
	if len(leftSubgrade) > 0 {
		leftOuterAnchor = leftSubgrade[0][0]
	} else {
		leftOuterAnchor = lOuterX
	}
	if len(rightSubgrade) > 0 {
		rightOuterAnchor = rightSubgrade[len(rightSubgrade)-1][0]
	} else {
		rightOuterAnchor = rOuterX
	}

	// 13. 组装最终设计线和完成面线
	designList := make([][]float64, 0)
	finishedList := make([][]float64, 0)

	// 左侧外边坡部分（从原始设计线中截取）
	addOuterPart(rawDesign, &designList, &finishedList, leftOuterAnchor, -1)

	// 左侧结构层
	for _, p := range leftSubgrade {
		designList = append(designList, []float64{p[0], p[1]})
	}
	// 左侧路面
	for _, p := range dao.ZuoCeCheDaoJueDui {
		finishedList = append(finishedList, []float64{p.WidthX, p.GaoChengY})
	}

	// 中间部分（路面内部）
	leftInnerAnchor := leftSubgrade[len(leftSubgrade)-1][0]
	rightInnerAnchor := rightSubgrade[0][0]
	for _, p := range rawDesign {
		if p[0] >= leftInnerAnchor && p[0] <= rightInnerAnchor {
			designList = append(designList, p)
			finishedList = append(finishedList, p)
		}
	}

	// 右侧结构层
	for _, p := range rightSubgrade {
		designList = append(designList, []float64{p[0], p[1]})
	}
	// 右侧路面
	for _, p := range dao.YouCeCheDaoJueDui {
		finishedList = append(finishedList, []float64{p.WidthX, p.GaoChengY})
	}

	// 右侧外边坡部分
	addOuterPart(rawDesign, &designList, &finishedList, rightOuterAnchor, 1)

	designMat := toMat(designList)

	// 14. 计算填挖面积
	res := llutil.Hua_CutAndFillArea(cleared, designMat, 8)
	fill := abs(res[0])
	cut := abs(res[1])
	minX := res[2]
	maxX := res[3]
	minY := res[4]

	// 15. 计算清表面积
	clearPoly := llutil.BuildClearPolygon(ground, cleared, minX, maxX)
	clearArea := llutil.Hua_PolygonArea(clearPoly)

	// 16. 提取最终设计线（从 res 中）
	count := int(res[6])
	finalDesign := make([][]float64, count)
	idx := 7
	for i := 0; i < count; i++ {
		finalDesign[i] = []float64{res[idx], res[idx+1]}
		idx += 2
	}

	// 17. 裁剪完成面线到设计线范围
	minTrim := finalDesign[0][0]
	maxTrim := finalDesign[len(finalDesign)-1][0]
	trimmedFinished := make([][]float64, 0)
	for _, p := range finishedList {
		if p[0] >= minTrim && p[0] <= maxTrim {
			trimmedFinished = append(trimmedFinished, p)
		}
	}
	finalFinished := toMat(trimmedFinished)

	// 18. 计算各结构层面积
	layerAreaTexts := make([]string, 0)
	layerPolygons := make([][][]float64, 0)

	for i := 0; i < len(leftPolygons); i++ {
		poly := leftPolygons[i]
		area := abs(llutil.Hua_PolygonArea(poly))
		name := ""
		if i < len(sc.leftStructures) && sc.leftStructures[i].LayerName != "" {
			name = sc.leftStructures[i].LayerName
		} else {
			name = "L_Lay" + string(rune(i+1))
		}
		layerAreaTexts = append(layerAreaTexts, "L:"+name+":"+formatFloat(area, 3))
		layerPolygons = append(layerPolygons, poly)
	}

	for i := 0; i < len(rightPolygons); i++ {
		poly := rightPolygons[i]
		area := abs(llutil.Hua_PolygonArea(poly))
		name := ""
		if i < len(sc.rightStructures) && sc.rightStructures[i].LayerName != "" {
			name = sc.rightStructures[i].LayerName
		} else {
			name = "R_Lay" + string(rune(i+1))
		}
		layerAreaTexts = append(layerAreaTexts, "R:"+name+":"+formatFloat(area, 3))
		layerPolygons = append(layerPolygons, poly)
	}

	// 19. 返回结果
	return &SectionResult{
		Station:          station,
		CenterY:          centerY,
		LOuterX:          lOuterX,
		LOuterY:          lOuterY,
		ROuterX:          rOuterX,
		ROuterY:          rOuterY,
		LeftCrossfall:    leftCrossfall,
		RightCrossfall:   rightCrossfall,
		FinalDesign:      finalDesign,
		FinalFinished:    finalFinished,
		Ground:           ground,
		Cleared:          cleared,
		LeftSubgrade:     leftSubgrade,
		RightSubgrade:    rightSubgrade,
		LayerPolygons:    layerPolygons,
		LayerAreaTexts:   layerAreaTexts,
		FillArea:         fill,
		CutArea:          cut,
		ClearArea:        clearArea,
		MinX:             minX,
		MaxX:             maxX,
		MinY:             minY,
		LeftSlopePoints:  finalLeft,
		RightSlopePoints: finalRight,
		LeftToeX:         finalDesign[0][0],
		LeftToeY:         finalDesign[0][1],
		RightToeX:        finalDesign[len(finalDesign)-1][0],
		RightToeY:        finalDesign[len(finalDesign)-1][1],
	}
}

// ========== 辅助函数 ==========

func toMat(list [][]float64) [][]float64 {
	if len(list) == 0 {
		return [][]float64{}
	}
	result := make([][]float64, len(list))
	for i, p := range list {
		result[i] = []float64{p[0], p[1]}
	}
	return result
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func formatFloat(v float64, prec int) string {
	return strconv.FormatFloat(v, 'f', prec, 64)
}

// addOuterPart 添加外边坡部分（从原始设计线中截取）
func addOuterPart(raw [][]float64, design, finished *[][]float64, anchor float64, side int) {
	for i := 0; i < len(raw); i++ {
		p := raw[i]
		condition := false
		if side < 0 {
			condition = p[0] <= anchor
		} else {
			condition = p[0] >= anchor
		}
		if condition {
			*design = append(*design, p)
			*finished = append(*finished, p)
		}
		if i > 0 {
			prev := raw[i-1]
			var cross bool
			if side < 0 {
				cross = prev[0] < anchor && p[0] > anchor
			} else {
				cross = prev[0] > anchor && p[0] < anchor
			}
			if cross {
				x1, y1 := prev[0], prev[1]
				x2, y2 := p[0], p[1]
				y := y1 + (y2-y1)*(anchor-x1)/(x2-x1)
				pt := []float64{anchor, y}
				*design = append(*design, pt)
				*finished = append(*finished, pt)
			}
		}
	}
}