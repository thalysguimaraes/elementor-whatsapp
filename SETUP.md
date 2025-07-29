# Setup Guide - Multi-Form Webhook System

This guide will help you set up the multi-form webhook management system.

## Prerequisites

- Node.js 16+
- Cloudflare account with Workers enabled
- Wrangler CLI configured (`wrangler login`)

## Step 1: Install Dependencies

```bash
# Install main dependencies
npm install

# Install manager dependencies
npm run setup
```

## Step 2: Create D1 Database

```bash
# Create the database
npm run db:create
```

This will output a database ID. Copy it and update `wrangler.toml`:

```toml
[[d1_databases]]
binding = "DB"
database_name = "elementor-whatsapp-forms"
database_id = "YOUR_DATABASE_ID_HERE"  # <-- Replace this
```

## Step 3: Initialize Database Schema

```bash
# Initialize the database with tables and default data
npm run db:init
```

## Step 4: Configure Manager CLI

Copy the environment file:

```bash
cp manager/.env.example manager/.env
```

Edit `manager/.env` with your details:

```env
# Get from Cloudflare dashboard > Account ID
CLOUDFLARE_ACCOUNT_ID=your_account_id

# Create at: https://dash.cloudflare.com/profile/api-tokens
# Permissions needed: D1:Edit
CLOUDFLARE_API_TOKEN=your_api_token

# The database ID from step 2
DATABASE_ID=your_database_id

# Your worker URL
WORKER_URL=https://elementor-whatsapp.thalys.workers.dev
```

## Step 5: Deploy the Worker

```bash
npm run deploy
```

## Step 6: Run the Manager

```bash
npm run panel
```

## Usage

### Creating a New Form

1. Run `npm run panel`
2. Select "Create New Form"
3. Enter form details
4. Add field mappings (match Elementor field IDs)
5. Add WhatsApp numbers
6. Copy the generated webhook URL to Elementor

### Testing

- Use the manager's "Test Webhook" option
- Or use the test scripts:
  ```bash
  ./test-webhook.sh         # Test default form
  ./test-json.sh           # Test with JSON
  ./test-health.sh         # Check service health
  ```

### Backward Compatibility

The old webhook URL still works:
- `https://your-worker.workers.dev/webhook/elementor`

It will use the hardcoded configuration from environment variables.

## Troubleshooting

### Database Connection Issues

Check if database is properly bound:
```bash
wrangler d1 execute elementor-whatsapp-forms --command="SELECT COUNT(*) FROM forms"
```

### Manager API Errors

Ensure your API token has the correct permissions:
- Account:Read
- D1:Edit

### Webhook Not Working

1. Check logs: `npm run tail`
2. Verify form exists: Use manager to list forms
3. Test with manager's test feature
4. Check Z-API credentials in `wrangler.toml`

## Architecture

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│   Elementor     │────▶│   Worker     │────▶│   Z-API     │
│     Forms       │     │  /webhook/:id│     │  WhatsApp   │
└─────────────────┘     └──────┬───────┘     └─────────────┘
                               │
                               ▼
                        ┌──────────────┐
                        │      D1      │
                        │   Database   │
                        └──────┬───────┘
                               │
                               ▼
                        ┌──────────────┐
                        │   Manager    │
                        │     CLI      │
                        └──────────────┘
```

## Security Notes

- Never commit `.env` files
- Keep API tokens secure
- Use read-only tokens where possible
- Regularly rotate credentials