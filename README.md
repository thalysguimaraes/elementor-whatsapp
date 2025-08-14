# ewctl - Elementor WhatsApp Manager 🚀

![ewctl Banner](banner-image.jpg)

A powerful terminal UI for managing Elementor form webhooks with WhatsApp notifications. Built with Go and Bubble Tea for a beautiful, efficient experience. Deploy webhooks on Cloudflare Workers edge network and manage everything from your terminal.

## 🎯 Why ewctl?

Managing form submissions and WhatsApp notifications shouldn't require juggling multiple dashboards. **ewctl** solves this by providing:

- **🖥️ Single Control Center** - Manage forms, contacts, and webhooks from one beautiful TUI
- **⚡ Edge Performance** - Cloudflare Workers process webhooks at the edge for minimal latency
- **🔄 Real-time Sync** - Direct integration with Cloudflare D1 for instant updates
- **🎨 Beautiful Interface** - Not your typical CLI - a modern TUI with animations and mouse support
- **🚀 Zero Config Deploy** - Automated setup wizard handles everything

## ✨ Key Features

### 📝 Forms Management
- **Visual Form Builder** - Create forms with field mappings through intuitive UI
- **Multi-recipient Support** - Route submissions to multiple WhatsApp numbers
- **Field Validation** - Built-in validation for Elementor field types
- **Bulk Operations** - Import/export forms for backup and migration

### 📞 Contact Management
- **Centralized Database** - Reusable contacts across multiple forms
- **Smart Assignment** - Checkbox-based multi-select for form recipients  
- **Contact Details** - Store name, company, role, and notes
- **Import/Export** - CSV support for bulk contact management

### 🔄 Webhook Processing
- **Auto-detection** - Supports 3 different Elementor payload formats
- **Edge Processing** - Sub-50ms response times via Cloudflare Workers
- **Retry Logic** - Automatic retries with exponential backoff
- **Error Recovery** - Detailed logging and error categorization

### 🎨 Terminal UI Experience
- **Keyboard Navigation** - Vim-style bindings available
- **Mouse Support** - Click and scroll naturally
- **Real-time Updates** - Live dashboard with statistics
- **Theme Support** - Multiple color schemes (Charm, Dark, Light)
- **Responsive Layout** - Adapts to terminal size

### 🔒 Security & Reliability
- **Encrypted Storage** - Credentials stored securely
- **API Token Scoping** - Minimal permissions required
- **Audit Logging** - Track all operations
- **Health Monitoring** - Built-in connection status checks

## 🏗️ Architecture

```
┌──────────────────┐        ┌───────────────────┐       ┌──────────────────┐
│                  │        │                   │       │                  │
│  Elementor Form  │───────▶│ Cloudflare Worker │──────▶│      Z-API       │
│   (WordPress)    │ webhook│   (Edge Network)  │  API  │  (WhatsApp API)  │
│                  │        │                   │       │                  │
└──────────────────┘        └───────────────────┘       └──────────────────┘
                                      │                           │
                                      ▼                           ▼
                            ┌───────────────────┐       ┌──────────────────┐
                            │   Cloudflare D1   │       │  WhatsApp Users  │
                            │   (SQLite Edge)   │       │   (Recipients)   │
                            │                   │       │                  │
                            └───────────────────┘       └──────────────────┘
                                      ▲
                                      │
                            ┌───────────────────┐
                            │      ewctl        │
                            │   (Terminal UI)   │
                            │                   │
                            │ ┌───────────────┐ │
                            │ │   Dashboard   │ │
                            │ ├───────────────┤ │
                            │ │     Forms     │ │
                            │ ├───────────────┤ │
                            │ │   Contacts    │ │
                            │ ├───────────────┤ │
                            │ │  Webhook Test │ │
                            │ ├───────────────┤ │
                            │ │   Settings    │ │
                            │ └───────────────┘ │
                            └───────────────────┘
```

## 📋 Prerequisites

- **Go** 1.21+ (for building from source)
- **Cloudflare Account** (free tier works)
- **Z-API Account** for WhatsApp integration
- **Elementor Pro** with webhook support

## 🚀 Quick Start

### Automated Setup (Recommended)

```bash
# Install ewctl
curl -sSL https://raw.githubusercontent.com/thalysguimaraes/elementor-whatsapp/main/install.sh | bash

# Run setup wizard
ewctl setup

# The wizard will:
# ✅ Configure Cloudflare credentials
# ✅ Set up Z-API integration  
# ✅ Deploy the worker
# ✅ Initialize database
# ✅ Create your first form
```

### Build from Source

