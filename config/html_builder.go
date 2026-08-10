package config

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// SectionData 断面数据
type SectionData struct {
	Station          string
	FillArea         float64
	CutArea          float64
	ClearArea        float64
	FinalDesign      [][]float64
	Ground           [][]float64
	Cleared          [][]float64
	LeftSubgrade     [][]float64
	RightSubgrade    [][]float64
	LayerPolygons    [][][]float64
	LayerAreaTexts   []string
	LeftSlopePoints  [][]float64
	RightSlopePoints [][]float64
	CenterY          float64
	LOuterX          float64
	LOuterY          float64
	ROuterX          float64
	ROuterY          float64
	LeftCrossfall    float64
	RightCrossfall   float64
	LeftToeX         float64
	LeftToeY         float64
	RightToeX        float64
	RightToeY        float64
}

// BuildHTML 生成完整 HTML
func BuildHTML(projectName string, sections []SectionData) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + projectName + ` 横断面成果图</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: 'Segoe UI', 'Microsoft YaHei', sans-serif;
            background: #e8ecf1;
            padding: 20px;
        }
        .header {
            background: white;
            padding: 20px 30px;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            margin-bottom: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
        }
        .header h1 { color: #1a2a3a; font-size: 24px; }
        .header .info { color: #6a7a8a; font-size: 14px; }
        .controls {
            display: flex;
            gap: 12px;
            align-items: center;
            flex-wrap: wrap;
        }
        .controls input {
            padding: 6px 14px;
            border: 1px solid #c0d0e0;
            border-radius: 6px;
            font-size: 14px;
            width: 160px;
        }
        .controls button {
            padding: 6px 18px;
            background: #3b7cff;
            color: white;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
        }
        .controls button:hover { background: #2a6ae0; }
        .section-container {
            background: white;
            margin: 16px 0;
            padding: 20px 25px;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.08);
        }
        .section-title {
            font-size: 18px;
            font-weight: 600;
            color: #1a2a3a;
            margin-bottom: 8px;
        }
        .section-stats {
            display: flex;
            flex-wrap: wrap;
            gap: 20px;
            font-size: 14px;
            color: #4a5a6a;
            margin-bottom: 12px;
            padding-bottom: 10px;
            border-bottom: 1px solid #eef2f7;
        }
        .section-stats .fill { color: #2e7d32; }
        .section-stats .cut { color: #c62828; }
        .section-stats .clear { color: #f9a825; }
        .section-stats span { font-weight: 500; }
        svg {
            display: block;
            width: 100%;
            height: auto;
            background: #fafcfe;
            border-radius: 8px;
            border: 1px solid #e8ecf1;
        }
        .legend {
            display: flex;
            flex-wrap: wrap;
            gap: 18px;
            padding: 10px 0 4px 0;
            font-size: 13px;
            color: #3a4a5a;
        }
        .legend-item {
            display: flex;
            align-items: center;
            gap: 6px;
        }
        .legend-color {
            width: 28px;
            height: 3px;
            border-radius: 2px;
        }
        .footer {
            text-align: center;
            color: #8a9aaa;
            font-size: 12px;
            padding: 20px 0 10px 0;
        }

        /* ===== 文字大小控制面板 ===== */
        .text-control {
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 999;
            background: white;
            border-radius: 12px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.15);
            padding: 12px 14px;
            border: 1px solid #e0e6ed;
        }
        .text-control .toggle-btn {
            background: none;
            border: none;
            font-size: 24px;
            cursor: pointer;
            padding: 4px 8px;
            border-radius: 50%;
            transition: background 0.2s;
            width: 40px;
            height: 40px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #3a4a5a;
        }
        .text-control .toggle-btn:hover {
            background: #eef2f7;
        }
        .text-control .panel {
            display: none;
            margin-top: 10px;
            padding-top: 10px;
            border-top: 1px solid #eef2f7;
        }
        .text-control .panel.open {
            display: block;
        }
        .text-control .panel-title {
            font-weight: 600;
            font-size: 14px;
            color: #1a2a3a;
            text-align: center;
            margin-bottom: 12px;
        }
        .text-control .control-group {
            margin-bottom: 10px;
        }
        .text-control .control-group:last-child {
            margin-bottom: 0;
        }
        .text-control .control-group label {
            display: flex;
            justify-content: space-between;
            font-size: 13px;
            color: #3a4a5a;
            font-weight: 500;
        }
        .text-control .control-group input[type="range"] {
            width: 100%;
            margin: 2px 0;
            accent-color: #3b7cff;
        }
        .text-control .control-group .value {
            color: #3b7cff;
            font-weight: 600;
        }
        .text-control .reset-btn {
            background: #eef2f7;
            border: none;
            padding: 4px 14px;
            border-radius: 4px;
            font-size: 12px;
            cursor: pointer;
            color: #3a4a5a;
            margin-top: 4px;
            width: 100%;
        }
        .text-control .reset-btn:hover {
            background: #dce0e8;
        }

        @media (max-width: 600px) {
            .text-control {
                top: 10px;
                right: 10px;
                padding: 8px 10px;
            }
            .text-control .toggle-btn {
                font-size: 20px;
                width: 32px;
                height: 32px;
            }
            .header { flex-direction: column; align-items: flex-start; gap: 10px; }
            .controls { width: 100%; }
            .controls input { flex: 1; }
        }
    </style>
</head>
<body>
    <!-- ===== 文字大小控制面板 ===== -->
    <div class="text-control" id="textControl">
        <button class="toggle-btn" onclick="togglePanel()" title="调整文字大小">⚙️</button>
        <div class="panel" id="controlPanel">
            <div class="panel-title">🔤 调整文字大小</div>
            <div class="control-group">
                <label>📌 桩号 <span class="value" id="valStation">13</span></label>
                <input type="range" id="sizeStation" min="6" max="30" value="13" step="0.5" oninput="updateSize('station', this.value)">
            </div>
            <div class="control-group">
                <label>📐 设计高程 <span class="value" id="valElevation">11</span></label>
                <input type="range" id="sizeElevation" min="6" max="30" value="11" step="0.5" oninput="updateSize('elevation', this.value)">
            </div>
            <div class="control-group">
                <label>📏 坡比 <span class="value" id="valSlope">9</span></label>
                <input type="range" id="sizeSlope" min="5" max="24" value="9" step="0.5" oninput="updateSize('slope', this.value)">
            </div>
            <div class="control-group">
                <label>📊 横坡 <span class="value" id="valCrossfall">10</span></label>
                <input type="range" id="sizeCrossfall" min="6" max="24" value="10" step="0.5" oninput="updateSize('crossfall', this.value)">
            </div>
            <button class="reset-btn" onclick="resetSizes()">恢复默认</button>
        </div>
    </div>

    <div class="header">
        <div>
            <h1>📐 ` + projectName + ` 横断面成果图</h1>
            <div class="info">共 ` + fmt.Sprintf("%d", len(sections)) + ` 个断面 | 生成时间: ` + time.Now().Format("2006-01-02 15:04:05") + `</div>
        </div>
        <div class="controls">
            <input type="text" id="searchInput" placeholder="🔍 搜索桩号..." oninput="filterSections()">
            <button onclick="scrollToTop()">⬆ 回到顶部</button>
        </div>
    </div>
    <div id="sectionsContainer">
`)

	for _, sec := range sections {
		sb.WriteString(sec.toHTML())
	}

	sb.WriteString(`
    </div>
    <div class="legend" style="background:white;padding:12px 25px;border-radius:12px;margin-top:16px;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
        <div class="legend-item"><span class="legend-color" style="background:#d32f2f;"></span> 地面线</div>
        <div class="legend-item"><span class="legend-color" style="background:#1976d2;"></span> 设计线</div>
        <div class="legend-item"><span class="legend-color" style="background:#388e3c;"></span> 清表线</div>
        <div class="legend-item"><span class="legend-color" style="background:#f57c00;"></span> 结构层</div>
        <div class="legend-item"><span class="legend-color" style="background:#7b1fa2;"></span> 结构层填充</div>
    </div>
    <div class="footer">生成于测绘自动化系统 v1.0</div>
    <script>
        // ===== 文字大小控制 =====
        const SIZE_KEY = 'llhdm_text_sizes';

        function getDefaultSizes() {
            return { station: 13, elevation: 11, slope: 9, crossfall: 10 };
        }

        function loadSizes() {
            try {
                const saved = localStorage.getItem(SIZE_KEY);
                if (saved) {
                    const parsed = JSON.parse(saved);
                    const def = getDefaultSizes();
                    return { ...def, ...parsed };
                }
            } catch (e) {}
            return getDefaultSizes();
        }

        function saveSizes(sizes) {
            localStorage.setItem(SIZE_KEY, JSON.stringify(sizes));
        }

        function applySizes(sizes) {
            document.getElementById('valStation').textContent = sizes.station;
            document.getElementById('valElevation').textContent = sizes.elevation;
            document.getElementById('valSlope').textContent = sizes.slope;
            document.getElementById('valCrossfall').textContent = sizes.crossfall;
            document.getElementById('sizeStation').value = sizes.station;
            document.getElementById('sizeElevation').value = sizes.elevation;
            document.getElementById('sizeSlope').value = sizes.slope;
            document.getElementById('sizeCrossfall').value = sizes.crossfall;

            document.querySelectorAll('.text-station').forEach(el => el.style.fontSize = sizes.station + 'px');
            document.querySelectorAll('.text-elevation').forEach(el => el.style.fontSize = sizes.elevation + 'px');
            document.querySelectorAll('.text-slope').forEach(el => el.style.fontSize = sizes.slope + 'px');
            document.querySelectorAll('.text-crossfall').forEach(el => el.style.fontSize = sizes.crossfall + 'px');
        }

        function updateSize(type, value) {
            const sizes = loadSizes();
            sizes[type] = parseFloat(value);
            saveSizes(sizes);
            applySizes(sizes);
        }

        function resetSizes() {
            const def = getDefaultSizes();
            saveSizes(def);
            applySizes(def);
        }

        function togglePanel() {
            const panel = document.getElementById('controlPanel');
            panel.classList.toggle('open');
        }

        document.addEventListener('DOMContentLoaded', function() {
            const sizes = loadSizes();
            applySizes(sizes);
        });

        // ===== 搜索过滤 =====
        function filterSections() {
            const keyword = document.getElementById('searchInput').value.trim().toLowerCase();
            const containers = document.querySelectorAll('.section-container');
            containers.forEach(el => {
                const title = el.querySelector('.section-title').textContent.toLowerCase();
                el.style.display = (!keyword || title.includes(keyword)) ? 'block' : 'none';
            });
        }

        function scrollToTop() { window.scrollTo({ top: 0, behavior: 'smooth' }); }
    </script>
</body>
</html>`)
	return sb.String()
}

