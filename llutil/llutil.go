package llutil

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
)

// ---------- 内部辅助函数（未导出） ----------

func FromXgetY(points [][]float64, targetX float64) float64 {
	const tolerance = 1e-6
	n := len(points)
	if n == 0 {
		return 0.0
	}
	low, high := 0, n-1
	index := -1
	for low <= high {
		mid := (low + high) >> 1
		midX := points[mid][0]
		if midX == targetX {
			index = mid
			break
		}
		if midX < targetX {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	if index >= 0 {
		return points[index][1]
	}
	rightIndex := low
	leftIndex := rightIndex - 1
	if rightIndex >= n {
		return points[n-1][1]
	}
	if leftIndex < 0 {
		return points[0][1]
	}
	xLeft, xRight := points[leftIndex][0], points[rightIndex][0]
	if math.Abs(xRight-xLeft) < tolerance {
		if points[leftIndex][1] > points[rightIndex][1] {
			return points[leftIndex][1]
		}
		return points[rightIndex][1]
	}
	yLeft, yRight := points[leftIndex][1], points[rightIndex][1]
	return yLeft + (yRight-yLeft)*(targetX-xLeft)/(xRight-xLeft)
}

func FromYgetX(points [][]float64, targetY float64) []float64 {
	const tolerance = 1e-12
	result := make([]float64, 0)
	n := len(points)
	if n == 0 {
		return result
	}
	for i := 0; i < n-1; i++ {
		x1, y1 := points[i][0], points[i][1]
		x2, y2 := points[i+1][0], points[i+1][1]
		y1Below := y1 <= targetY+tolerance
		y1Above := y1 >= targetY-tolerance
		y2Below := y2 <= targetY+tolerance
		y2Above := y2 >= targetY-tolerance
		cross := (y1Below && y2Above) || (y1Above && y2Below)
		if !cross {
			continue
		}
		if math.Abs(y2-y1) < tolerance {
			if len(result) == 0 || math.Abs(result[len(result)-1]-x1) > tolerance {
				result = append(result, x1)
			}
			if i == n-2 && (len(result) == 0 || math.Abs(result[len(result)-1]-x2) > tolerance) {
				result = append(result, x2)
			}
			continue
		}
		t := (targetY - y1) / (y2 - y1)
		xIntersect := x1 + t*(x2-x1)
		if len(result) == 0 || math.Abs(result[len(result)-1]-xIntersect) > tolerance {
			result = append(result, xIntersect)
		}
	}
	return result
}

func Hua_radiansToDMS(radians float64) float64 {
	degrees := radians * (180 / math.Pi)
	d := math.Floor(degrees)
	remaining := (degrees - d) * 60
	m := math.Floor(remaining)
	s := (remaining - m) * 60
	mm := ""
	if m < 10 {
		mm = "0" + strconv.Itoa(int(m))
	} else {
		mm = strconv.Itoa(int(m))
	}
	ss := ""
	if s < 10 {
		ss = "0" + strconv.FormatFloat(s, 'f', 1, 64)
	} else {
		ss = strconv.FormatFloat(s, 'f', 1, 64)
	}
	dfm := strconv.Itoa(int(d)) + mm + ss
	val, _ := strconv.ParseFloat(dfm, 64)
	return math.Round(val/10000.0*100000) / 100000
}

func Hua_radiansToDMS_度分秒(radians float64) string {
	degrees := radians * (180 / math.Pi)
	d := int(math.Floor(degrees))
	remaining := (degrees - float64(d)) * 60
	m := int(math.Floor(remaining))
	s := (remaining - float64(m)) * 60
	mm := ""
	if m < 10 {
		mm = "0" + strconv.Itoa(m)
	} else {
		mm = strconv.Itoa(m)
	}
	ss := ""
	if s < 10 {
		ss = "0" + strconv.FormatFloat(s, 'f', 1, 64)
	} else {
		ss = strconv.FormatFloat(s, 'f', 1, 64)
	}
	return strconv.Itoa(d) + "°" + mm + "′" + ss + "″"
}

func Hua_DmsToRadians(dms float64) float64 {
	degrees := int(dms)
	fractional := dms - float64(degrees)
	minutes := int(fractional * 100)
	seconds := (fractional*100 - float64(minutes)) * 100
	totalDegrees := float64(degrees) + float64(minutes)/60.0 + seconds/3600.0
	return totalDegrees * math.Pi / 180
}

func Hua_Num2K(meters float64) string {
	km := int(math.Floor(meters / 1000))
	m := meters - float64(km)*1000
	m = math.Round(m*100) / 100
	if m == float64(int(m)) {
		return "K" + strconv.Itoa(km) + "+" + strconv.Itoa(int(m)) + "000"[len(strconv.Itoa(int(m))):]
	}
	return "K" + strconv.Itoa(km) + "+" + strconv.FormatFloat(m, 'f', 2, 64)
}

func BuildClearPolygon(ground, cleared [][]float64, minX, maxX float64) [][]float64 {
	list := make([][]float64, 0)
	list = append(list, []float64{minX, FromXgetY(ground, minX)})
	for i := 0; i < len(ground); i++ {
		if ground[i][0] > minX && ground[i][0] < maxX {
			list = append(list, []float64{ground[i][0], ground[i][1]})
		}
	}
	list = append(list, []float64{maxX, FromXgetY(ground, maxX)})
	list = append(list, []float64{maxX, FromXgetY(cleared, maxX)})
	for i := len(cleared) - 1; i >= 0; i-- {
		if cleared[i][0] > minX && cleared[i][0] < maxX {
			list = append(list, []float64{cleared[i][0], cleared[i][1]})
		}
	}
	list = append(list, []float64{minX, FromXgetY(cleared, minX)})
	mat := make([][]float64, len(list))
	for i := 0; i < len(list); i++ {
		mat[i] = []float64{list[i][0], list[i][1]}
	}
	return mat
}

func AppendDxfLwPolyline(sb *strings.Builder, points [][]float64, offsetX, offsetY float64, layer string, colorIndex int) {
	if layer == "" {
		layer = "0"
	}
	if colorIndex == 0 {
		colorIndex = 7
	}
	vertexCount := len(points)
	if vertexCount < 2 {
		return
	}
	sb.WriteString("  0\n")
	sb.WriteString("POLYLINE\n")
	sb.WriteString("  8\n")
	sb.WriteString(layer + "\n")
	sb.WriteString(" 62\n")
	sb.WriteString(strconv.Itoa(colorIndex) + "\n")
	sb.WriteString(" 66\n")
	sb.WriteString("  1\n")
	for i := 0; i < vertexCount; i++ {
		absoluteX := points[i][0] + offsetX
		absoluteY := points[i][1] + offsetY
		sb.WriteString("  0\n")
		sb.WriteString("VERTEX\n")
		sb.WriteString("  8\n")
		sb.WriteString(layer + "\n")
		sb.WriteString(" 62\n")
		sb.WriteString(strconv.Itoa(colorIndex) + "\n")
		sb.WriteString(" 10\n")
		sb.WriteString(strconv.FormatFloat(absoluteX, 'f', 3, 64) + "\n")
		sb.WriteString(" 20\n")
		sb.WriteString(strconv.FormatFloat(absoluteY, 'f', 3, 64) + "\n")
	}
	sb.WriteString("  0\n")
	sb.WriteString("SEQEND\n")
	sb.WriteString("  8\n")
	sb.WriteString(layer + "\n")
}

func AppendDxfText(sb *strings.Builder, content string, x, y, height float64, layer string) {
	if layer == "" {
		layer = "TEXT_INFO"
	}
	if height == 0 {
		height = 2.5
	}
	sb.WriteString("  0\n")
	sb.WriteString("TEXT\n")
	sb.WriteString("  8\n")
	sb.WriteString(layer + "\n")
	sb.WriteString(" 10\n")
	sb.WriteString(strconv.FormatFloat(x, 'f', 3, 64) + "\n")
	sb.WriteString(" 20\n")
	sb.WriteString(strconv.FormatFloat(y, 'f', 3, 64) + "\n")
	sb.WriteString(" 40\n")
	sb.WriteString(strconv.FormatFloat(height, 'f', 2, 64) + "\n")
	sb.WriteString("  1\n")
	sb.WriteString(content + "\n")
}

func AppendDxfTextCenter(sb *strings.Builder, content string, x, y, height float64, layer string, rotation float64) {
	if layer == "" {
		layer = "TEXT_INFO"
	}
	if height == 0 {
		height = 2.5
	}
	sb.WriteString("  0\n")
	sb.WriteString("TEXT\n")
	sb.WriteString("  8\n")
	sb.WriteString(layer + "\n")
	sb.WriteString(" 10\n")
	sb.WriteString(strconv.FormatFloat(x, 'f', 3, 64) + "\n")
	sb.WriteString(" 20\n")
	sb.WriteString(strconv.FormatFloat(y, 'f', 3, 64) + "\n")
	sb.WriteString(" 40\n")
	sb.WriteString(strconv.FormatFloat(height, 'f', 3, 64) + "\n")
	sb.WriteString(" 50\n")
	sb.WriteString(strconv.FormatFloat(rotation, 'f', 3, 64) + "\n")
	sb.WriteString("  1\n")
	sb.WriteString(content + "\n")
	if math.Abs(rotation) < 0.001 {
		sb.WriteString(" 72\n")
		sb.WriteString("  1\n")
		sb.WriteString(" 11\n")
		sb.WriteString(strconv.FormatFloat(x, 'f', 3, 64) + "\n")
		sb.WriteString(" 21\n")
		sb.WriteString(strconv.FormatFloat(y, 'f', 3, 64) + "\n")
	}
}

func AppendDxfTextSlope(sb *strings.Builder, content string, x, y, height, rotationDegrees float64, layer string) {
	if layer == "" {
		layer = "SLOPE_TEXT_SIDE"
	}
	cleanAngle := rotationDegrees
	for cleanAngle > 90.0 {
		cleanAngle -= 180.0
	}
	for cleanAngle <= -90.0 {
		cleanAngle += 180.0
	}
	angleRad := cleanAngle * (math.Pi / 180.0)
	offsetDistance := height * 0.6
	offsetX := -math.Sin(angleRad) * offsetDistance
	offsetY := math.Cos(angleRad) * offsetDistance
	finalX := x + offsetX
	finalY := y + offsetY
	sb.WriteString("  0\n")
	sb.WriteString("TEXT\n")
	sb.WriteString("  8\n")
	sb.WriteString(layer + "\n")
	sb.WriteString(" 10\n")
	sb.WriteString(strconv.FormatFloat(finalX, 'f', 3, 64) + "\n")
	sb.WriteString(" 20\n")
	sb.WriteString(strconv.FormatFloat(finalY, 'f', 3, 64) + "\n")
	sb.WriteString(" 40\n")
	sb.WriteString(strconv.FormatFloat(height, 'f', 3, 64) + "\n")
	sb.WriteString(" 50\n")
	sb.WriteString(strconv.FormatFloat(cleanAngle, 'f', 3, 64) + "\n")
	sb.WriteString("  1\n")
	sb.WriteString(content + "\n")
	sb.WriteString(" 72\n")
	sb.WriteString("  1\n")
	sb.WriteString(" 73\n")
	sb.WriteString("  1\n")
	sb.WriteString(" 11\n")
	sb.WriteString(strconv.FormatFloat(finalX, 'f', 3, 64) + "\n")
	sb.WriteString(" 21\n")
	sb.WriteString(strconv.FormatFloat(finalY, 'f', 3, 64) + "\n")
}

// ---------- 内部计算函数（未导出） ----------

func Hua_getKBZ(x1, y1, x2, y2 float64, points [][]float64) [][]float64 {
	if points == nil {
		panic("points is nil")
	}
	rows := len(points)
	if rows == 0 {
		return make([][]float64, 0)
	}
	cols := len(points[0])
	if cols < 3 {
		panic("输入点集必须至少包含 X, Y, Z 三列")
	}
	dx := x2 - x1
	dy := y2 - y1
	len2 := dx*dx + dy*dy
	if len2 == 0 {
		panic("直线两点重合")
	}
	length := math.Sqrt(len2)
	a := y1 - y2
	b := x2 - x1
	c := x1*y2 - x2*y1
	result := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		x0 := points[i][0]
		y0 := points[i][1]
		z := points[i][2]
		dot := (x0-x1)*dx + (y0-y1)*dy
		signedDistToFoot := dot / length
		signedVerticalDist := (a*x0 + b*y0 + c) / length
		result[i] = []float64{signedDistToFoot, signedVerticalDist, z}
	}
	return result
}

func Hua_Fwj(x0, y0, x1, y1 float64) []float64 {
	x := x1 - x0
	y := y1 - y0
	cd := math.Sqrt(x*x + y*y)
	hd := math.Atan2(y, x)
	if hd < 0 {
		hd += 2 * math.Pi
	}
	return []float64{cd, hd}
}

func Hua_Zs(xyk, xyx, xyy, xyhd, xycd, xyqdr, xyzdr, xyzy, jsk, jsb, jd float64) []float64 {
	jdRad := Hua_DmsToRadians(jd)
	if math.Abs(xyqdr-xyzdr) < 0.01 && xyqdr > 0 {
		centerX := xyx + xyqdr*math.Cos(xyhd+xyzy*math.Pi/2)
		centerY := xyy + xyqdr*math.Sin(xyhd+xyzy*math.Pi/2)
		deltaArcLength := jsk - xyk
		deltaAngle := deltaArcLength / xyqdr * xyzy
		targetAzimuth := xyhd + deltaAngle
		if targetAzimuth < 0 {
			targetAzimuth += 2 * math.Pi
		}
		angle := xyhd + xyzy*math.Pi/2 + deltaAngle
		targetX := centerX - xyqdr*math.Cos(angle) + jsb*math.Cos(angle-xyzy*math.Pi/2+jdRad)
		targetY := centerY - xyqdr*math.Sin(angle) + jsb*math.Sin(angle-xyzy*math.Pi/2+jdRad)
		return []float64{targetX, targetY, targetAzimuth}
	}
	if xyqdr < 0.01 && xyzdr < 0.01 && xyzy < 0.01 {
		targetX := xyx + (jsk-xyk)*math.Cos(xyhd) + jsb*math.Cos(xyhd+jdRad)
		targetY := xyy + (jsk-xyk)*math.Sin(xyhd) + jsb*math.Sin(xyhd+jdRad)
		return []float64{targetX, targetY, xyhd}
	}
	if xyqdr < 0.001 {
		xyqdr = 99999999
	}
	if xyzdr < 0.001 {
		xyzdr = 99999999
	}
	f0 := xyhd
	q := xyzy
	c := 1 / xyqdr
	d := (xyqdr - xyzdr) / (2 * xycd * xyqdr * xyzdr)
	rr := []float64{0, 0.1739274226, 0.3260725774, 0.3260725774, 0.1739274226}
	vv := []float64{0, 0.0694318442, 0.3300094782, 1 - 0.3300094782, 1 - 0.0694318442}
	w := jsk - xyk
	xs := 0.0
	ys := 0.0
	for i := 1; i < 5; i++ {
		ff := f0 + q*vv[i]*w*(c+vv[i]*w*d)
		xs += rr[i] * math.Cos(ff)
		ys += rr[i] * math.Sin(ff)
	}
	fhz3 := f0 + q*w*(c+w*d)
	if fhz3 < 0 {
		fhz3 += 2 * math.Pi
	}
	if fhz3 >= 2*math.Pi {
		fhz3 -= 2 * math.Pi
	}
	fhz1 := xyx + w*xs + jsb*math.Cos(fhz3+jdRad)
	fhz2 := xyy + w*ys + jsb*math.Sin(fhz3+jdRad)
	return []float64{fhz1, fhz2, fhz3}
}

func Hua_Fs(pqx [][]float64, fsx, fsy float64) []float64 {
	jljd := Hua_Fwj(pqx[0][1], pqx[0][2], fsx, fsy)
	k := pqx[0][0]
	hudu := Hua_DmsToRadians(pqx[0][3])
	cz := jljd[0] * math.Cos(jljd[1]-hudu)
	pj := jljd[0] * math.Sin(jljd[1]-hudu)
	hang := len(pqx) - 1
	qdlc := pqx[0][0]
	zdlc := pqx[hang][0] + pqx[hang][4]
	jisuancishu := 0
	for math.Abs(cz) > 0.01 {
		k += cz
		jisuancishu++
		if k < qdlc {
			return []float64{-1, -1}
		}
		if k > zdlc {
			return []float64{-2, -2}
		}
		if jisuancishu > 15 {
			return []float64{-3, -3}
		}
		xy := Hua_Dantiaoxianludange(pqx, k, 0, 0)
		jljd = Hua_Fwj(xy[0], xy[1], fsx, fsy)
		cz = jljd[0] * math.Cos(jljd[1]-xy[2])
		pj = jljd[0] * math.Sin(jljd[1]-xy[2])
	}
	return []float64{math.Round(k*1000) / 1000, math.Round(pj*1000) / 1000}
}

func Hua_Gauss_proj(L, B, lonCenter float64) []float64 {
	pi := math.Pi
	e := 0.00669438002290
	e1 := 0.00673949677548
	b := 6356752.3141
	a := 6378137.0
	B = B * pi / 180
	L = L * pi / 180
	var L_num, L_center float64
	if lonCenter >= 359 {
		L_num = math.Floor(L*180/pi/3.0 + 0.5)
		L_center = 3 * L_num
	} else {
		L_center = lonCenter
	}
	l := (L/pi*180 - L_center) * 3600
	p0 := 206264.8062470963551564
	M0 := a * (1 - e)
	M2 := 3.0 / 2.0 * e * M0
	M4 := 5.0 / 4.0 * e * M2
	M6 := 7.0 / 6.0 * e * M4
	M8 := 9.0 / 8.0 * e * M6
	a0 := M0 + M2/2.0 + 3.0/8.0*M4 + 5.0/16.0*M6 + 35.0/128.0*M8
	a2 := M2/2.0 + M4/2 + 15.0/32.0*M6 + 7.0/16.0*M8
	a4 := M4/8.0 + 3.0/16.0*M6 + 7.0/32.0*M8
	a6 := M6/32.0 + M8/16.0
	a8 := M8 / 128.0
	Xz := a0*B - a2/2.0*math.Sin(2*B) + a4/4.0*math.Sin(4*B) - a6/6.0*math.Sin(6*B) + a8/8.0*math.Sin(8*B)
	c := a * a / b
	V := math.Sqrt(1 + e1*math.Cos(B)*math.Cos(B))
	N := c / V
	t := math.Tan(B)
	n := e1 * math.Cos(B) * math.Cos(B)
	m1 := N * math.Cos(B)
	m2 := N / 2.0 * math.Sin(B) * math.Cos(B)
	m3 := N / 6.0 * math.Pow(math.Cos(B), 3) * (1 - t*t + n)
	m4 := N / 24.0 * math.Sin(B) * math.Pow(math.Cos(B), 3) * (5 - t*t + 9*n)
	m5 := N / 120.0 * math.Pow(math.Cos(B), 5) * (5 - 18*t*t + math.Pow(t, 4) + 14*n - 58*n*t*t)
	m6 := N / 720.0 * math.Sin(B) * math.Pow(math.Cos(B), 5) * (61 - 58*t*t + math.Pow(t, 4))
	x := Xz + m2*l*l/math.Pow(p0, 2) + m4*math.Pow(l, 4)/math.Pow(p0, 4) + m6*math.Pow(l, 6)/math.Pow(p0, 6)
	y0 := m1*l/p0 + m3*math.Pow(l, 3)/math.Pow(p0, 3) + m5*math.Pow(l, 5)/math.Pow(p0, 5)
	y := y0 + 500000
	return []float64{x, y, L_center}
}

func Hua_Gauss_unproj(x, y, l0 float64) []float64 {
	pi := math.Pi
	e := 0.00669438002290
	e1 := 0.00673949677548
	b := 6356752.3141
	a := 6378137.0
	y1 := y - 500000
	M0 := a * (1 - e)
	M2 := 3.0 / 2.0 * e * M0
	M4 := 5.0 / 4.0 * e * M2
	M6 := 7.0 / 6.0 * e * M4
	M8 := 9.0 / 8.0 * e * M6
	a0 := M0 + M2/2.0 + 3.0/8.0*M4 + 5.0/16.0*M6 + 35.0/128.0*M8
	a2 := M2/2.0 + M4/2 + 15.0/32.0*M6 + 7.0/16.0*M8
	a4 := M4/8.0 + 3.0/16.0*M6 + 7.0/32.0*M8
	a6 := M6/32.0 + M8/16.0
	Bf := x / a0
	B0 := Bf
	for math.Abs(Bf-B0) > 0.0000001 || B0 == Bf {
		B0 = Bf
		FBf := -a2/2.0*math.Sin(2*B0) + a4/4.0*math.Sin(4*B0) - a6/6.0*math.Sin(6*B0)
		Bf = (x - FBf) / a0
	}
	t := math.Tan(Bf)
	c := a * a / b
	V := math.Sqrt(1 + e1*math.Cos(Bf)*math.Cos(Bf))
	N := c / V
	M := c / math.Pow(V, 3)
	n := e1 * math.Cos(Bf) * math.Cos(Bf)
	n1 := 1 / (N * math.Cos(Bf))
	n2 := -t / (2.0 * M * N)
	n3 := -(1 + 2*t*t + n) / (6.0 * math.Pow(N, 3) * math.Cos(Bf))
	n4 := t * (5 + 3*t*t + n - 9*n*t*t) / (24.0 * M * math.Pow(N, 3))
	n5 := (5 + 28*t*t + 24*math.Pow(t, 4) + 6*n + 8*n*t*t) / (120.0 * math.Pow(N, 5) * math.Cos(Bf))
	n6 := -t * (61 + 90*t*t + 45*math.Pow(t, 4)) / (720.0 * M * math.Pow(N, 5))
	B := (Bf + n2*y1*y1 + n4*math.Pow(y1, 4) + n6*math.Pow(y1, 6)) / pi * 180
	l := n1*y1 + n3*math.Pow(y1, 3) + n5*math.Pow(y1, 5)
	L := l0 + l/pi*180
	return []float64{L, B}
}

func Hua_Utm_proj(longitude, latitude float64) []float64 {
	EQUATORIAL_RADIUS := 6378137.0
	FLATTENING := 1 / 298.257223563
	ECC_SQUARED := 2*FLATTENING - math.Pow(FLATTENING, 2)
	ECC_PRIME_SQUARED := ECC_SQUARED / (1 - ECC_SQUARED)
	SCALE_FACTOR := 0.9996
	FALSE_EASTING := 500000.0
	FALSE_NORTHING_S := 10000000.0
	zoneNumber := int(math.Floor((longitude+180)/6)) + 1
	centralMeridian := float64(zoneNumber-1)*6 - 180 + 3
	latRad := latitude * math.Pi / 180.0
	lonRad := longitude * math.Pi / 180.0
	lonCenterRad := centralMeridian * math.Pi / 180.0
	N := EQUATORIAL_RADIUS / math.Sqrt(1-ECC_SQUARED*math.Pow(math.Sin(latRad), 2))
	T := math.Pow(math.Tan(latRad), 2)
	C := ECC_PRIME_SQUARED * math.Pow(math.Cos(latRad), 2)
	A := (lonRad - lonCenterRad) * math.Cos(latRad)
	M := EQUATORIAL_RADIUS * ((1-ECC_SQUARED/4-3*math.Pow(ECC_SQUARED, 2)/64-5*math.Pow(ECC_SQUARED, 3)/256)*latRad -
		(3*ECC_SQUARED/8+3*math.Pow(ECC_SQUARED, 2)/32+45*math.Pow(ECC_SQUARED, 3)/1024)*math.Sin(2*latRad) +
		(15*math.Pow(ECC_SQUARED, 2)/256+45*math.Pow(ECC_SQUARED, 3)/1024)*math.Sin(4*latRad) -
		(35*math.Pow(ECC_SQUARED, 3)/3072)*math.Sin(6*latRad))
	easting := SCALE_FACTOR*N*(A+(1-T+C)*math.Pow(A, 3)/6+(5-18*T+math.Pow(T, 2)+72*C-58*ECC_PRIME_SQUARED)*math.Pow(A, 5)/120) + FALSE_EASTING
	northing := SCALE_FACTOR * (M + N*math.Tan(latRad)*(math.Pow(A, 2)/2+(5-T+9*C+4*math.Pow(C, 2))*math.Pow(A, 4)/24+(61-58*T+math.Pow(T, 2)+600*C-330*ECC_PRIME_SQUARED)*math.Pow(A, 6)/720))
	if latitude < 0 {
		northing += FALSE_NORTHING_S
	}
	return []float64{northing, easting, float64(zoneNumber)}
}

func Hua_Utm_unproj(northing, easting float64, isNorthern bool, zoneNumber int) []float64 {
	EQUATORIAL_RADIUS := 6378137.0
	FLATTENING := 1 / 298.257223563
	ECC_SQUARED := 2*FLATTENING - math.Pow(FLATTENING, 2)
	ECC_PRIME_SQUARED := ECC_SQUARED / (1 - ECC_SQUARED)
	SCALE_FACTOR := 0.9996
	FALSE_EASTING := 500000.0
	FALSE_NORTHING_S := 10000000.0
	x := easting - FALSE_EASTING
	y := northing
	if !isNorthern {
		y -= FALSE_NORTHING_S
	}
	centralMeridian := float64(zoneNumber-1)*6 - 180 + 3
	lonCenterRad := centralMeridian * math.Pi / 180.0
	M := y / SCALE_FACTOR
	mu := M / (EQUATORIAL_RADIUS * (1 - ECC_SQUARED/4 - 3*math.Pow(ECC_SQUARED, 2)/64.0 - 5*math.Pow(ECC_SQUARED, 3)/256.0))
	e1 := (1 - math.Sqrt(1-ECC_SQUARED)) / (1 + math.Sqrt(1-ECC_SQUARED))
	phi1Rad := mu + (3*e1/2-27*math.Pow(e1, 3)/32)*math.Sin(2*mu) + (21*math.Pow(e1, 2)/16-55*math.Pow(e1, 4)/32)*math.Sin(4*mu) + (151*math.Pow(e1, 3)/96)*math.Sin(6*mu)
	N1 := EQUATORIAL_RADIUS / math.Sqrt(1-ECC_SQUARED*math.Pow(math.Sin(phi1Rad), 2))
	T1 := math.Pow(math.Tan(phi1Rad), 2)
	C1 := ECC_PRIME_SQUARED * math.Pow(math.Cos(phi1Rad), 2)
	R1 := EQUATORIAL_RADIUS * (1 - ECC_SQUARED) / math.Pow(1-ECC_SQUARED*math.Pow(math.Sin(phi1Rad), 2), 1.5)
	D := x / (N1 * SCALE_FACTOR)
	latRad := phi1Rad - (N1*math.Tan(phi1Rad)/R1)*(math.Pow(D, 2)/2-(5+3*T1+10*C1-4*math.Pow(C1, 2)-9*ECC_PRIME_SQUARED)*math.Pow(D, 4)/24+(61+90*T1+298*C1+45*math.Pow(T1, 2)-252*ECC_PRIME_SQUARED-3*math.Pow(C1, 2))*math.Pow(D, 6)/720)
	lonRad := lonCenterRad + (D-(1+2*T1+C1)*math.Pow(D, 3)/6+(5-2*C1+28*T1-3*math.Pow(C1, 2)+8*ECC_PRIME_SQUARED+24*math.Pow(T1, 2))*math.Pow(D, 5)/120)/math.Cos(phi1Rad)
	return []float64{lonRad * 180 / math.Pi, latRad * 180 / math.Pi}
}

func utm_zone(longitude float64) int {
	return int(math.Floor((longitude+180)/6)) + 1
}

func Hua_Cs4(source, target []float64) []float64 {
	if source == nil || target == nil || len(source) != len(target) {
		panic("坐标数组长度必须相等")
	}
	if len(source) < 4 || len(source)%2 != 0 {
		panic("至少需要2个点且坐标为偶数")
	}
	pointCount := len(source) / 2
	sumX1, sumY1, sumX2, sumY2 := 0.0, 0.0, 0.0, 0.0
	for i := 0; i < pointCount; i++ {
		sumX1 += source[2*i]
		sumY1 += source[2*i+1]
		sumX2 += target[2*i]
		sumY2 += target[2*i+1]
	}
	meanX1 := sumX1 / float64(pointCount)
	meanY1 := sumY1 / float64(pointCount)
	meanX2 := sumX2 / float64(pointCount)
	meanY2 := sumY2 / float64(pointCount)
	centeredSource := make([]float64, len(source))
	centeredTarget := make([]float64, len(target))
	for i := 0; i < pointCount; i++ {
		centeredSource[2*i] = source[2*i] - meanX1
		centeredSource[2*i+1] = source[2*i+1] - meanY1
		centeredTarget[2*i] = target[2*i] - meanX2
		centeredTarget[2*i+1] = target[2*i+1] - meanY2
	}
	H11, H12, H21, H22 := 0.0, 0.0, 0.0, 0.0
	B1, B2 := 0.0, 0.0
	for i := 0; i < pointCount; i++ {
		x1 := centeredSource[2*i]
		y1 := centeredSource[2*i+1]
		x2 := centeredTarget[2*i]
		y2 := centeredTarget[2*i+1]
		H11 += x1*x1 + y1*y1
		H22 += x1*x1 + y1*y1
		B1 += x1*x2 + y1*y2
		B2 += x1*y2 - y1*x2
	}
	det := H11*H22 - H12*H21
	if math.Abs(det) < 1e-15 {
		panic("矩阵奇异，无法求解参数")
	}
	a := (H22*B1 - H12*B2) / det
	b := (-H21*B1 + H11*B2) / det
	scale := math.Sqrt(a*a + b*b)
	rotation := math.Atan2(b, a)
	deltaX := meanX2 - (a*meanX1 - b*meanY1)
	deltaY := meanY2 - (b*meanX1 + a*meanY1)
	return []float64{deltaX, deltaY, rotation, scale}
}

func Hua_FourParameterTransform(x, y, deltaX, deltaY, rotation, scale float64) []float64 {
	convertedX := scale*(x*math.Cos(rotation)-y*math.Sin(rotation)) + deltaX
	convertedY := scale*(x*math.Sin(rotation)+y*math.Cos(rotation)) + deltaY
	return []float64{convertedX, convertedY}
}

func mapGps2XyBatch(lonlatPoints [][]float64, proj string, controlLonLat, controlXy [][]float64, center float64) [][]float64 {
	if lonlatPoints == nil || controlLonLat == nil || controlXy == nil {
		panic("参数不能为空")
	}
	pointCount := len(controlLonLat)
	if pointCount < 2 || len(controlXy) < pointCount {
		panic("控制点数量不足（至少需要2个）或控制点坐标数组不匹配")
	}
	sourceXy := make([]float64, pointCount*2)
	targetXy := make([]float64, pointCount*2)
	isGauss := strings.Contains(proj, "gao")
	isUtm := strings.Contains(proj, "utm")
	if !isGauss && !isUtm {
		panic("投影类型必须包含 'gao' 或 'utm'")
	}
	var cs4 []float64
	if isGauss {
		for i := 0; i < pointCount; i++ {
			ls := Hua_Gauss_proj(controlLonLat[i][0], controlLonLat[i][1], center)
			sourceXy[2*i] = ls[0]
			sourceXy[2*i+1] = ls[1]
			targetXy[2*i] = controlXy[i][0]
			targetXy[2*i+1] = controlXy[i][1]
		}
		cs4 = Hua_Cs4(sourceXy, targetXy)
	} else {
		for i := 0; i < pointCount; i++ {
			ls := Hua_Utm_proj(controlLonLat[i][0], controlLonLat[i][1])
			sourceXy[2*i] = ls[1] // easting -> X
			sourceXy[2*i+1] = ls[0] // northing -> Y
			targetXy[2*i] = controlXy[i][0]
			targetXy[2*i+1] = controlXy[i][1]
		}
		cs4 = Hua_Cs4(sourceXy, targetXy)
	}
	rows := len(lonlatPoints)
	result := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		lon, lat := lonlatPoints[i][0], lonlatPoints[i][1]
		var projCoord []float64
		if isGauss {
			ls := Hua_Gauss_proj(lon, lat, center)
			projCoord = []float64{ls[0], ls[1]}
		} else {
			ls := Hua_Utm_proj(lon, lat)
			projCoord = []float64{ls[1], ls[0]}
		}
		transformed := Hua_FourParameterTransform(projCoord[0], projCoord[1], cs4[0], cs4[1], cs4[2], cs4[3])
		result[i] = []float64{transformed[0], transformed[1]}
	}
	return result
}

