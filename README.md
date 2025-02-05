# 财经市场命令行工具 (Financial Market Command Line - FMCL)

[English](README.en.md) | 简体中文

一个用 Go 语言编写的命令行财经数据监控工具，支持实时显示重要经济指标、央行利率等信息。基于 Go 1.20+ 开发。

[![GitHub](https://img.shields.io/github/license/bunnyf/FMCL)](https://github.com/bunnyf/FMCL)
[![Go Version](https://img.shields.io/github/go-mod/go-version/bunnyf/FMCL)](https://github.com/bunnyf/FMCL)
[![Latest Release](https://img.shields.io/github/v/release/bunnyf/FMCL)](https://github.com/bunnyf/FMCL/releases)

## 功能特点

- 实时监控财经数据和重要经济指标
- 支持多种显示模式（仅显示高重要性、显示利率信息、显示全部）
- 自动定时刷新数据
- 支持键盘快捷键操作
- 状态栏显示系统运行状态和倒计时
- 支持暂停/继续数据刷新

## 快捷键

- `q`: 退出程序
- `r`: 强制刷新数据
- `p`: 暂停/继续数据刷新
- `m`: 切换显示模式
- `h`: 显示帮助信息

## 配置说明

### 配置文件

配置文件位于 `config.yaml`，支持以下配置项：

```yaml
# 数据刷新间隔（秒）
refresh_interval: 15

# 默认显示模式
# 0: 仅显示高重要性
# 1: 显示利率信息
# 2: 显示全部
default_display_mode: 0

# 终端显示设置
display:
  # 是否显示时间戳
  show_timestamp: true
  # 是否使用彩色输出
  use_color: true
  # 终端宽度（字符数）
  terminal_width: 120
```

### 配置项说明

1. `refresh_interval`
   - 数据自动刷新的时间间隔（秒）
   - 建议值：15-60秒
   - 重要数据监控建议使用较短间隔（15-30秒）
   - 一般用途可使用较长间隔（60秒）以减少资源占用

2. `default_display_mode`
   - 程序启动时的默认显示模式
   - 可选值：
     - 0：仅显示高重要性数据（建议日常监控使用）
     - 1：显示利率信息
     - 2：显示所有数据（需要查看详细信息时使用）

3. `display`
   - `show_timestamp`: 是否在数据项旁显示时间戳
   - `use_color`: 是否使用彩色输出（建议保持开启）
   - `terminal_width`: 终端显示宽度，用于对齐和格式化

## 运行要求

- Go 1.20 或更高版本
- 支持 ANSI 转义序列的终端
- 终端窗口宽度建议设置为 120 字符以获得最佳显示效果
