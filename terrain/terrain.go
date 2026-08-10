package terrain

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	// 引入成熟的 Go 语言开源 Delaunay 剖分库
	"github.com/fogleman/delaunay"
)

// ==========================================
// 1. 基础数据结构与接口定义
// ==========================================

// IPoint 定义点接口
type IPoint interface {
	GetX() float64
	GetY() float64
	GetZ() float64
}

// Point3D 实现了 IPoint 接口
type Point3D struct {
	X float64
	Y float64
	Z float64
}

func NewPoint3D(x, y, z float64) *Point3D {
	return &Point3D{X: x, Y: y, Z: z}
}

func (p *Point3D) GetX() float64 { return p.X }
func (p *Point3D) GetY() float64 { return p.Y }
func (p *Point3D) GetZ() float64 { return p.Z }

type PointD2 struct {
	X float64
	Y float64
}

type PointD3 struct {
	X float64
	Y float64
	Z float64
}

type sectionResult struct {
	dist float64
	z    float64
}

// ==========================================
// 2. TerrainMesh 地形三角网核心类 (上)
// ==========================================

type TerrainMesh struct {
	points    []*Point3D
	triangles []int
	grid      *SpatialGrid
}

// TriangleCount 获取三角形数量
func (tm *TerrainMesh) TriangleCount() int {
	return len(tm.triangles) / 3
}

// FromTextFile 从文本文件（CSV）加载地形点并构建地形三角网
func FromTextFile(filePath string, maxEdgeLength float64, gridCellSize float64) (*TerrainMesh, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("找不到指定的点云文件: %w", err)
	}
	defer file.Close()

	// 预设初始容量减少扩容开销
	pointsList := make([]PointD3, 0, 24096)
	scanner := bufio.NewScanner(file)

	// 忽略第一行标题 (N,E,Z)
	if scanner.Scan() {
		_ = scanner.Text()
	}

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 兼容英文逗号、中文逗号或空格、制表符分隔
		tokens := strings.FieldsFunc(line, func(r rune) bool {
			return r == ',' || r == '，' || r == ' ' || r == '\t'
		})

		if len(tokens) < 3 {
			continue
		}

		x, err1 := strconv.ParseFloat(tokens[0], 64)
		y, err2 := strconv.ParseFloat(tokens[1], 64)
		z, err3 := strconv.ParseFloat(tokens[2], 64)

		if err1 == nil && err2 == nil && err3 == nil {
			pointsList = append(pointsList, PointD3{X: x, Y: y, Z: z})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	count := len(pointsList)
	if count == 0 {
		return nil, errors.New("文件中未解析出任何有效的 [X, Y, Z] 地形点数据")
	}

	// 将数据转换为构造网格所需的二维切片
	terrainPoints := make([][]float64, count)
	for i := 0; i < count; i++ {
		terrainPoints[i] = []float64{pointsList[i].X, pointsList[i].Y, pointsList[i].Z}
	}

	return NewTerrainMesh(terrainPoints, maxEdgeLength, gridCellSize), nil
}

// NewTerrainMesh 构建地形三角网
func NewTerrainMesh(terrainPoints [][]float64, maxEdgeLength float64, gridCellSize float64) *TerrainMesh {
	rows := len(terrainPoints)
	points := make([]*Point3D, rows)
	delaunayPoints := make([]delaunay.Point, rows)

	for i := 0; i < rows; i++ {
		points[i] = NewPoint3D(terrainPoints[i][0], terrainPoints[i][1], terrainPoints[i][2])
		delaunayPoints[i] = delaunay.Point{X: terrainPoints[i][0], Y: terrainPoints[i][1]}
	}

	// 直接调用成熟的第三方库进行高性能三角剖分
	triangulation, err := delaunay.Triangulate(delaunayPoints)
	var triangles []int
	if err == nil {
		triangles = triangulation.Triangles
	} else {
		triangles = make([]int, 0)
	}

	// 大于0时自动过滤超长三角形边长
	if maxEdgeLength > 0 {
		filtered := make([]int, 0)
		maxSq := maxEdgeLength * maxEdgeLength

		for i := 0; i < len(triangles); i += 3 {
			i0 := triangles[i]
			i1 := triangles[i+1]
			i2 := triangles[i+2]

			dx01, dy01, dz01 := points[i0].X-points[i1].X, points[i0].Y-points[i1].Y, points[i0].Z-points[i1].Z
			dx12, dy12, dz12 := points[i1].X-points[i2].X, points[i1].Y-points[i2].Y, points[i1].Z-points[i2].Z
			dx20, dy20, dz20 := points[i2].X-points[i0].X, points[i2].Y-points[i0].Y, points[i2].Z-points[i0].Z

			if (dx01*dx01+dy01*dy01+dz01*dz01) <= maxSq &&
				(dx12*dx12+dy12*dy12+dz12*dz12) <= maxSq &&
				(dx20*dx20+dy20*dy20+dz20*dz20) <= maxSq {
				filtered = append(filtered, i0, i1, i2)
			}
		}
		triangles = filtered
	}

	// 构建空间网格（自动计算单元大小）
	grid := NewSpatialGrid(points, triangles, gridCellSize)

	return &TerrainMesh{
		points:    points,
		triangles: triangles,
		grid:      grid,
	}
}