func mapXy2GpsBatch(xyPoints [][]float64, proj string, controlLonLat, controlXy [][]float64) [][]float64 {
	if xyPoints == nil || controlLonLat == nil || controlXy == nil {
		panic("参数不能为空")
	}
	pointCount := len(controlLonLat)
	if pointCount < 2 || len(controlXy) < pointCount {
		panic("控制点数量不足（至少需要2个）或控制点坐标数组不匹配")
	}
	center := 0.0
	for i := 0; i < pointCount; i++ {
		center += controlLonLat[i][0]
	}
	center /= float64(pointCount)
	sourceXY := make([]float64, pointCount*2)
	targetXY := make([]float64, pointCount*2)
	isGauss := strings.Contains(strings.ToLower(proj), "gao")
	isUtm := strings.Contains(strings.ToLower(proj), "utm")
	if !isGauss && !isUtm {
		panic("投影类型必须包含 'gao' 或 'utm'")
	}
	var cs4 []float64
	if isGauss {
		for i := 0; i < pointCount; i++ {
			projected := Hua_Gauss_proj(controlLonLat[i][0], controlLonLat[i][1], center)
			sourceXY[2*i] = projected[0]
			sourceXY[2*i+1] = projected[1]
			targetXY[2*i] = controlXy[i][0]
			targetXY[2*i+1] = controlXy[i][1]
		}
		cs4 = Hua_Cs4(targetXY, sourceXY)
	} else {
		for i := 0; i < pointCount; i++ {
			projected := Hua_Utm_proj(controlLonLat[i][0], controlLonLat[i][1])
			sourceXY[2*i] = projected[1]
			sourceXY[2*i+1] = projected[0]
			targetXY[2*i] = controlXy[i][0]
			targetXY[2*i+1] = controlXy[i][1]
		}
		cs4 = Hua_Cs4(targetXY, sourceXY)
	}
	rows := len(xyPoints)
	result := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		x, y := xyPoints[i][0], xyPoints[i][1]
		lsxy := Hua_FourParameterTransform(x, y, cs4[0], cs4[1], cs4[2], cs4[3])
		var gps []float64
		if isGauss {
			gps = Hua_Gauss_unproj(lsxy[0], lsxy[1], center)
		} else {
			north := controlLonLat[0][1] >= 0
			zone := utm_zone(center)
			unproj := Hua_Utm_unproj(lsxy[1], lsxy[0], north, zone)
			gps = []float64{unproj[0], unproj[1]}
		}
		result[i] = []float64{gps[0], gps[1]}
	}
	return result
}

