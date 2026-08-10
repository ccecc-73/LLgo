package topslope

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// DuanMianShuJu 断面数据：桩号 + 板块序列 [宽度, 横坡%]
type DuanMianShuJu struct {
	ZhuangHao    float64
	BanKuaiJiHe [][]float64 // 每个元素为 [宽度, 横坡%]
}

// JueDuiBanKuaiDian 绝对板块点
type JueDuiBanKuaiDian struct {
	WidthX   float64 // 绝对横坐标（中桩为0，左负右正）
	GaoChengY float64 // 绝对高程
}

// LuJiCheDaoJieGuoBao 路基车道结果包
type LuJiCheDaoJieGuoBao struct {
	ZuoCeCheDaoJueDui  []JueDuiBanKuaiDian // 左侧所有拐点（已反转，从外侧向内指向中桩）
	YouCeCheDaoJueDui  []JueDuiBanKuaiDian // 右侧所有拐点（原序，从中桩向外指向外侧）
}

// NewLuJiCheDaoJieGuoBao 构造结果包
func NewLuJiCheDaoJieGuoBao() *LuJiCheDaoJieGuoBao {
	return &LuJiCheDaoJieGuoBao{
		ZuoCeCheDaoJueDui: make([]JueDuiBanKuaiDian, 0),
		YouCeCheDaoJueDui: make([]JueDuiBanKuaiDian, 0),
	}
}

// ParseWidthFile 解析宽度文件（跳过首行说明，忽略注释和空行）
func ParseWidthFile(filePath string) ([]DuanMianShuJu, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var result []DuanMianShuJu
	isFirstLine := true

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}
		if isFirstLine {
			isFirstLine = false
			continue // 跳过首行说明
		}

		// 按逗号、空格、制表符分割
		tokens := strings.FieldsFunc(line, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\t'
		})
		if len(tokens) < 3 {
			continue
		}

		zhuangHao, err := strconv.ParseFloat(tokens[0], 64)
		if err != nil {
			continue
		}
		dm := DuanMianShuJu{ZhuangHao: zhuangHao}
		for j := 1; j < len(tokens); j += 2 {
			if j+1 >= len(tokens) {
				break
			}
			w, err1 := strconv.ParseFloat(tokens[j], 64)
			s, err2 := strconv.ParseFloat(tokens[j+1], 64)
			if err1 != nil || err2 != nil {
				continue
			}
			dm.BanKuaiJiHe = append(dm.BanKuaiJiHe, []float64{w, s})
		}
		result = append(result, dm)
	}
	return result, scanner.Err()
}

