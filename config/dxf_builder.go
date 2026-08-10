package config

import "strings"

// BuildDxf 构建完整的 DXF 文件内容
// entitiesBody: 已构建好的 ENTITIES 段内的实体内容（完整的 DXF 组码）
// 返回完整的 DXF 文件字符串
func BuildDxf(entitiesBody string) string {
	var sb strings.Builder
	sb.WriteString("  0\n")
	sb.WriteString("SECTION\n")
	sb.WriteString("  2\n")
	sb.WriteString("HEADER\n")
	sb.WriteString("  9\n")
	sb.WriteString("$DWGCODEPAGE\n")
	sb.WriteString("  3\n")
	sb.WriteString("UTF-8\n")
	sb.WriteString("  0\n")
	sb.WriteString("ENDSEC\n")
	sb.WriteString("  0\n")
	sb.WriteString("SECTION\n")
	sb.WriteString("  2\n")
	sb.WriteString("ENTITIES\n")
	sb.WriteString(entitiesBody)
	sb.WriteString("  0\n")
	sb.WriteString("ENDSEC\n")
	sb.WriteString("  0\n")
	sb.WriteString("EOF\n")
	return sb.String()
}

// BuildDxfFromBuilder 从 strings.Builder 构建 DXF（避免重复拷贝）
// entitiesBuilder: 已写入实体内容的 strings.Builder
// 返回完整的 DXF 文件字符串
func BuildDxfFromBuilder(entitiesBuilder *strings.Builder) string {
	return BuildDxf(entitiesBuilder.String())
}