// ---------- 核心导出函数（供外部调用） ----------

// Hua_Dantiaoxianludange 坐标正算（里程k，偏距b，边桩角度z）
func Hua_Dantiaoxianludange(pqx [][]float64, k, b, z float64) []float64 {
	for i := 0; i < len(pqx); i++ {
		dtk := pqx[i][0]
		dtx := pqx[i][1]
		dty := pqx[i][2]
		dtfwj := pqx[i][3]
		dtcd := pqx[i][4]
		dtr1 := pqx[i][5]
		dtr2 := pqx[i][6]
		dtzy := pqx[i][7]
		if k >= dtk && k <= dtk+dtcd {
			hudu := Hua_DmsToRadians(dtfwj)
			jsxy1 := Hua_Zs(dtk, dtx, dty, hudu, dtcd, dtr1, dtr2, dtzy, k, b, z)
			return []float64{math.Round(jsxy1[0]*1000) / 1000, math.Round(jsxy1[1]*1000) / 1000, jsxy1[2]}
		}
	}
	return []float64{0, 0, 0}
}

// Hua_H 竖曲线高程计算
func Hua_H(sqx [][]float64, k float64) float64 {
	n := len(sqx)
	if n == 0 {
		panic("变坡点数据不能为空")
	}
	firstZ := sqx[0][0]
	lastZ := sqx[n-1][0]
	if k < firstZ || k > lastZ {
		return -1
	}
	if n == 1 {
		if math.Abs(k-firstZ) < 1e-9 {
			return sqx[0][1]
		}
		return -1
	}
	idx := -1
	for i := 0; i < n-1; i++ {
		if k >= sqx[i][0] && k <= sqx[i+1][0] {
			idx = i
			break
		}
	}
	if idx == -1 {
		return -1
	}
	z1, h1 := sqx[idx][0], sqx[idx][1]
	z2, h2 := sqx[idx+1][0], sqx[idx+1][1]
	slope := (h2 - h1) / (z2 - z1)
	result := h1 + slope*(k-z1)
	points := []int{idx, idx + 1}
	for _, p := range points {
		if p <= 0 || p >= n-1 {
			continue
		}
		r := sqx[p][2]
		if r <= 0 {
			continue
		}
		zv, hv := sqx[p][0], sqx[p][1]
		zPrev, hPrev := sqx[p-1][0], sqx[p-1][1]
		i1 := (hv - hPrev) / (zv - zPrev)
		zNext, hNext := sqx[p+1][0], sqx[p+1][1]
		i2 := (hNext - hv) / (zNext - zv)
		omega := i2 - i1
		T := r * math.Abs(omega) / 2.0
		start := zv - T
		end := zv + T
		if k >= start && k <= end {
			var tanElev float64
			if k <= zv {
				tanElev = hv + i1*(k-zv)
			} else {
				tanElev = hv + i2*(k-zv)
			}
			var x float64
			if k <= zv {
				x = k - start
			} else {
				x = end - k
			}
			y := x * x / (2.0 * r) * math.Copysign(1, omega)
			result = tanElev + y
		}
	}
	return math.Round(result*1000) / 1000
}

