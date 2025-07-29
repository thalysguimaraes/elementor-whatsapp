-- Migration: Add contacts table and update form_numbers
-- Created: 2025-01-29

-- Create contacts table
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

-- Add contact_id to form_numbers table
ALTER TABLE form_numbers ADD COLUMN contact_id INTEGER REFERENCES contacts(id);

-- Create index for contact lookups
CREATE INDEX IF NOT EXISTS idx_form_numbers_contact_id ON form_numbers(contact_id);

-- Migrate existing numbers to contacts
-- This creates a contact for each unique phone number in form_numbers
INSERT INTO contacts (phone_number, name)
SELECT DISTINCT 
  phone_number,
  COALESCE(label, 'Contact ' || ROW_NUMBER() OVER (ORDER BY phone_number)) as name
FROM form_numbers
WHERE phone_number IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM contacts c WHERE c.phone_number = form_numbers.phone_number
  );

-- Update form_numbers to link to the new contacts
UPDATE form_numbers
SET contact_id = (
  SELECT id FROM contacts WHERE contacts.phone_number = form_numbers.phone_number
)
WHERE phone_number IS NOT NULL;