// GetTrianglesAsArray 获取所有三角形顶点坐标（每行9列）
func (tm *TerrainMesh) GetTrianglesAsArray() [][]float64 {
	triCount := len(tm.triangles) / 3
	result := make([][]float64, triCount)

	for i := 0; i < triCount; i++ {
		i0 := tm.triangles[3*i]
		i1 := tm.triangles[3*i+1]
		i2 := tm.triangles[3*i+2]

		result[i] = []float64{
			tm.points[i0].X, tm.points[i0].Y, tm.points[i0].Z,
			tm.points[i1].X, tm.points[i1].Y, tm.points[i1].Z,
			tm.points[i2].X, tm.points[i2].Y, tm.points[i2].Z,
		}
	}
	return result
}
// ==========================================
// 2. TerrainMesh 地形三角网核心类 (下)
// ==========================================

// ExtractSections 用多条线段切割地形，返回各交点相对于线段中点的距离（左负右正）及高程
func (tm *TerrainMesh) ExtractSections(segments [][]float64) [][]float64 {
	var allResults []sectionResult
	segRows := len(segments)

	// 复用列表内存，防止频繁 GC 开销
	intersectPoints := make([]PointD2, 0, 4)
	localPoints := make([]PointD3, 0, 16)

	const tolerance = 0.1
	const toleranceSq = tolerance * tolerance

	for s := 0; s < segRows; s++ {
		start := PointD2{X: segments[s][0], Y: segments[s][1]}
		end := PointD2{X: segments[s][2], Y: segments[s][3]}
		vx := end.X - start.X
		vy := end.Y - start.Y
		lenSq := vx*vx + vy*vy
		if lenSq < 1e-12 {
			continue
		}
		length := math.Sqrt(lenSq)

		// 线段包围盒（扩展容差）
		minX := math.Min(start.X, end.X) - 1e-6
		maxX := math.Max(start.X, end.X) + 1e-6
		minY := math.Min(start.Y, end.Y) - 1e-6
		maxY := math.Max(start.Y, end.Y) + 1e-6

		// 从空间网格中获取候选三角形索引
		candidateIndices := tm.grid.Query(minX, maxX, minY, maxY)
		localPoints = localPoints[:0]

		for _, triIdx := range candidateIndices {
			i0 := tm.triangles[3*triIdx]
			i1 := tm.triangles[3*triIdx+1]
			i2 := tm.triangles[3*triIdx+2]

			a := PointD2{X: tm.points[i0].X, Y: tm.points[i0].Y}
			b := PointD2{X: tm.points[i1].X, Y: tm.points[i1].Y}
			c := PointD2{X: tm.points[i2].X, Y: tm.points[i2].Y}

			// 快速包围盒剔除
			if math.Max(a.X, math.Max(b.X, c.X)) < minX || math.Min(a.X, math.Min(b.X, c.X)) > maxX ||
				math.Max(a.Y, math.Max(b.Y, c.Y)) < minY || math.Min(a.Y, math.Min(b.Y, c.Y)) > maxY {
				continue
			}

			intersectPoints = intersectPoints[:0]
			intersectPoints = tm.addEdgeIntersections(a, b, start, end, intersectPoints)
			intersectPoints = tm.addEdgeIntersections(b, c, start, end, intersectPoints)
			intersectPoints = tm.addEdgeIntersections(c, a, start, end, intersectPoints)

			if len(intersectPoints) == 0 {
				continue
			}

			// 平面去重逻辑
			for idx := 0; idx < len(intersectPoints); idx++ {
				p := intersectPoints[idx]
				dup := false
				for j := 0; j < idx; j++ {
					q := intersectPoints[j]
					dx := p.X - q.X
					dy := p.Y - q.Y
					if dx*dx+dy*dy < toleranceSq {
						dup = true
						break
					}
				}
				if dup {
					continue
				}

				// 重心插值算高程 Z
				denom2 := (b.Y-c.Y)*(a.X-c.X) + (c.X-b.X)*(a.Y-c.Y)
				if math.Abs(denom2) < 1e-12 {
					continue
				}
				w1 := ((b.Y-c.Y)*(p.X-c.X) + (c.X-b.X)*(p.Y-c.Y)) / denom2
				w2 := ((c.Y-a.Y)*(p.X-c.X) + (a.X-c.X)*(p.Y-c.Y)) / denom2
				w3 := 1 - w1 - w2
				z := w1*tm.points[i0].Z + w2*tm.points[i1].Z + w3*tm.points[i2].Z

				newPt := PointD3{X: p.X, Y: p.Y, Z: z}

				// 全局断面内点2D/3D去重
				dupLocal := false
				for _, e := range localPoints {
					dxLocal := e.X - newPt.X
					dyLocal := e.Y - newPt.Y
					if dxLocal*dxLocal+dyLocal*dyLocal < toleranceSq {
						dupLocal = true
						break
					}
				}
				if !dupLocal {
					localPoints = append(localPoints, newPt)
				}
			}
		}

		// 计算投影距离
		segmentResults := make([]sectionResult, 0, len(localPoints))
		for _, p := range localPoints {
			t := ((p.X-start.X)*vx + (p.Y-start.Y)*vy) / lenSq
			dist := (t - 0.5) * length
			segmentResults = append(segmentResults, sectionResult{dist: dist, z: p.Z})
		}

		// 对投影切片结果进行按距离升序排列
		sort.Slice(segmentResults, func(i, j int) bool {
			return segmentResults[i].dist < segmentResults[j].dist
		})

		allResults = append(allResults, segmentResults...)
	}

	// 重新写回二维切片
	result := make([][]float64, len(allResults))
	for i := 0; i < len(allResults); i++ {
		result[i] = []float64{allResults[i].dist, allResults[i].z}
	}
	return result
}