// Hua_JD2PQX 交点转平曲线（内部使用，已修正未使用变量）
func Hua_JD2PQX(data [][]float64) [][]float64 {
	listXY1 := make([]float64, 0)
	lr18 := data
	sk := lr18[0][2]
	JD2Mileage := 0.0
	hzx, hzy := 0.0, 0.0
	JD1 := make([]float64, 2)
	JD2 := make([]float64, 2)
	JD3 := make([]float64, 2)
	for i := 1; i < len(lr18)-1; i++ {
		JD1[0] = lr18[i-1][0]
		JD1[1] = lr18[i-1][1]
		if i > 1 {
			JD1[0] = hzx
			JD1[1] = hzy
		}
		JD2[0] = lr18[i][0]
		JD2[1] = lr18[i][1]
		JD3[0] = lr18[i+1][0]
		JD3[1] = lr18[i+1][1]
		R := lr18[i][4]
		Ls1 := lr18[i][2]
		Ls2 := lr18[i][3]
		xyzy := 1.0
		dx12 := JD2[0] - JD1[0]
		dy12 := JD2[1] - JD1[1]
		azimuth12 := math.Atan2(dy12, dx12)
		if azimuth12 < 0 {
			azimuth12 += 2 * math.Pi
		}
		dx23 := JD3[0] - JD2[0]
		dy23 := JD3[1] - JD2[1]
		azimuth23 := math.Atan2(dy23, dx23)
		if azimuth23 < 0 {
			azimuth23 += 2 * math.Pi
		}
		alpha := azimuth23 - azimuth12
		if alpha < 0 {
			alpha = -alpha
		}
		if alpha > math.Pi {
			alpha = math.Pi*2 - alpha
		}
		area := 0.5 * (JD1[0]*(JD2[1]-JD3[1]) + JD2[0]*(JD3[1]-JD1[1]) + JD3[0]*(JD1[1]-JD2[1]))
		if area < 0 {
			xyzy = -1
		}
		m1 := Ls1/2 - math.Pow(Ls1, 3)/(240*math.Pow(R, 2)) - math.Pow(Ls1, 5)/(34560*math.Pow(R, 4))
		m2 := Ls2/2 - math.Pow(Ls2, 3)/(240*math.Pow(R, 2)) - math.Pow(Ls2, 5)/(34560*math.Pow(R, 4))
		p1 := math.Pow(Ls1, 2)/(24*R) - math.Pow(Ls1, 4)/(2688*R*R*R)
		p2 := math.Pow(Ls2, 2)/(24*R) - math.Pow(Ls2, 4)/(2688*R*R*R)
		T1 := m1 + (R+p2-(R+p1)*math.Cos(alpha))/math.Sin(alpha)
		T2 := m2 + (R+p1-(R+p2)*math.Cos(alpha))/math.Sin(alpha)
		beta01 := Ls1 / (2 * R)
		beta02 := Ls2 / (2 * R)
		Ly := R * (alpha - beta01 - beta02)
		L := Ls1 + Ls2 + Ly
		var ZH, HY, YH float64 // 已删除 QZ, HZ
		dist1 := math.Sqrt(math.Pow(JD2[0]-JD1[0], 2) + math.Pow(JD2[1]-JD1[1], 2))
		JD2Mileage = sk + dist1
		ZH = JD2Mileage - T1
		HY = ZH + Ls1
		YH = ZH + L - Ls2
		if dist1 > T1 && dist1-T1 > 0.01 {
			xycd := dist1 - T1
			listXY1 = append(listXY1, sk, JD1[0], JD1[1], azimuth12, xycd, 0, 0, 0)
		}
		zhx := JD2[0] - T1*math.Cos(azimuth12)
		zhy := JD2[1] - T1*math.Sin(azimuth12)
		zhk := sk + dist1 - T1
		if Ls1 != 0 {
			listXY1 = append(listXY1, zhk, zhx, zhy, azimuth12, Ls1, 0, R, xyzy)
		}
		xy := []float64{0, 0, 0}
		if Ly != 0 && Ls1 != 0 {
			xy = Hua_Zs(ZH, zhx, zhy, azimuth12, Ls1, 0, R, xyzy, ZH+Ls1, 0, 90)
			listXY1 = append(listXY1, HY, xy[0], xy[1], xy[2], Ly, R, R, xyzy)
		} else {
			listXY1 = append(listXY1, HY, zhx, zhy, azimuth12, Ly, R, R, xyzy)
		}
		if Ls2 != 0 {
			xy = Hua_Zs(HY, xy[0], xy[1], xy[2], Ly, R, R, xyzy, HY+Ly, 0, 90)
			listXY1 = append(listXY1, YH, xy[0], xy[1], xy[2], Ls2, R, 0, xyzy)
		}
		hzx = JD2[0] + T2*math.Cos(azimuth23)
		hzy = JD2[1] + T2*math.Sin(azimuth23)
		sk = zhk + L
		if i == len(lr18)-2 {
			dist2 := math.Sqrt(math.Pow(JD3[0]-JD2[0], 2) + math.Pow(JD3[1]-JD2[1], 2))
			if dist2-T2 > 0.01 {
				listXY1 = append(listXY1, sk, hzx, hzy, azimuth23, dist2-T2, 0, 0, 0)
			}
		}
	}
	rows := len(listXY1) / 8
	pqx := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		pqx[i] = make([]float64, 8)
		for j := 0; j < 8; j++ {
			val := listXY1[i*8+j]
			if j != 3 {
				pqx[i][j] = math.Round(val*1000) / 1000
			} else {
				pqx[i][3] = Hua_radiansToDMS(val)
			}
		}
	}
	return pqx
}

