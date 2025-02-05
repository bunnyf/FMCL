# FMCL (Financial Market Calendar Lite)

English | [简体中文](README.md)

## Introduction
FMCL is a lightweight terminal-based financial market calendar application that displays real-time financial event information with a Text User Interface (TUI).

## Features
- Real-time financial calendar events display
- Multiple display modes:
  - High importance events only
  - All events
  - High importance + rates information
  - High importance + important events
- Color-coded importance levels
- Live countdown timer for data refresh
- Keyboard shortcuts for easy operation

## Keyboard Shortcuts
- `q`: Quit application
- `r`: Force refresh data
- `p`: Pause/resume auto-refresh
- `m`: Switch display mode
- `h`: Show/hide help menu
- `ESC`: Close help menu

## Configuration
The application can be configured through `config.yaml`:
```yaml
refresh_interval: 15      # Data refresh interval in seconds
default_display_mode: 0   # Default display mode (0-3)
ui:
  time_width: 8          # Width of time column
  importance_width: 6    # Width of importance column
  value_width: 12        # Width of value columns
```

## Requirements
- Go 1.20 or higher
- Terminal with ANSI escape sequence support
- Terminal width of 120 characters recommended for optimal display

## Installation
```bash
git clone https://github.com/your-username/FMCL.git
cd FMCL
go mod download
```

## Running
```bash
go run cmd/main/main.go
```

## Display Modes
1. High Importance Only (Mode 0)
   - Shows only events marked as high importance
2. All Events (Mode 1)
   - Displays all financial calendar events
3. High Importance + Rates (Mode 2)
   - Shows high importance events and central bank rates
4. High Importance + Important Events (Mode 3)
   - Shows high importance events and other important market events

## UI Layout
- Header: Shows application name and startup time
- Main Display: Financial calendar events in tabular format
- Status Bar: Current mode, running status, and refresh countdown
- Help Menu: Accessible via 'h' key, closeable with ESC

## Color Coding
- High Importance: Red
- Medium Importance: Yellow
- Low Importance: White
- Current Values: Green
- Time Information: Cyan
- Headers: Green/Cyan
- Status Information: Green
