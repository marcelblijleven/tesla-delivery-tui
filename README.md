# Tesla Delivery TUI

A terminal user interface (TUI) for tracking Tesla vehicle orders and delivery status.

## Features

- **OAuth2 Authentication** - Secure login with Tesla account using PKCE flow
- **Order List View** - See all your Tesla orders with status badges
- **Order Details** - VIN, delivery window, vehicle location, tasks, and more
- **VIN Decoder** - Decode manufacturer, model, body type, powertrain, and production details
- **Vehicle Options** - Decode option codes to human-readable descriptions
- **Order Timeline** - Visual progress from order placed to delivery
- **Delivery Readiness** - Track customer and Tesla tasks blocking delivery
- **Change Detection** - Track order updates over time with history
- **Trade-In Details** - View trade-in vehicle information
- **Demo Mode** - Run with mock data for demos and screenshots
- **Secure Token Storage** - System keychain (macOS/Linux/Windows) with encrypted file fallback

## Installation

### Go Install

```bash
go install github.com/marcelblijleven/tesla-delivery-tui/cmd/tesla-delivery-tui@latest
```

### Binary Download

Download pre-built binaries from the [releases page](https://github.com/marcelblijleven/tesla-delivery-tui/releases).

## Usage

```bash
# Run the TUI
tesla-delivery-tui

# Run in demo mode (mock data)
tesla-delivery-tui --demo
```

### First Run

1. Launch the application
2. Press `Enter` to start authentication
3. Copy the URL and open it in your browser
4. Log in with your Tesla account. An error will be displayed, this is expected
5. Copy the callback URL from your browser and paste it back
6. Your orders will be displayed

### Key Bindings

| Key | Action |
|-----|--------|
| `↑`/`k` | Move up |
| `↓`/`j` | Move down |
| `←`/`h` | Move left |
| `→`/`l` | Move right |
| `Enter` | Select / View details |
| `Tab` | Next tab |
| `Shift+Tab` | Previous tab |
| `Esc` | Go back |
| `r` | Refresh data |
| `L` | Logout |
| `q` | Quit |

### Tabs in Detail View

- **Details** - Order timeline, VIN decoder, vehicle options, trade-in info
- **Tasks** - Delivery readiness checklist with customer and Tesla tasks
- **History** - Change history with timestamps
- **JSON** - Raw API response data

## Configuration

Configuration and tokens are stored in:

- **macOS/Linux**: `~/.config/tesla-delivery-tui/`
- **Windows**: `%APPDATA%\tesla-delivery-tui\`

Tokens are stored securely using:

1. System keychain (macOS Keychain, Linux Secret Service, Windows Credential Manager)
2. Fallback: AES-256-GCM encrypted file

## Building from Source

### Requirements

- Go 1.22 or later

### Build

```bash
git clone https://github.com/marcelblijleven/tesla-delivery-tui.git
cd tesla-delivery-tui
go build ./cmd/tesla-delivery-tui
```

### Run Tests

```bash
go test ./...
```

### Run with Coverage

```bash
go test ./... -cover
```

## Demo Mode

Run with mock data for demonstrations, screenshots, or VHS recordings:

```bash
tesla-delivery-tui --demo
```

This displays a sample Model Y order with realistic data without requiring authentication.

## Credits

Inspired by [tesla-delivery-status-web](https://github.com/GewoonJaap/tesla-delivery-status-web) by GewoonJaap.

Built with:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## License

MIT License - see [LICENSE](LICENSE) for details.

## Disclaimer

This project is not affiliated with, endorsed by, or connected to Tesla, Inc. in any way. Use at your own risk.