// Hua_PolygonArea 多边形面积
func Hua_PolygonArea(poly [][]float64) float64 {
	if len(poly) < 3 {
		return 0
	}
	area := 0.0
	n := len(poly)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += poly[i][0]*poly[j][1]
		area -= poly[i][1]*poly[j][0]
	}
	return math.Abs(area) / 2.0
}

func Hua_FootPoint(px, py, x1, y1, x2, y2 float64) []float64 {
	dx := x2 - x1
	dy := y2 - y1
	len2 := dx*dx + dy*dy
	if len2 < 1e-12 {
		return []float64{x1, y1, 0}
	}
	t := ((px-x1)*dx + (py-y1)*dy) / len2
	footX := x1 + t*dx
	footY := y1 + t*dy
	flag := 0
	if t < 0 {
		flag = 2
	} else if t > 1 {
		flag = 1
	}
	return []float64{footX, footY, float64(flag)}
}

func Hua_SegmentIntersection(x1, y1, x2, y2, x3, y3, x4, y4 float64) []float64 {
	dx1 := x2 - x1
	dy1 := y2 - y1
	dx2 := x4 - x3
	dy2 := y4 - y3
	denom := dx1*dy2 - dy1*dx2
	if math.Abs(denom) < 1e-12 {
		return []float64{0, 0, -1, -1}
	}
	t := ((x3-x1)*dy2 - (y3-y1)*dx2) / denom
	u := ((x3-x1)*dy1 - (y3-y1)*dx1) / denom
	ix := x1 + t*dx1
	iy := y1 + t*dy1
	flag1 := 0
	if t < 0 {
		flag1 = 2
	} else if t > 1 {
		flag1 = 1
	}
	flag2 := 0
	if u < 0 {
		flag2 = 2
	} else if u > 1 {
		flag2 = 1
	}
	return []float64{ix, iy, float64(flag1), float64(flag2)}
}