// addEdgeIntersections 计算三角形边与切割线段的交点，处理共线情况
func (tm *TerrainMesh) addEdgeIntersections(p1, p2, start, end PointD2, output []PointD2) []PointD2 {
	cross := (p2.X-p1.X)*(end.Y-start.Y) - (p2.Y-p1.Y)*(end.X-start.X)

	if math.Abs(cross) < 1e-12 {
		lenSq := (end.X-start.X)*(end.X-start.X) + (end.Y-start.Y)*(end.Y-start.Y)
		if lenSq < 1e-12 {
			return output
		}

		t1 := ((p1.X-start.X)*(end.X-start.X) + (p1.Y-start.Y)*(end.Y-start.Y)) / lenSq
		t2 := ((p2.X-start.X)*(end.X-start.X) + (p2.Y-start.Y)*(end.Y-start.Y)) / lenSq

		if t1 >= -1e-9 && t1 <= 1+1e-9 {
			output = append(output, PointD2{X: p1.X, Y: p1.Y})
		}
		if t2 >= -1e-9 && t2 <= 1+1e-9 {
			output = append(output, PointD2{X: p2.X, Y: p2.Y})
		}
		return output
	}

	denom := (p2.Y-p1.Y)*(end.X-start.X) - (p2.X-p1.X)*(end.Y-start.Y)
	if math.Abs(denom) < 1e-12 {
		return output
	}

	u := ((p2.X-p1.X)*(start.Y-p1.Y) - (p2.Y-p1.Y)*(start.X-p1.X)) / denom
	v := ((end.X-start.X)*(start.Y-p1.Y) - (end.Y-start.Y)*(start.X-p1.X)) / denom

	if u >= -1e-9 && u <= 1+1e-9 && v >= -1e-9 && v <= 1+1e-9 {
		cu := u
		if u < 0 {
			cu = 0
		} else if u > 1 {
			cu = 1
		}
		output = append(output, PointD2{
			X: start.X + cu*(end.X-start.X),
			Y: start.Y + cu*(end.Y-start.Y),
		})
	}
	return output
}

// ==========================================
// 3. SpatialGrid 均匀网格空间加速类
// ==========================================

