import fetch from 'node-fetch';
import dotenv from 'dotenv';

dotenv.config();

export class D1Client {
  constructor() {
    this.accountId = process.env.CLOUDFLARE_ACCOUNT_ID;
    this.apiToken = process.env.CLOUDFLARE_API_TOKEN;
    this.databaseId = process.env.DATABASE_ID;
    this.baseUrl = `https://api.cloudflare.com/client/v4/accounts/${this.accountId}/d1/database/${this.databaseId}`;
  }

  async query(sql, params = []) {
    try {
      const response = await fetch(`${this.baseUrl}/query`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.apiToken}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          sql,
          params,
        }),
      });

      const data = await response.json();
      
      if (!response.ok) {
        throw new Error(data.errors?.[0]?.message || 'Database query failed');
      }

      return data.result[0];
    } catch (error) {
      console.error('Database error:', error.message);
      throw error;
    }
  }

  // Contact operations
  async getAllContacts() {
    const result = await this.query(`
      SELECT * FROM contacts
      ORDER BY name ASC
    `);
    return result.results || [];
  }

  async getContact(id) {
    const result = await this.query(
      'SELECT * FROM contacts WHERE id = ? LIMIT 1',
      [id]
    );
    return result.results[0];
  }

  async createContact(contact) {
    const result = await this.query(
      `INSERT INTO contacts (phone_number, name, company, role, notes)
       VALUES (?, ?, ?, ?, ?)`,
      [contact.phone_number, contact.name, contact.company, contact.role, contact.notes]
    );
    return result.meta.last_row_id;
  }

  async updateContact(id, contact) {
    await this.query(
      `UPDATE contacts 
       SET phone_number = ?, name = ?, company = ?, role = ?, notes = ?, 
           updated_at = CURRENT_TIMESTAMP
       WHERE id = ?`,
      [contact.phone_number, contact.name, contact.company, contact.role, contact.notes, id]
    );
  }

  async deleteContact(id) {
    // First remove references from form_numbers
    await this.query(
      'UPDATE form_numbers SET contact_id = NULL WHERE contact_id = ?',
      [id]
    );
    
    // Then delete the contact
    await this.query('DELETE FROM contacts WHERE id = ?', [id]);
  }

  async getContactFormCount(contactId) {
    const result = await this.query(
      'SELECT COUNT(DISTINCT form_id) as count FROM form_numbers WHERE contact_id = ?',
      [contactId]
    );
    return result.results[0]?.count || 0;
  }

  async getFormsUsingContact(contactId) {
    const result = await this.query(
      `SELECT DISTINCT f.id, f.name 
       FROM forms f
       JOIN form_numbers fn ON f.id = fn.form_id
       WHERE fn.contact_id = ?`,
      [contactId]
    );
    return result.results || [];
  }

  async getContactByPhone(phoneNumber) {
    const result = await this.query(
      'SELECT * FROM contacts WHERE phone_number = ? LIMIT 1',
      [phoneNumber]
    );
    return result.results[0];
  }

  // Forms operations
  async getAllForms() {
    const result = await this.query(`
      SELECT f.*, 
        COUNT(DISTINCT ff.id) as field_count,
        COUNT(DISTINCT fn.id) as number_count
      FROM forms f
      LEFT JOIN form_fields ff ON f.id = ff.form_id
      LEFT JOIN form_numbers fn ON f.id = fn.form_id
      GROUP BY f.id
      ORDER BY f.created_at DESC
    `);
    return result.results || [];
  }

  async getForm(formId) {
    const formResult = await this.query(
      'SELECT * FROM forms WHERE id = ?',
      [formId]
    );
    
    if (!formResult.results || formResult.results.length === 0) {
      return null;
    }

    const form = formResult.results[0];

    // Get fields
    const fieldsResult = await this.query(
      'SELECT * FROM form_fields WHERE form_id = ? ORDER BY field_order',
      [formId]
    );

    // Get numbers
    const numbersResult = await this.query(
      'SELECT * FROM form_numbers WHERE form_id = ? ORDER BY id',
      [formId]
    );

    return {
      ...form,
      fields: fieldsResult.results || [],
      numbers: numbersResult.results || [],
    };
  }

  async createForm(form) {
    // Insert form
    await this.query(
      'INSERT INTO forms (id, name, description) VALUES (?, ?, ?)',
      [form.id, form.name, form.description]
    );

    // Insert fields
    for (let i = 0; i < form.fields.length; i++) {
      const field = form.fields[i];
      await this.query(
        'INSERT INTO form_fields (form_id, field_id, field_label, field_order) VALUES (?, ?, ?, ?)',
        [form.id, field.field_id, field.field_label, i]
      );
    }

    // Insert numbers with contact references
    for (const number of form.numbers) {
      await this.query(
        'INSERT INTO form_numbers (form_id, phone_number, label, contact_id) VALUES (?, ?, ?, ?)',
        [form.id, number.phone_number, number.label, number.contact_id || null]
      );
    }

    return form;
  }

  async updateForm(formId, updates) {
    // Update form details
    if (updates.name || updates.description) {
      await this.query(
        'UPDATE forms SET name = COALESCE(?, name), description = COALESCE(?, description), updated_at = CURRENT_TIMESTAMP WHERE id = ?',
        [updates.name, updates.description, formId]
      );
    }

    // Update fields if provided
    if (updates.fields) {
      // Delete existing fields
      await this.query('DELETE FROM form_fields WHERE form_id = ?', [formId]);
      
      // Insert new fields
      for (let i = 0; i < updates.fields.length; i++) {
        const field = updates.fields[i];
        await this.query(
          'INSERT INTO form_fields (form_id, field_id, field_label, field_order) VALUES (?, ?, ?, ?)',
          [formId, field.field_id, field.field_label, i]
        );
      }
    }

    // Update numbers if provided
    if (updates.numbers) {
      // Delete existing numbers
      await this.query('DELETE FROM form_numbers WHERE form_id = ?', [formId]);
      
      // Insert new numbers with contact references
      for (const number of updates.numbers) {
        await this.query(
          'INSERT INTO form_numbers (form_id, phone_number, label, contact_id) VALUES (?, ?, ?, ?)',
          [formId, number.phone_number, number.label, number.contact_id || null]
        );
      }
    }
  }

  async deleteForm(formId) {
    await this.query('DELETE FROM forms WHERE id = ?', [formId]);
  }
}