func Hua_slope(mileage float64, points [][]float64, interpType int) float64 {
	if points == nil {
		panic("点表不能为空")
	}
	n := len(points)
	if n == 0 {
		panic("点表不能为空")
	}
	if len(points[0]) < 2 {
		panic("点表必须包含至少两列：里程和横坡")
	}
	for i := 1; i < n; i++ {
		if points[i][0] <= points[i-1][0] {
			panic("里程必须严格递增")
		}
	}
	if mileage <= points[0][0] {
		return points[0][1]
	}
	if mileage >= points[n-1][0] {
		return points[n-1][1]
	}
	if interpType == 0 {
		for i := 0; i < n-1; i++ {
			k1, k2 := points[i][0], points[i+1][0]
			if mileage >= k1 && mileage <= k2 {
				s1, s2 := points[i][1], points[i+1][1]
				if math.Abs(k2-k1) < 1e-9 {
					return s1
				}
				t := (mileage - k1) / (k2 - k1)
				return s1 + t*(s2-s1)
			}
		}
	} else if interpType == 1 {
		i := 0
		for ; i < n-1; i++ {
			if mileage <= points[i+1][0] {
				break
			}
		}
		left, mid, right := 0, 0, 0
		if i+2 < n {
			left, mid, right = i, i+1, i+2
		} else if i-2 >= 0 {
			left, mid, right = i-2, i-1, i
		} else {
			return Hua_slope(mileage, points, 0)
		}
		x0, y0 := points[left][0], points[left][1]
		x1, y1 := points[mid][0], points[mid][1]
		x2, y2 := points[right][0], points[right][1]
		x := mileage
		result := y0*(x-x1)*(x-x2)/((x0-x1)*(x0-x2)) +
			y1*(x-x0)*(x-x2)/((x1-x0)*(x1-x2)) +
			y2*(x-x0)*(x-x1)/((x2-x0)*(x2-x1))
		return result
	} else {
		panic("不支持的插值类型")
	}
	return points[0][1]
}

