-- Elementor WhatsApp Forms Database Schema

-- Forms table: stores form configurations
CREATE TABLE IF NOT EXISTS forms (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Form fields mapping: maps Elementor field IDs to friendly labels
CREATE TABLE IF NOT EXISTS form_fields (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  form_id TEXT NOT NULL,
  field_id TEXT NOT NULL,        -- The Elementor field ID (e.g., 'nome', 'field_cef3ba0')
  field_label TEXT NOT NULL,     -- The friendly label (e.g., 'Nome', 'Telefone')
  field_order INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE,
  UNIQUE(form_id, field_id)
);

-- Contacts table: centralized contact management
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

-- WhatsApp numbers per form
CREATE TABLE IF NOT EXISTS form_numbers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  form_id TEXT NOT NULL,
  phone_number TEXT NOT NULL,
  label TEXT,                    -- Optional label (e.g., 'Sales Manager')
  contact_id INTEGER,            -- Reference to contacts table
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE,
  FOREIGN KEY (contact_id) REFERENCES contacts(id),
  UNIQUE(form_id, phone_number)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_form_fields_form_id ON form_fields(form_id);
CREATE INDEX IF NOT EXISTS idx_form_numbers_form_id ON form_numbers(form_id);
CREATE INDEX IF NOT EXISTS idx_form_numbers_contact_id ON form_numbers(contact_id);
CREATE INDEX IF NOT EXISTS idx_contacts_phone ON contacts(phone_number);

-- Insert a default form (the current hardcoded configuration)
INSERT INTO forms (id, name, description) 
VALUES ('default', 'Default Form', 'Original hardcoded form configuration');

-- Insert default form fields
INSERT INTO form_fields (form_id, field_id, field_label, field_order) VALUES
  ('default', 'nome', 'Nome', 1),
  ('default', 'empresa', 'Empresa', 2),
  ('default', 'site', 'Site', 3),
  ('default', 'telefone', 'Telefone', 4),
  ('default', 'e-mail', 'E-mail', 5),
  ('default', 'quer adiantar alguma informação? (opcional)', 'Mensagem', 6),
  ('default', 'name', 'Nome', 1),
  ('default', 'message', 'Site', 3),
  ('default', 'field_cef3ba0', 'Telefone', 4),
  ('default', 'field_389b567', 'E-mail', 5),
  ('default', 'field_69b2d23', 'Mensagem', 6);

-- Insert default WhatsApp numbers
INSERT INTO form_numbers (form_id, phone_number, label) VALUES
  ('default', '5534984106712', 'Number 1'),
  ('default', '5534984106954', 'Number 2'),
  ('default', '5534991606334', 'Number 3'),
  ('default', '5534991517110', 'Number 4');

-- Monitoring tables (replace KV usage)
CREATE TABLE IF NOT EXISTS monitoring_state (
  key TEXT PRIMARY KEY,
  connected INTEGER NOT NULL,
  session INTEGER,
  status_json TEXT,
  last_changed DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS monitoring_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  key TEXT NOT NULL,
  connected INTEGER NOT NULL,
  session INTEGER,
  status_json TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_monitoring_history_key_id
  ON monitoring_history(key, id DESC);
