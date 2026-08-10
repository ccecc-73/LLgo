# LLhdm-go 测绘自动化系统

> 基于 Go 语言开发的横断面计算与图纸生成工具，专为测绘与道路工程设计打造。

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20Android-lightgrey)]()

---

## 📖 项目简介

**LLhdm-go** 是一套完整的测绘横断面自动化处理系统，从地形点云、平曲线、竖曲线、边坡、结构层到横断面生成、方量统计、DXF 图纸输出，全流程自动化。支持 **Windows / Linux / Android** 跨平台运行，**单文件部署，无需任何依赖**。

---

## ✨ 功能特性

- **平曲线正算/反算**：支持任意里程坐标计算与坐标反算里程偏距
- **竖曲线高程计算**：基于变坡点数据插值设计高程
- **地形三角网构建**：基于 Delaunay 三角剖分，支持空间网格加速
- **横断面切割**：从地形网格提取任意断面的地面线
- **边坡处理**：填/挖方边坡自动判断与绝对坐标生成
- **结构层计算**：左右侧路面结构层逐层生成
- **清表线计算**：基于设计深度自动偏移
- **填挖面积计算**：断面法精确计算填方/挖方面积
- **方量统计**：按桩号累加生成 CSV 报表
- **DXF 图纸生成**：符合 A4 图幅的完整横断面图
- **HTML 可视化**：SVG 交互式横断面预览（支持桩号搜索）

---

## 📁 项目结构

---

## 🚀 快速开始

### 1. 编译运行

```bash
# 安装依赖
go mod tidy

# 运行程序
go run main.go

# 编译单文件
go build -ldflags="-s -w" -o LLhdm.exe   # Windows
go build -ldflags="-s -w" -o LLhdm        # Linux/macOS