// Hua_CutAndFillArea 填挖面积（已删除未使用的D变量）
func Hua_CutAndFillArea(dmx, sjx [][]float64, extendDist float64) []float64 {
	if extendDist > 0 {
		n := len(dmx)
		if n >= 2 {
			x1, y1 := dmx[0][0], dmx[0][1]
			x2, y2 := dmx[1][0], dmx[1][1]
			dx, dy := x1-x2, y1-y2
			length := math.Sqrt(dx*dx + dy*dy)
			if length > 1e-12 {
				dx /= length
				dy /= length
				dmx[0][0] = x1 + dx*extendDist
				dmx[0][1] = y1 + dy*extendDist
			}
			x1, y1 = dmx[n-2][0], dmx[n-2][1]
			x2, y2 = dmx[n-1][0], dmx[n-1][1]
			dx, dy = x2-x1, y2-y1
			length = math.Sqrt(dx*dx + dy*dy)
			if length > 1e-12 {
				dx /= length
				dy /= length
				dmx[n-1][0] = x2 + dx*extendDist
				dmx[n-1][1] = y2 + dy*extendDist
			}
		}
	}
	xys := make([][]float64, 0)
	fill, cut := 0.0, 0.0
	lenA := len(sjx)
	lenB := len(dmx)
	for i := 0; i < lenA-1; i++ {
		x1, y1 := sjx[i][0], sjx[i][1]
		x2, y2 := sjx[i+1][0], sjx[i+1][1]
		for j := 0; j < lenB-1; j++ {
			x3, y3 := dmx[j][0], dmx[j][1]
			x4, y4 := dmx[j+1][0], dmx[j+1][1]
			denom := (x1-x2)*(y3-y4) - (y1-y2)*(x3-x4)
			if denom != 0 {
				t := ((x1-x3)*(y3-y4) - (y1-y3)*(x3-x4)) / denom
				u := ((x1-x3)*(y1-y2) - (y1-y3)*(x1-x2)) / denom
				if t >= 0 && t <= 1 && u >= 0 && u <= 1 {
					px := x1 + t*(x2-x1)
					py := y1 + t*(y2-y1)
					xys = append(xys, []float64{px, py, float64(i), float64(j)})
				}
			}
		}
	}
	minX, maxX, minY, maxY := 0.0, 0.0, 0.0, 0.0
	leftXysIdx, rightXysIdx := 0, 0
	if len(xys) > 0 {
		minX, maxX = xys[0][0], xys[0][0]
		minY, maxY = xys[0][1], xys[0][1]
		for i := 1; i < len(xys); i++ {
			x, y := xys[i][0], xys[i][1]
			if x < minX {
				minX = x
				leftXysIdx = i
			} else if x > maxX {
				maxX = x
				rightXysIdx = i
			}
			if y < minY {
				minY = y
			} else if y > maxY {
				maxY = y
			}
		}
	}
	for idx := 0; idx < len(xys)-1; idx++ {
		xy0, xy1 := xys[idx], xys[idx+1]
		x0, y0 := xy0[0], xy0[1]
		x1p, y1p := xy1[0], xy1[1]
		i0, i1 := int(xy0[2]), int(xy1[2])
		j0, j1 := int(xy0[3]), int(xy1[3])
		pts := make([][]float64, 0)
		pts = append(pts, []float64{x0, y0})
		for k := i0 + 1; k <= i1; k++ {
			pts = append(pts, []float64{sjx[k][0], sjx[k][1]})
		}
		pts = append(pts, []float64{x1p, y1p})
		for k := j1; k > j0; k-- {
			pts = append(pts, []float64{dmx[k][0], dmx[k][1]})
		}
		pts = append(pts, []float64{x0, y0})
		signedArea := 0.0
		for k := 0; k < len(pts)-1; k++ {
			p1, p2 := pts[k], pts[k+1]
			signedArea += p1[0]*p2[1] - p1[1]*p2[0]
		}
		area := signedArea / 2.0
		if signedArea > 0 {
			cut += area
		} else {
			fill += area
		}
	}
	finalSjxList := make([][]float64, 0)
	if len(xys) >= 2 {
		leftIntersection := xys[leftXysIdx]
		rightIntersection := xys[rightXysIdx]
		leftSjxSegIdx := int(leftIntersection[2])
		rightSjxSegIdx := int(rightIntersection[2])
		finalSjxList = append(finalSjxList, []float64{leftIntersection[0], leftIntersection[1]})
		for k := leftSjxSegIdx + 1; k <= rightSjxSegIdx; k++ {
			finalSjxList = append(finalSjxList, []float64{sjx[k][0], sjx[k][1]})
		}
		finalSjxList = append(finalSjxList, []float64{rightIntersection[0], rightIntersection[1]})
	} else {
		for k := 0; k < len(sjx); k++ {
			finalSjxList = append(finalSjxList, []float64{sjx[k][0], sjx[k][1]})
		}
	}
	finalResults := make([]float64, 0)
	finalResults = append(finalResults, math.Round(fill*10000)/10000, math.Round(cut*10000)/10000, minX, maxX, minY, maxY, float64(len(finalSjxList)))
	for _, pt := range finalSjxList {
		finalResults = append(finalResults, pt[0], pt[1])
	}
	return finalResults
}