type SpatialGrid struct {
	cells    [][]int
	cellSize float64
	minX     float64
	minY     float64
	nx       int
	ny       int
}

func NewSpatialGrid(points []*Point3D, triangles []int, cellSize float64) *SpatialGrid {
	if len(points) == 0 {
		return &SpatialGrid{cells: [][]int{}}
	}

	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	minY, maxY := math.MaxFloat64, -math.MaxFloat64
	for _, p := range points {
		if p.X < minX { minX = p.X }
		if p.X > maxX { maxX = p.X }
		if p.Y < minY { minY = p.Y }
		if p.Y > maxY { maxY = p.Y }
	}

	const eps = 1e-9
	minX -= eps; maxX += eps; minY -= eps; maxY += eps

	// 自动计算网格大小：取三角形平均边长的 2 倍
	if cellSize <= 0 {
		totalEdge := 0.0
		edgeCount := 0
		for i := 0; i < len(triangles); i += 3 {
			i0, i1, i2 := triangles[i], triangles[i+1], triangles[i+2]
			totalEdge += gridDistance(points[i0], points[i1])
			totalEdge += gridDistance(points[i1], points[i2])
			totalEdge += gridDistance(points[i2], points[i0])
			edgeCount += 3
		}
		avgEdge := 1.0
		if edgeCount > 0 {
			avgEdge = totalEdge / float64(edgeCount)
		}
		cellSize = math.Max(avgEdge*2.0, 1e-6)
	}

	rangeX := maxX - minX
	rangeY := maxY - minY
	nx := int(math.Max(1, math.Ceil(rangeX/cellSize)))
	ny := int(math.Max(1, math.Ceil(rangeY/cellSize)))

	// 初始化内部均匀一维矩阵切片
	cells := make([][]int, nx*ny)
	for i := range cells {
		cells[i] = make([]int, 0)
	}

	sg := &SpatialGrid{
		cells:    cells,
		cellSize: cellSize,
		minX:     minX,
		minY:     minY,
		nx:       nx,
		ny:       ny,
	}

	triCount := len(triangles) / 3
	for t := 0; t < triCount; t++ {
		i0 := triangles[3*t]
		i1 := triangles[3*t+1]
		i2 := triangles[3*t+2]

		triMinX := math.Min(points[i0].X, math.Min(points[i1].X, points[i2].X))
		triMaxX := math.Max(points[i0].X, math.Max(points[i1].X, points[i2].X))
		triMinY := math.Min(points[i0].Y, math.Min(points[i1].Y, points[i2].Y))
		triMaxY := math.Max(points[i0].Y, math.Max(points[i1].Y, points[i2].Y))

		startX := int(math.Max(0, math.Floor((triMinX-sg.minX)/cellSize)))
		endX := int(math.Min(float64(nx-1), math.Floor((triMaxX-sg.minX)/cellSize)))
		startY := int(math.Max(0, math.Floor((triMinY-sg.minY)/cellSize)))
		endY := int(math.Min(float64(ny-1), math.Floor((triMaxY-sg.minY)/cellSize)))

		for ix := startX; ix <= endX; ix++ {
			for iy := startY; iy <= endY; iy++ {
				cellIdx := ix*ny + iy
				sg.cells[cellIdx] = append(sg.cells[cellIdx], t)
			}
		}
	}

	return sg
}

func gridDistance(a, b *Point3D) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// Query 查询与给定矩形相交的所有三角形索引（通过 map 模拟哈希集合去重）
func (sg *SpatialGrid) Query(minX, maxX, minY, maxY float64) []int {
	startX := int(math.Max(0, math.Floor((minX-sg.minX)/sg.cellSize)))
	endX := int(math.Min(float64(sg.nx-1), math.Floor((maxX-sg.minX)/sg.cellSize)))
	startY := int(math.Max(0, math.Floor((minY-sg.minY)/sg.cellSize)))
	endY := int(math.Min(float64(sg.ny-1), math.Floor((maxY-sg.minY)/sg.cellSize)))

	seen := make(map[int]struct{})
	for ix := startX; ix <= endX; ix++ {
		for iy := startY; iy <= endY; iy++ {
			cellIdx := ix*sg.ny + iy
			for _, triIdx := range sg.cells[cellIdx] {
				seen[triIdx] = struct{}{}
			}
		}
	}

	result := make([]int, 0, len(seen))
	for triIdx := range seen {
		result = append(result, triIdx)
	}
	return result
}
