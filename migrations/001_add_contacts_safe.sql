-- Migration: Add contacts table and update form_numbers (safe version)
-- Created: 2025-01-29

-- Create contacts table if not exists
CREATE TABLE IF NOT EXISTS contacts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  phone_number TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  company TEXT,
  role TEXT,
  notes TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create index for phone number lookups
CREATE INDEX IF NOT EXISTS idx_contacts_phone ON contacts(phone_number);

-- Add contact_id to form_numbers table if it doesn't exist
-- SQLite doesn't support ALTER TABLE ADD COLUMN IF NOT EXISTS, so we need to check first
-- This will be done manually after checking the schema