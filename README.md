# ewctl - Elementor WhatsApp Manager

![ewctl Banner](banner-image.jpg)

A terminal UI for managing Elementor form webhooks that send WhatsApp notifications. Built with Go and Bubble Tea because I needed a better way to manage form submissions without constantly switching between dashboards.

## Why I Built This

I was managing multiple WordPress sites with Elementor forms that needed to send WhatsApp notifications. The workflow was painful - configuring webhooks in Elementor, managing contact numbers in environment variables, debugging failed submissions through Cloudflare logs. I wanted a single tool to handle everything from my terminal.

The original version was a Node.js CLI with Inquirer.js, but I decided to rewrite it in Go with Bubble Tea for better performance and a more enjoyable development experience. Plus, I really wanted to learn Bubble Tea and this seemed like the perfect project.

## What It Does

- **Manages Forms**: Create and configure webhook endpoints for your Elementor forms
- **Handles Contacts**: Store WhatsApp numbers and assign them to forms (no more hardcoded numbers!)
- **Processes Webhooks**: Cloudflare Worker receives form submissions and sends WhatsApp messages
- **Tests Webhooks**: Send test payloads to debug your setup
- **Shows Statistics**: Dashboard with real-time stats from Cloudflare D1

## Technical Architecture

```
Elementor Form → Cloudflare Worker → Z-API → WhatsApp
                        ↓
                  Cloudflare D1
                        ↑
                     ewctl TUI
```

The Cloudflare Worker handles the webhook processing at the edge for low latency. Form configurations and contacts are stored in D1 (Cloudflare's distributed SQLite). The TUI connects directly to D1's API to manage everything.

## Stack Choices

- **Go + Bubble Tea**: I wanted a compiled binary for easy distribution and Bubble Tea makes beautiful TUIs
- **Cloudflare Workers**: Already using Cloudflare for other projects, Workers are fast and have a generous free tier
- **Cloudflare D1**: Needed a database that works well with Workers, D1 is SQLite at the edge
- **Z-API**: Most reliable WhatsApp Business API I've found (tried several others)

## Installation

### Prerequisites

- Go 1.21+ (if building from source)
- Cloudflare account
- Z-API account for WhatsApp
- Elementor Pro with webhook support

### Build from Source

```bash
git clone https://github.com/thalysguimaraes/elementor-whatsapp.git
cd elementor-whatsapp
make build
./bin/ewctl
```

### Download Binary

Check the [releases page](https://github.com/thalysguimaraes/elementor-whatsapp/releases) for pre-built binaries.

## Configuration

Create `~/.config/ewctl/config.yaml`:

```yaml
cloudflare:
  account_id: your-account-id
  api_token: your-api-token  # needs D1 edit permissions
  database_id: your-database-id
  worker_url: https://your-worker.workers.dev

zapi:
  instance_id: your-instance-id
  instance_token: your-instance-token
  client_token: your-client-token

ui:
  theme: charm  # I like the Charm theme
  mouse: true
  animations: true
```

## Usage

Start the TUI:
```bash
ewctl
```

Navigation:
- Number keys (1-5) for quick navigation from dashboard
- `n` to create new form, `a` to add contact
- `e` to edit, `d` to delete
- `ESC` to go back
- Mouse works too if you prefer clicking

## Setting Up Webhooks

1. Deploy the worker:
```bash
wrangler deploy
```

2. Initialize the database:
```bash
wrangler d1 execute elementor-whatsapp-forms --file=./schema.sql --remote
```

3. In Elementor, add a webhook action to your form:
   - URL: `https://your-worker.workers.dev/webhook/{form-id}`
   - The form-id comes from ewctl when you create a form

## How It Handles Elementor Data

Elementor sends form data in different formats depending on the setup. The worker detects and handles three formats:

1. Nested JSON with form metadata
2. URL-encoded form data
3. Direct JSON payload

The field mapping is configured per form in ewctl, so you can map Elementor field IDs to readable labels in the WhatsApp message.

## Project Structure

```
cmd/ewctl/          # CLI entry point
internal/
  tui/              # Bubble Tea application
  views/            # Different screens (forms, contacts, etc)
  database/         # Cloudflare D1 client
  config/           # Configuration handling
worker.js           # Cloudflare Worker that processes webhooks
schema.sql          # Database structure
```

## Development

```bash
# Run tests
make test

# Build for all platforms
make build-all

# Development mode with hot reload
make dev
```

The codebase follows standard Go project layout. The TUI is built with Bubble Tea's Model-Update-View pattern. Each view is its own model with Init, Update, and View methods.

## Things I Learned

- Bubble Tea is fantastic for building TUIs - the elm-inspired architecture makes state management clear
- Cloudflare D1's HTTP API is straightforward but has some quirks with data types
- Huh (the form library) doesn't play perfectly with dynamic data, had to work around some limitations
- Go's compile-time guarantees caught many bugs that would've been runtime errors in the Node version

## Limitations

- D1 API requires explicit type handling for query parameters
- The TUI requires a TTY (won't work in CI/CD pipelines)
- Z-API rate limits apply (usually not an issue for form submissions)

## Future Ideas

- [ ] Support for other messaging platforms (Telegram, Slack)
- [ ] Webhook replay for failed deliveries
- [ ] Form analytics dashboard
- [ ] Multi-language message templates
- [ ] Backup/restore functionality

## Contributing

Feel free to open issues or PRs. The codebase is fairly straightforward - start with `cmd/ewctl/main.go` and follow the imports.

## License

MIT - see [LICENSE](LICENSE)

## Credits

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) and the amazing Charm libraries
- [Cloudflare Workers](https://workers.cloudflare.com/) and D1
- [Z-API](https://z-api.io/) for WhatsApp

Special thanks to the Charm team - their work on terminal UIs is inspiring.

---

If you're dealing with Elementor forms and WhatsApp notifications, I hope this helps. Feel free to reach out if you have questions.

[@thalysguimaraes](https://github.com/thalysguimaraes)