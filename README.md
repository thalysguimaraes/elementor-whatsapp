# ewctl — Elementor WhatsApp Manager

A terminal UI for managing Elementor form webhooks that send WhatsApp notifications via Z-API. Built with Go and Bubble Tea.

```
Elementor Form -> Cloudflare Worker -> Z-API -> WhatsApp
                        |
                  Cloudflare D1
                        |
                     ewctl TUI
```

## Features

- **Form management** — create and configure webhook endpoints for Elementor forms
- **Contact management** — store WhatsApp numbers and assign them to forms
- **Webhook processing** — Cloudflare Worker receives submissions, sends WhatsApp messages
- **Webhook testing** — send test payloads to debug your setup
- **Statistics dashboard** — real-time stats from Cloudflare D1

## Install

```bash
git clone https://github.com/thalysguimaraes/elementor-whatsapp.git
cd elementor-whatsapp
make build
./bin/ewctl
```

Pre-built binaries available on the [releases page](https://github.com/thalysguimaraes/elementor-whatsapp/releases).

Requires a Cloudflare account, Z-API account, and Elementor Pro.

## Configuration

Create `~/.config/ewctl/config.yaml`:

```yaml
cloudflare:
  account_id: your-account-id
  api_token: your-api-token
  database_id: your-database-id
  worker_url: https://your-worker.workers.dev

zapi:
  instance_id: your-instance-id
  instance_token: your-instance-token
  client_token: your-client-token
```

## Usage

```bash
ewctl   # launch the TUI
```

- Number keys `1`-`5` for quick navigation
- `n` to create form, `a` to add contact
- `e` to edit, `d` to delete, `Esc` to go back

### Setting up webhooks

1. Deploy the worker: `wrangler deploy`
2. Init the database: `wrangler d1 execute elementor-whatsapp-forms --file=./schema.sql --remote`
3. In Elementor, add a webhook action with URL: `https://your-worker.workers.dev/webhook/{form-id}`

## Built with

Go, [Bubble Tea](https://github.com/charmbracelet/bubbletea), Cloudflare Workers, D1, Z-API

## License

MIT