func (s SectionData) toHTML() string {
	var sb strings.Builder
	sb.WriteString(`<div class="section-container" data-station="` + s.Station + `">`)
	sb.WriteString(`<div class="section-title">📍 桩号: ` + s.Station + `</div>`)
	sb.WriteString(`<div class="section-stats">`)
	sb.WriteString(`<span class="fill">🟢 填方: ` + fmt.Sprintf("%.3f", s.FillArea) + ` m²</span>`)
	sb.WriteString(`<span class="cut">🔴 挖方: ` + fmt.Sprintf("%.3f", s.CutArea) + ` m²</span>`)
	sb.WriteString(`<span class="clear">🟡 清表: ` + fmt.Sprintf("%.3f", s.ClearArea) + ` m²</span>`)
	for _, txt := range s.LayerAreaTexts {
		sb.WriteString(`<span style="color:#6a4a8a;">📐 ` + txt + `</span>`)
	}
	sb.WriteString(`</div>`)
	sb.WriteString(s.toSVG())
	sb.WriteString(`</div>`)
	return sb.String()
}

func (s SectionData) toSVG() string {
	// 收集所有点计算边界
	allPoints := make([][]float64, 0)
	allPoints = append(allPoints, s.Ground...)
	allPoints = append(allPoints, s.FinalDesign...)
	allPoints = append(allPoints, s.Cleared...)
	allPoints = append(allPoints, s.LeftSubgrade...)
	allPoints = append(allPoints, s.RightSubgrade...)
	for _, poly := range s.LayerPolygons {
		allPoints = append(allPoints, poly...)
	}
	if len(s.LeftSlopePoints) > 0 {
		allPoints = append(allPoints, s.LeftSlopePoints...)
	}
	if len(s.RightSlopePoints) > 0 {
		allPoints = append(allPoints, s.RightSlopePoints...)
	}

	if len(allPoints) == 0 {
		return `<svg viewBox="0 0 800 400"><text x="400" y="200" text-anchor="middle" fill="#999">无数据</text></svg>`
	}

	minX, maxX := math.Inf(1), math.Inf(-1)
	minY, maxY := math.Inf(1), math.Inf(-1)
	for _, p := range allPoints {
		if p[0] < minX {
			minX = p[0]
		}
		if p[0] > maxX {
			maxX = p[0]
		}
		if p[1] < minY {
			minY = p[1]
		}
		if p[1] > maxY {
			maxY = p[1]
		}
	}
	if maxX-minX < 0.1 {
		maxX = minX + 1
	}
	if maxY-minY < 0.1 {
		maxY = minY + 1
	}

	margin := 60.0
	viewWidth, viewHeight := 900.0, 500.0
	scaleX := (viewWidth - 2*margin) / (maxX - minX)
	scaleY := (viewHeight - 2*margin) / (maxY - minY)
	scale := math.Min(scaleX, scaleY)

	offsetX := (viewWidth - (maxX-minX)*scale) / 2
	offsetY := (viewHeight - (maxY-minY)*scale) / 2

	toX := func(x float64) float64 { return offsetX + (x-minX)*scale }
	toY := func(y float64) float64 { return viewHeight - offsetY - (y-minY)*scale }

	var sb strings.Builder
	sb.WriteString(`<svg viewBox="0 0 ` + fmt.Sprintf("%.0f %.0f", viewWidth, viewHeight) + `" xmlns="http://www.w3.org/2000/svg" style="font-family:'Microsoft YaHei',sans-serif;">`)

	sb.WriteString(`<rect width="` + fmt.Sprintf("%.0f", viewWidth) + `" height="` + fmt.Sprintf("%.0f", viewHeight) + `" fill="#fafcfe"/>`)

	// 网格
	gridStepX := math.Pow(10, math.Ceil(math.Log10((maxX-minX)/10)))
	gridStepY := math.Pow(10, math.Ceil(math.Log10((maxY-minY)/10)))
	if gridStepX < 1 {
		gridStepX = 1
	}
	if gridStepY < 1 {
		gridStepY = 1
	}

	for x := math.Ceil(minX/gridStepX) * gridStepX; x <= maxX; x += gridStepX {
		sx := toX(x)
		sb.WriteString(`<line x1="` + fmt.Sprintf("%.1f", sx) + `" y1="` + fmt.Sprintf("%.1f", margin) + `" x2="` + fmt.Sprintf("%.1f", sx) + `" y2="` + fmt.Sprintf("%.1f", viewHeight-margin) + `" stroke="#e8ecf1" stroke-width="0.5"/>`)
	}
	for y := math.Ceil(minY/gridStepY) * gridStepY; y <= maxY; y += gridStepY {
		sy := toY(y)
		sb.WriteString(`<line x1="` + fmt.Sprintf("%.1f", margin) + `" y1="` + fmt.Sprintf("%.1f", sy) + `" x2="` + fmt.Sprintf("%.1f", viewWidth-margin) + `" y2="` + fmt.Sprintf("%.1f", sy) + `" stroke="#e8ecf1" stroke-width="0.5"/>`)
	}

	// 结构层填充
	for _, poly := range s.LayerPolygons {
		if len(poly) < 3 {
			continue
		}
		sb.WriteString(`<polygon points="`)
		for _, p := range poly {
			sb.WriteString(fmt.Sprintf("%.2f,%.2f ", toX(p[0]), toY(p[1])))
		}
		sb.WriteString(`" fill="#f3e5f5" stroke="#7b1fa2" stroke-width="1" stroke-opacity="0.5" fill-opacity="0.25"/>`)
	}

	// 地面线
	sb.WriteString(polylineSVG(s.Ground, toX, toY, "#d32f2f", 2))
	// 清表线
	sb.WriteString(polylineSVG(s.Cleared, toX, toY, "#388e3c", 2, "dashed"))
	// 结构层
	if len(s.LeftSubgrade) > 0 {
		sb.WriteString(polylineSVG(s.LeftSubgrade, toX, toY, "#f57c00", 2))
	}
	if len(s.RightSubgrade) > 0 {
		sb.WriteString(polylineSVG(s.RightSubgrade, toX, toY, "#f57c00", 2))
	}
	// 设计线
	sb.WriteString(polylineSVG(s.FinalDesign, toX, toY, "#1976d2", 3))

	// === 标注（带 CSS class 以便前端控制字号） ===

	// 1. 中桩设计高程
	cx, cy := toX(0), toY(s.CenterY)
	sb.WriteString(`<circle cx="` + fmt.Sprintf("%.2f", cx) + `" cy="` + fmt.Sprintf("%.2f", cy) + `" r="3" fill="#1976d2"/>`)
	sb.WriteString(`<text class="text-elevation" x="` + fmt.Sprintf("%.2f", cx) + `" y="` + fmt.Sprintf("%.2f", cy+12) + `" text-anchor="left" font-size="11" fill="#1976d2" font-weight="bold" transform="rotate(-90,` + fmt.Sprintf("%.2f", cx) + `,` + fmt.Sprintf("%.2f", cy) + `)">` + fmt.Sprintf("%.3f", s.CenterY) + `</text>`)

	// 2. 左右边桩设计高程
	if s.LOuterX < -0.01 {
		lx, ly := toX(s.LOuterX), toY(s.LOuterY)
		sb.WriteString(`<text class="text-elevation" x="` + fmt.Sprintf("%.2f", lx) + `" y="` + fmt.Sprintf("%.2f", ly+12) + `" text-anchor="left" font-size="10" fill="#1976d2" transform="rotate(-90,` + fmt.Sprintf("%.2f", lx) + `,` + fmt.Sprintf("%.2f", ly) + `)">` + fmt.Sprintf("%.3f", s.LOuterY) + `</text>`)
	}
	if s.ROuterX > 0.01 {
		rx, ry := toX(s.ROuterX), toY(s.ROuterY)
		sb.WriteString(`<text class="text-elevation" x="` + fmt.Sprintf("%.2f", rx) + `" y="` + fmt.Sprintf("%.2f", ry+12) + `" text-anchor="left" font-size="10" fill="#1976d2" transform="rotate(-90,` + fmt.Sprintf("%.2f", rx) + `,` + fmt.Sprintf("%.2f", ry) + `)">` + fmt.Sprintf("%.3f", s.ROuterY) + `</text>`)
	}

	// 3. 横坡标注
	if s.LOuterX < -0.01 {
		midX := s.LOuterX / 2
		midY := s.CenterY + math.Abs(s.LOuterX/2)*(s.LeftCrossfall/100)
		mx, my := toX(midX), toY(midY)
		sb.WriteString(`<text class="text-crossfall" x="` + fmt.Sprintf("%.2f", mx) + `" y="` + fmt.Sprintf("%.2f", my-10) + `" text-anchor="middle" font-size="10" fill="#e65100" font-weight="bold">` + fmt.Sprintf("%.2f%%", s.LeftCrossfall) + `</text>`)
	}
	if s.ROuterX > 0.01 {
		midX := s.ROuterX / 2
		midY := s.CenterY + math.Abs(s.ROuterX/2)*(s.RightCrossfall/100)
		mx, my := toX(midX), toY(midY)
		sb.WriteString(`<text class="text-crossfall" x="` + fmt.Sprintf("%.2f", mx) + `" y="` + fmt.Sprintf("%.2f", my-10) + `" text-anchor="middle" font-size="10" fill="#e65100" font-weight="bold">` + fmt.Sprintf("%.2f%%", s.RightCrossfall) + `</text>`)
	}

	// 4. 边坡坡率标注
	if len(s.LeftSlopePoints) >= 1 {
		pts := make([][]float64, 0)
		pts = append(pts, []float64{s.LOuterX, s.LOuterY})
		for _, p := range s.LeftSlopePoints {
			if p[0] > s.LOuterX {
				pts = append(pts, p)
			}
		}
		pts = append(pts, []float64{s.LeftToeX, s.LeftToeY})

		for i := 0; i < len(pts)-1; i++ {
			p1, p2 := pts[i], pts[i+1]
			dx, dy := p2[0]-p1[0], p2[1]-p1[1]
			if math.Abs(dx) > 0.01 && math.Abs(dy) > 0.01 {
				ratio := math.Abs(dx) / math.Abs(dy)
				midX, midY := (p1[0]+p2[0])/2, (p1[1]+p2[1])/2
				mx, my := toX(midX), toY(midY)
				angle := -math.Atan2(dy, dx) * 180 / math.Pi
				if angle > 90 {
					angle -= 180
				} else if angle < -90 {
					angle += 180
				}
				sb.WriteString(`<text class="text-slope" x="` + fmt.Sprintf("%.2f", mx) + `" y="` + fmt.Sprintf("%.2f", my-4) + `" text-anchor="middle" font-size="9" fill="#1565c0" transform="rotate(` + fmt.Sprintf("%.1f", angle) + `,` + fmt.Sprintf("%.2f", mx) + `,` + fmt.Sprintf("%.2f", my) + `)">1:` + fmt.Sprintf("%.2f", ratio) + `</text>`)
			}
		}
	}

	if len(s.RightSlopePoints) >= 1 {
		pts := make([][]float64, 0)
		pts = append(pts, []float64{s.ROuterX, s.ROuterY})
		for _, p := range s.RightSlopePoints {
			if p[0] < s.ROuterX {
				pts = append(pts, p)
			}
		}
		pts = append(pts, []float64{s.RightToeX, s.RightToeY})

		for i := 0; i < len(pts)-1; i++ {
			p1, p2 := pts[i], pts[i+1]
			dx, dy := p2[0]-p1[0], p2[1]-p1[1]
			if math.Abs(dx) > 0.01 && math.Abs(dy) > 0.01 {
				ratio := math.Abs(dx) / math.Abs(dy)
				midX, midY := (p1[0]+p2[0])/2, (p1[1]+p2[1])/2
				mx, my := toX(midX), toY(midY)
				angle := -math.Atan2(dy, dx) * 180 / math.Pi
				if angle > 90 {
					angle -= 180
				} else if angle < -90 {
					angle += 180
				}
				sb.WriteString(`<text class="text-slope" x="` + fmt.Sprintf("%.2f", mx) + `" y="` + fmt.Sprintf("%.2f", my-4) + `" text-anchor="middle" font-size="9" fill="#1565c0" transform="rotate(` + fmt.Sprintf("%.1f", angle) + `,` + fmt.Sprintf("%.2f", mx) + `,` + fmt.Sprintf("%.2f", my) + `)">1:` + fmt.Sprintf("%.2f", ratio) + `</text>`)
			}
		}
	}

	// 桩号标注
	sb.WriteString(`<text class="text-station" x="` + fmt.Sprintf("%.2f", toX(0)) + `" y="` + fmt.Sprintf("%.2f", viewHeight-8) + `" text-anchor="middle" font-size="13" fill="#1a2a3a" font-weight="bold">` + s.Station + `</text>`)

	sb.WriteString(`</svg>`)
	return sb.String()
}

func polylineSVG(points [][]float64, toX, toY func(float64) float64, color string, strokeWidth float64, opts ...string) string {
	if len(points) < 2 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(`<polyline points="`)
	for _, p := range points {
		sb.WriteString(fmt.Sprintf("%.2f,%.2f ", toX(p[0]), toY(p[1])))
	}
	sb.WriteString(`" fill="none" stroke="` + color + `" stroke-width="` + fmt.Sprintf("%.1f", strokeWidth) + `"`)
	for _, opt := range opts {
		if opt == "dashed" {
			sb.WriteString(` stroke-dasharray="6,4"`)
		}
	}
	sb.WriteString(`/>`)
	return sb.String()
}