func Hua_OffsetPolyline(points [][]float64, offset float64) [][]float64 {
	if points == nil {
		return nil
	}
	ptCount := len(points)
	if ptCount < 2 {
		return points
	}
	segmentCount := ptCount - 1
	segLines := make([][]float64, segmentCount)
	validSeg := make([]bool, segmentCount)
	for i := 0; i < segmentCount; i++ {
		dx := points[i+1][0] - points[i][0]
		dy := points[i+1][1] - points[i][1]
		length := math.Sqrt(dx*dx + dy*dy)
		if length < 1e-8 {
			continue
		}
		validSeg[i] = true
		nx := -dy / length
		ny := dx / length
		segLines[i] = []float64{
			points[i][0] + nx*offset,
			points[i][1] + ny*offset,
			points[i+1][0] + nx*offset,
			points[i+1][1] + ny*offset,
		}
	}
	rawVertices := make([][]float64, 0)
	firstIdx := -1
	for i := 0; i < segmentCount; i++ {
		if validSeg[i] {
			firstIdx = i
			break
		}
	}
	if firstIdx == -1 {
		return make([][]float64, 0)
	}
	rawVertices = append(rawVertices, []float64{segLines[firstIdx][0], segLines[firstIdx][1]})
	prev := firstIdx
	for i := firstIdx + 1; i < segmentCount; i++ {
		if !validSeg[i] {
			continue
		}
		x1, y1 := segLines[prev][0], segLines[prev][1]
		x2, y2 := segLines[prev][2], segLines[prev][3]
		x3, y3 := segLines[i][0], segLines[i][1]
		x4, y4 := segLines[i][2], segLines[i][3]
		denom := (x1-x2)*(y3-y4) - (y1-y2)*(x3-x4)
		if math.Abs(denom) > 1e-8 {
			t1 := x1*y2 - y1*x2
			t2 := x3*y4 - y3*x4
			ix := (t1*(x3-x4) - (x1-x2)*t2) / denom
			iy := (t1*(y3-y4) - (y1-y2)*t2) / denom
			rawVertices = append(rawVertices, []float64{ix, iy})
		} else {
			rawVertices = append(rawVertices, []float64{(x2 + x3) / 2.0, (y2 + y3) / 2.0})
		}
		prev = i
	}
	rawVertices = append(rawVertices, []float64{segLines[prev][2], segLines[prev][3]})
	cleanVertices := make([][]float64, 0)
	if len(rawVertices) > 0 {
		cleanVertices = append(cleanVertices, rawVertices[0])
	}
	for i := 1; i < len(rawVertices); i++ {
		curr := rawVertices[i]
		for len(cleanVertices) > 1 {
			lastValid := cleanVertices[len(cleanVertices)-1]
			if curr[0] < lastValid[0]-1e-5 {
				cleanVertices = cleanVertices[:len(cleanVertices)-1]
			} else {
				break
			}
		}
		top := cleanVertices[len(cleanVertices)-1]
		dist := math.Sqrt(math.Pow(curr[0]-top[0], 2) + math.Pow(curr[1]-top[1], 2))
		if dist > 1e-4 {
			if curr[0] < top[0]+1e-4 && len(cleanVertices) > 1 {
				cleanVertices[len(cleanVertices)-1] = []float64{(top[0] + curr[0]) / 2.0, (top[1] + curr[1]) / 2.0}
			} else {
				cleanVertices = append(cleanVertices, curr)
			}
		}
	}
	output := make([][]float64, len(cleanVertices))
	for i, v := range cleanVertices {
		output[i] = []float64{v[0], v[1]}
	}
	return output
}


func Hua_area(x1, y1, x2, y2 float64, points, queryKeys [][]float64, tolerance float64) [][]interface{} {
	data := Hua_getKBZ(x1, y1, x2, y2, points)
	dataLen := len(data)
	used := make([]bool, dataLen)
	queryCount := len(queryKeys)
	result := make([][]interface{}, queryCount+2+dataLen)
	row := 0
	result[row] = []interface{}{"断面编号", "断面面积(m²)", "体积(m³)"}
	row++
	totalVolume := 0.0
	prevK, prevArea := 0.0, 0.0
	for idx := 0; idx < queryCount; idx++ {
		qk := queryKeys[idx][0]
		matched := make([][]float64, 0)
		for i := 0; i < dataLen; i++ {
			if math.Abs(data[i][0]-qk) <= tolerance {
				matched = append(matched, []float64{data[i][1], data[i][2]})
				used[i] = true
			}
		}
		// 排序
		for i := 0; i < len(matched); i++ {
			for j := i + 1; j < len(matched); j++ {
				if matched[i][0] > matched[j][0] {
					matched[i], matched[j] = matched[j], matched[i]
				}
			}
		}
		area := 0.0
		if len(matched) >= 3 {
			polygon := make([][]float64, len(matched))
			for i, m := range matched {
				polygon[i] = []float64{m[0], m[1]}
			}
			area = Hua_PolygonArea(polygon)
		}
		volume := 0.0
		if idx > 0 {
			deltaK := qk - prevK
			avgArea := (area + prevArea) / 2.0
			volume = avgArea * deltaK
			totalVolume += volume
		}
		result[row] = []interface{}{Hua_Num2K(qk), math.Round(area*1000) / 1000, math.Round(volume*1000) / 1000}
		prevK = qk
		prevArea = area
		row++
	}
	minK := queryKeys[0][0]
	maxK := queryKeys[queryCount-1][0]
	rangeStr := Hua_Num2K(minK) + "~" + Hua_Num2K(maxK)
	result[row] = []interface{}{rangeStr, "合计", math.Round(totalVolume*1000) / 1000}
	row++
	result[row] = []interface{}{"无效点", "桩号", "行号"}
	row++
	for i := 0; i < dataLen; i++ {
		if !used[i] {
			result[row] = []interface{}{-1, Hua_Num2K(data[i][0]), i}
			row++
		}
	}
	return result[:row]
}

// ReadDataFromFile 读取数据文件（导出）
func ReadDataFromFile(filePath string, columnCount int, encoding string) ([][]float64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	validRows := make([][]float64, 0)
	lineIndex := 0
	separators := []string{" ", ",", "，", "\t"}
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		var tokens []string
		for _, sep := range separators {
			if strings.Contains(trimmed, sep) {
				tokens = strings.Split(trimmed, sep)
				break
			}
		}
		if tokens == nil {
			tokens = strings.Fields(trimmed)
		}
		filtered := make([]string, 0)
		for _, t := range tokens {
			if t != "" {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) != columnCount {
			return nil, &FormatError{File: filePath, Line: lineIndex + 1, Expected: columnCount, Actual: len(filtered)}
		}
		doubleRow := make([]float64, columnCount)
		for i, token := range filtered {
			val, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return nil, &ParseError{File: filePath, Line: lineIndex + 1, Col: i + 1, Content: token}
			}
			doubleRow[i] = val
		}
		validRows = append(validRows, doubleRow)
		lineIndex++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return validRows, nil
}

// 错误类型
type FormatError struct {
	File     string
	Line     int
	Expected int
	Actual   int
}

func (e *FormatError) Error() string {
	return "文件 " + e.File + " 第 " + strconv.Itoa(e.Line) + " 行不是 " + strconv.Itoa(e.Expected) + " 列数据，实际列数: " + strconv.Itoa(e.Actual)
}

type ParseError struct {
	File    string
	Line    int
	Col     int
	Content string
}

func (e *ParseError) Error() string {
	return "解析失败 → 文件：" + e.File + " → 行号：" + strconv.Itoa(e.Line) + " → 列号：" + strconv.Itoa(e.Col) + " → 内容：" + e.Content
}