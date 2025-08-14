#!/bin/bash

echo "Setting up Cloudflare Workers secrets..."
echo "Please have your API keys ready."
echo ""

read -p "Enter ZAPI Instance ID: " ZAPI_INSTANCE_ID
echo "$ZAPI_INSTANCE_ID" | wrangler secret put ZAPI_INSTANCE_ID

read -p "Enter ZAPI Instance Token: " ZAPI_INSTANCE_TOKEN
echo "$ZAPI_INSTANCE_TOKEN" | wrangler secret put ZAPI_INSTANCE_TOKEN

read -p "Enter ZAPI Client Token: " ZAPI_CLIENT_TOKEN
echo "$ZAPI_CLIENT_TOKEN" | wrangler secret put ZAPI_CLIENT_TOKEN

read -p "Enter Resend API Key: " RESEND_API_KEY
echo "$RESEND_API_KEY" | wrangler secret put RESEND_API_KEY

echo ""
echo "✅ All secrets have been configured!"
echo ""
echo "To verify your secrets are set, run:"
echo "wrangler secret list"