```bash
# Clone repository
git clone https://github.com/thalysguimaraes/elementor-whatsapp.git
cd elementor-whatsapp

# Build with make
make build

# Or with go directly
go build -o bin/ewctl cmd/ewctl/main.go

# Run the TUI
./bin/ewctl
```

## 🔧 Configuration

Configuration file location: `~/.config/ewctl/config.yaml`

```yaml
# Cloudflare Configuration
cloudflare:
  account_id: your-account-id      # Found in dashboard sidebar
  api_token: your-api-token        # Create with D1 edit permissions
  database_id: your-database-id    # From D1 dashboard
  worker_url: https://your-worker.workers.dev

# Z-API Configuration  
zapi:
  instance_id: your-instance-id
  instance_token: your-instance-token
  client_token: your-client-token

# UI Preferences
ui:
  theme: charm              # Options: charm, dark, light, default
  mouse: true              # Enable mouse support
  animations: true         # Enable UI animations
  vim_bindings: false      # Use vim-style navigation
  auto_refresh: 30s        # Dashboard refresh interval
  confirm_destructive: true # Confirm before delete operations

# Environment Profiles
profiles:
  dev:
    worker_url: http://localhost:8787
  production:
    worker_url: https://your-worker.workers.dev
```

## 📖 Usage

### Terminal UI Navigation

```bash
# Start the TUI
ewctl

# Or with specific config
ewctl --config /path/to/config.yaml
```

**Keyboard Shortcuts:**
- `1-5` - Quick navigation (Dashboard, Forms, Contacts, Test, Settings)
- `n/a` - New form/contact
- `e` - Edit selected item
- `d` - Delete (with confirmation)
- `Enter` - View details
- `r` - Refresh data
- `ESC` - Go back
- `q` - Quit (from dashboard)
- `?` - Help

### CLI Commands

```bash
# Forms management
ewctl forms list              # List all forms
ewctl forms create            # Interactive form creation
ewctl forms edit <id>         # Edit existing form
ewctl forms delete <id>       # Delete form

# Contacts management  
ewctl contacts list           # List all contacts
ewctl contacts add            # Add new contact
ewctl contacts import <file>  # Import from CSV

# Webhook testing
ewctl webhook test <form-id>  # Send test webhook

# Configuration
ewctl config show             # Display current config
ewctl config edit             # Open config in editor
```

## 🌐 Elementor Integration

### 1. Add Webhook Action

In your Elementor form:
1. Go to **Actions After Submit**
2. Add **Webhook** action
3. Set webhook URL: `https://your-worker.workers.dev/webhook/{form-id}`

### 2. Configure Field Mapping

ewctl automatically detects these Elementor field formats:

```javascript
// Format 1: Nested JSON
{
  "form": { "id": "abc123", "name": "Contact Form" },
  "fields": {
    "name": { "value": "John Doe" },
    "email": { "value": "john@example.com" }
  }
}

// Format 2: URL-encoded
fields[name][value]=John+Doe&fields[email][value]=john@example.com

// Format 3: Direct JSON  
{
  "name": "John Doe",
  "email": "john@example.com"
}
```

### 3. WhatsApp Message Format

Recipients receive formatted messages:

```
*Nova submissão de formulário*
Data/Hora: 14/08/2025, 10:30:45

*Nome:* John Doe
*Email:* john@example.com
*Telefone:* +55 11 99999-9999
*Mensagem:* Interessado no produto X
```

## 🛠️ Development

### Project Structure

```
├── cmd/ewctl/           # CLI entry point
├── internal/            # Internal packages
│   ├── config/         # Configuration management
│   ├── database/       # Cloudflare D1 client
│   ├── tui/           # Terminal UI application
│   ├── views/         # UI views (forms, contacts, etc)
│   └── styles/        # UI theming
├── pkg/                # Public packages
│   └── webhook/       # Webhook client
├── worker.js          # Cloudflare Worker
├── schema.sql         # Database schema
└── wrangler.toml      # Worker configuration
```

### Building

```bash
# Development build
make dev

# Production build
make build

# Cross-platform builds
make build-all

# Run tests
make test

# Clean artifacts
make clean
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test ./internal/database -v

# Integration tests
make test-integration
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Built with these amazing tools:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The fun, functional TUI framework
- [Charm](https://charm.sh/) - Beautiful terminal tools and libraries
- [Cloudflare Workers](https://workers.cloudflare.com/) - Edge computing platform
- [Z-API](https://z-api.io/) - Reliable WhatsApp Business API
- [Elementor](https://elementor.com/) - WordPress page builder

Special thanks to the Charm team for creating such delightful terminal tools! 💖

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/thalysguimaraes">Thalys Guimarães</a>
</p>