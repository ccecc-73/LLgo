package bianpo

import (
	"os"
	"strconv"
	"strings"
)

// BianPoDuanLuo 边坡段落数据模型
type BianPoDuanLuo struct {
	QiShiZhuangHao  float64     // 起始桩号
	JieShuZhuangHao float64     // 结束桩号
	ZuoTian         [][]float64 // 左填相对位移矩阵 [[dX, dY], ...]
	ZuoWa           [][]float64 // 左挖相对位移矩阵
	YouTian         [][]float64 // 右填相对位移矩阵
	YouWa           [][]float64 // 右挖相对位移矩阵
}

// BianPoHouXuanBao 边坡候选包（绝对坐标结果）
type BianPoHouXuanBao struct {
	ZuoTianJueDui [][]float64
	ZuoWaJueDui   [][]float64
	YouTianJueDui [][]float64
	YouWaJueDui   [][]float64
}

// NewBianPoHouXuanBao 构造一个空的候选包
func NewBianPoHouXuanBao() *BianPoHouXuanBao {
	return &BianPoHouXuanBao{
		ZuoTianJueDui: make([][]float64, 0),
		ZuoWaJueDui:   make([][]float64, 0),
		YouTianJueDui: make([][]float64, 0),
		YouWaJueDui:   make([][]float64, 0),
	}
}

// ParseBianPo 解析边坡数据文件（跳过注释行，忽略空行）
func ParseBianPo(filePath string) ([]*BianPoDuanLuo, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")

	// 过滤：去除空行和以分号开头的注释行
	cleanLines := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ";") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}
	lines = cleanLines

	result := make([]*BianPoDuanLuo, 0)

	// 每5行一组：桩号行 + 左填 + 左挖 + 右填 + 右挖
	for i := 0; i < len(lines); i += 5 {
		if i+4 >= len(lines) {
			break
		}

		// 解析桩号行
		zhuangHaoParts := strings.Split(lines[i], ",")
		if len(zhuangHaoParts) < 2 {
			continue
		}
		qs, err1 := strconv.ParseFloat(strings.TrimSpace(zhuangHaoParts[0]), 64)
		js, err2 := strconv.ParseFloat(strings.TrimSpace(zhuangHaoParts[1]), 64)
		if err1 != nil || err2 != nil {
			continue
		}

		duanLuo := &BianPoDuanLuo{
			QiShiZhuangHao:  qs,
			JieShuZhuangHao: js,
			ZuoTian:         parsePointLine(lines[i+1]),
			ZuoWa:           parsePointLine(lines[i+2]),
			YouTian:         parsePointLine(lines[i+3]),
			YouWa:           parsePointLine(lines[i+4]),
		}
		result = append(result, duanLuo)
	}

	return result, nil
}

// parsePointLine 解析一行点数据：跳过第一个元素（类别名），后续两两配对为 [dX, dY]
func parsePointLine(line string) [][]float64 {
	tokens := strings.Split(line, ",")
	pts := make([][]float64, 0)
	for j := 1; j < len(tokens); j += 2 {
		if j+1 >= len(tokens) {
			break
		}
		x, err1 := strconv.ParseFloat(strings.TrimSpace(tokens[j]), 64)
		y, err2 := strconv.ParseFloat(strings.TrimSpace(tokens[j+1]), 64)
		if err1 != nil || err2 != nil {
			continue
		}
		pts = append(pts, []float64{x, y})
	}
	return pts
}

// K2BianPo 根据桩号查找对应的边坡段落
func K2BianPo(suoYouBianPo []*BianPoDuanLuo, muBiaoZhuangHao float64) *BianPoDuanLuo {
	for _, duanLuo := range suoYouBianPo {
		if muBiaoZhuangHao >= duanLuo.QiShiZhuangHao && muBiaoZhuangHao < duanLuo.JieShuZhuangHao {
			return duanLuo
		}
	}
	return nil
}

// GetAbsolute 计算边坡的绝对坐标（从左/右原点出发累加相对位移）
func GetAbsolute(
	dangQianBianPo *BianPoDuanLuo,
	zuoYuanX, zuoYuanY float64,
	youYuanX, youYuanY float64,
) *BianPoHouXuanBao {
	houxuan := NewBianPoHouXuanBao()
	if dangQianBianPo == nil {
		return houxuan
	}

	// 左填
	ztX, ztY := zuoYuanX, zuoYuanY
	for _, step := range dangQianBianPo.ZuoTian {
		ztX += step[0]
		ztY += step[1]
		houxuan.ZuoTianJueDui = append(houxuan.ZuoTianJueDui, []float64{ztX, ztY})
	}
	reverse2DSlice(houxuan.ZuoTianJueDui)

	// 左挖
	zwX, zwY := zuoYuanX, zuoYuanY
	for _, step := range dangQianBianPo.ZuoWa {
		zwX += step[0]
		zwY += step[1]
		houxuan.ZuoWaJueDui = append(houxuan.ZuoWaJueDui, []float64{zwX, zwY})
	}
	reverse2DSlice(houxuan.ZuoWaJueDui)

	// 右填
	ytX, ytY := youYuanX, youYuanY
	for _, step := range dangQianBianPo.YouTian {
		ytX += step[0]
		ytY += step[1]
		houxuan.YouTianJueDui = append(houxuan.YouTianJueDui, []float64{ytX, ytY})
	}
	// 右填不反转（原 C# 逻辑）

	// 右挖
	ywX, ywY := youYuanX, youYuanY
	for _, step := range dangQianBianPo.YouWa {
		ywX += step[0]
		ywY += step[1]
		houxuan.YouWaJueDui = append(houxuan.YouWaJueDui, []float64{ywX, ywY})
	}
	// 右挖不反转

	return houxuan
}

// reverse2DSlice 反转二维切片顺序
func reverse2DSlice(pts [][]float64) {
	for i, j := 0, len(pts)-1; i < j; i, j = i+1, j-1 {
		pts[i], pts[j] = pts[j], pts[i]
	}
}