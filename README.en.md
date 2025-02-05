# Financial Market Command Line (FMCL)

English | [简体中文](README.md)

A command-line financial market monitoring tool written in Go, providing real-time display of important economic indicators and central bank rates. Developed with Go 1.20+.

[![GitHub](https://img.shields.io/github/license/bunnyf/FMCL)](https://github.com/bunnyf/FMCL)
[![Go Version](https://img.shields.io/github/go-mod/go-version/bunnyf/FMCL)](https://github.com/bunnyf/FMCL)
[![Latest Release](https://img.shields.io/github/v/release/bunnyf/FMCL)](https://github.com/bunnyf/FMCL/releases)

## Features

- Real-time monitoring of financial data and key economic indicators
- Multiple display modes (High importance only, Rate information, All data)
- Automatic data refresh with configurable intervals
- Keyboard shortcut support
- Status bar with system status and countdown timer
- Pause/Resume data refresh functionality

## Keyboard Shortcuts

- `q`: Exit program
- `r`: Force refresh data
- `p`: Pause/Resume data refresh
- `m`: Toggle display mode
- `h`: Show help information

## Configuration

### Config File

The configuration file is located at `config.yaml` and supports the following options:

```yaml
# Data refresh interval (seconds)
refresh_interval: 15

# Default display mode
# 0: High importance only
# 1: Rate information
# 2: All data
default_display_mode: 0

# Terminal display settings
display:
  # Show timestamp with data
  show_timestamp: true
  # Use colored output
  use_color: true
  # Terminal width in characters
  terminal_width: 120
```

### Configuration Details

1. `refresh_interval`
   - Time interval for automatic data refresh (seconds)
   - Recommended range: 15-60 seconds
   - Use shorter intervals (15-30s) for critical data monitoring
   - Use longer intervals (60s) for general use to reduce resource usage

2. `default_display_mode`
   - Default display mode when starting the program
   - Available options:
     - 0: High importance only (recommended for daily monitoring)
     - 1: Rate information
     - 2: All data (for detailed information review)

3. `display`
   - `show_timestamp`: Enable/disable timestamp display with data
   - `use_color`: Enable/disable colored output (recommended to keep enabled)
   - `terminal_width`: Terminal display width for alignment and formatting

## System Requirements

- Go 1.20 or higher
- Terminal with ANSI escape sequence support
- Recommended terminal width: 120 characters for optimal display