// GetCrossSectionXY 根据目标桩号和中桩高程，计算左右侧车道绝对坐标
func GetCrossSectionXY(
	muBiaoZhuangHao float64,
	zhongZhuangGaoCheng float64,
	suoYouZuoData []DuanMianShuJu,
	suoYouYouData []DuanMianShuJu,
) *LuJiCheDaoJieGuoBao {
	jieGuoBao := NewLuJiCheDaoJieGuoBao()

	// 处理左侧
	if len(suoYouZuoData) > 0 {
		stBefore := suoYouZuoData[0].ZhuangHao
		stAfter := suoYouZuoData[len(suoYouZuoData)-1].ZhuangHao
		dmBefore := suoYouZuoData[0]
		dmAfter := suoYouZuoData[len(suoYouZuoData)-1]

		if muBiaoZhuangHao <= stBefore {
			dmAfter = dmBefore
			stAfter = stBefore
		} else if muBiaoZhuangHao >= stAfter {
			dmBefore = dmAfter
			stBefore = stAfter
		} else {
			for i := 1; i < len(suoYouZuoData); i++ {
				if suoYouZuoData[i].ZhuangHao >= muBiaoZhuangHao {
					stAfter = suoYouZuoData[i].ZhuangHao
					dmAfter = suoYouZuoData[i]
					stBefore = suoYouZuoData[i-1].ZhuangHao
					dmBefore = suoYouZuoData[i-1]
					break
				}
			}
		}

		ratio := 0.0
		if stBefore != stAfter {
			ratio = (muBiaoZhuangHao - stBefore) / (stAfter - stBefore)
		}
		curZuoX := 0.0
		curZuoY := zhongZhuangGaoCheng

		minLen := len(dmBefore.BanKuaiJiHe)
		if len(dmAfter.BanKuaiJiHe) < minLen {
			minLen = len(dmAfter.BanKuaiJiHe)
		}
		for k := 0; k < minLen; k++ {
			wBefore := dmBefore.BanKuaiJiHe[k][0]
			sBefore := dmBefore.BanKuaiJiHe[k][1]
			wAfter := dmAfter.BanKuaiJiHe[k][0]
			sAfter := dmAfter.BanKuaiJiHe[k][1]

			curWidth := wBefore + ratio*(wAfter-wBefore)
			curSlopeVal := (sBefore + ratio*(sAfter-sBefore)) / 100.0

			curZuoX -= curWidth
			curZuoY += curWidth * curSlopeVal
			jieGuoBao.ZuoCeCheDaoJueDui = append(jieGuoBao.ZuoCeCheDaoJueDui, JueDuiBanKuaiDian{WidthX: curZuoX, GaoChengY: curZuoY})
		}
		// 反转左侧列表（从外侧到中桩）
		for i, j := 0, len(jieGuoBao.ZuoCeCheDaoJueDui)-1; i < j; i, j = i+1, j-1 {
			jieGuoBao.ZuoCeCheDaoJueDui[i], jieGuoBao.ZuoCeCheDaoJueDui[j] = jieGuoBao.ZuoCeCheDaoJueDui[j], jieGuoBao.ZuoCeCheDaoJueDui[i]
		}
	}

	// 处理右侧
	if len(suoYouYouData) > 0 {
		stBefore := suoYouYouData[0].ZhuangHao
		stAfter := suoYouYouData[len(suoYouYouData)-1].ZhuangHao
		dmBefore := suoYouYouData[0]
		dmAfter := suoYouYouData[len(suoYouYouData)-1]

		if muBiaoZhuangHao <= stBefore {
			dmAfter = dmBefore
			stAfter = stBefore
		} else if muBiaoZhuangHao >= stAfter {
			dmBefore = dmAfter
			stBefore = stAfter
		} else {
			for i := 1; i < len(suoYouYouData); i++ {
				if suoYouYouData[i].ZhuangHao >= muBiaoZhuangHao {
					stAfter = suoYouYouData[i].ZhuangHao
					dmAfter = suoYouYouData[i]
					stBefore = suoYouYouData[i-1].ZhuangHao
					dmBefore = suoYouYouData[i-1]
					break
				}
			}
		}

		ratio := 0.0
		if stBefore != stAfter {
			ratio = (muBiaoZhuangHao - stBefore) / (stAfter - stBefore)
		}
		curYouX := 0.0
		curYouY := zhongZhuangGaoCheng

		minLen := len(dmBefore.BanKuaiJiHe)
		if len(dmAfter.BanKuaiJiHe) < minLen {
			minLen = len(dmAfter.BanKuaiJiHe)
		}
		for k := 0; k < minLen; k++ {
			wBefore := dmBefore.BanKuaiJiHe[k][0]
			sBefore := dmBefore.BanKuaiJiHe[k][1]
			wAfter := dmAfter.BanKuaiJiHe[k][0]
			sAfter := dmAfter.BanKuaiJiHe[k][1]

			curWidth := wBefore + ratio*(wAfter-wBefore)
			curSlopeVal := (sBefore + ratio*(sAfter-sBefore)) / 100.0

			curYouX += curWidth
			curYouY += curWidth * curSlopeVal
			jieGuoBao.YouCeCheDaoJueDui = append(jieGuoBao.YouCeCheDaoJueDui, JueDuiBanKuaiDian{WidthX: curYouX, GaoChengY: curYouY})
		}
	}

	return jieGuoBao
}