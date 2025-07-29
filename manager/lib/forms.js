import chalk from 'chalk';
import Table from 'cli-table3';
import ora from 'ora';
import fetch from 'node-fetch';
import { D1Client } from './database.js';
import * as prompts from './prompts.js';
import { CancelledError } from './prompts.js';

const db = new D1Client();

export async function listForms() {
  const spinner = ora('Loading forms...').start();
  
  try {
    const forms = await db.getAllForms();
    spinner.stop();
    
    if (forms.length === 0) {
      console.log(chalk.yellow('\nNo forms found. Create your first form!\n'));
      return;
    }

    const table = new Table({
      head: ['ID', 'Name', 'Fields', 'Numbers', 'Created'],
      style: { head: ['cyan'] },
    });

    forms.forEach(form => {
      table.push([
        form.id,
        form.name,
        form.field_count.toString(),
        form.number_count.toString(),
        new Date(form.created_at).toLocaleDateString(),
      ]);
    });

    console.log('\n' + table.toString() + '\n');
  } catch (error) {
    spinner.fail(`Failed to load forms: ${error.message}`);
  }
}

export async function createForm() {
  try {
    const formDetails = await prompts.promptFormDetails();
    const fields = await prompts.promptFields();
    
    // Get all contacts
    const allContacts = await db.getAllContacts();
    
    if (allContacts.length === 0) {
      console.log(chalk.yellow('\nNo contacts found. Please add contacts first.\n'));
      return;
    }
    
    // Select contacts for this form
    let selectedContactIds = await prompts.selectContactsForForm(allContacts);
    
    // Handle "Add new contact" option
    while (selectedContactIds.includes('NEW')) {
      const newContact = await prompts.promptContactDetails();
      const newContactId = await db.createContact(newContact);
      
      // Refresh contacts list
      const updatedContacts = await db.getAllContacts();
      
      // Re-prompt with updated list, pre-selecting the new contact
      selectedContactIds = selectedContactIds.filter(id => id !== 'NEW');
      selectedContactIds.push(newContactId);
      
      const newSelection = await prompts.selectContactsForForm(updatedContacts, selectedContactIds);
      selectedContactIds = newSelection;
    }
    
    // Get selected contacts with their details
    const selectedContacts = allContacts.filter(c => selectedContactIds.includes(c.id));
    const numbers = selectedContacts.map(contact => ({
      phone_number: contact.phone_number,
      label: contact.name,
      contact_id: contact.id
    }));

    const spinner = ora('Creating form...').start();

    const form = {
      ...formDetails,
      fields,
      numbers,
    };

    await db.createForm(form);
    spinner.succeed('Form created successfully!');

    const webhookUrl = `${process.env.WORKER_URL}/webhook/${form.id}`;
    
    console.log(chalk.green('\n✅ Form created successfully!\n'));
    console.log(chalk.cyan('📎 Webhook URL:'), chalk.yellow(webhookUrl));
    console.log(chalk.gray('\nCopy this URL to your Elementor form webhook action.\n'));

  } catch (error) {
    if (error instanceof CancelledError) {
      console.log(chalk.gray('\nForm creation cancelled.\n'));
      return;
    }
    console.error(chalk.red(`\nError creating form: ${error.message}\n`));
  }
}

export async function editForm() {
  const spinner = ora('Loading forms...').start();
  
  try {
    const forms = await db.getAllForms();
    spinner.stop();
    
    if (forms.length === 0) {
      console.log(chalk.yellow('\nNo forms found to edit.\n'));
      return;
    }

    const formId = await prompts.selectForm(forms, 'Select form to edit:');
    const form = await db.getForm(formId);

    console.log(chalk.cyan(`\nEditing form: ${form.name}\n`));

    const formDetails = await prompts.promptFormDetails(form);
    const fields = await prompts.promptFields(form.fields);
    
    // Get all contacts
    const allContacts = await db.getAllContacts();
    
    // Get current contact IDs for this form
    const currentContactIds = form.numbers
      .filter(n => n.contact_id)
      .map(n => n.contact_id);
    
    // Select contacts for this form
    let selectedContactIds = await prompts.selectContactsForForm(allContacts, currentContactIds);
    
    // Handle "Add new contact" option
    while (selectedContactIds.includes('NEW')) {
      const newContact = await prompts.promptContactDetails();
      const newContactId = await db.createContact(newContact);
      
      // Refresh contacts list
      const updatedContacts = await db.getAllContacts();
      
      // Re-prompt with updated list, pre-selecting the new contact
      selectedContactIds = selectedContactIds.filter(id => id !== 'NEW');
      selectedContactIds.push(newContactId);
      
      const newSelection = await prompts.selectContactsForForm(updatedContacts, selectedContactIds);
      selectedContactIds = newSelection;
    }
    
    // Get selected contacts with their details
    const selectedContacts = await Promise.all(
      selectedContactIds.map(id => db.getContact(id))
    );
    const numbers = selectedContacts.map(contact => ({
      phone_number: contact.phone_number,
      label: contact.name,
      contact_id: contact.id
    }));

    const updateSpinner = ora('Updating form...').start();

    await db.updateForm(formId, {
      name: formDetails.name,
      description: formDetails.description,
      fields,
      numbers,
    });

    updateSpinner.succeed('Form updated successfully!');
    
  } catch (error) {
    if (spinner && spinner.isSpinning) spinner.stop();
    if (error instanceof CancelledError) {
      console.log(chalk.gray('\nEdit cancelled.\n'));
      return;
    }
    console.error(chalk.red(`\nError editing form: ${error.message}\n`));
  }
}

export async function deleteForm() {
  const spinner = ora('Loading forms...').start();
  
  try {
    const forms = await db.getAllForms();
    spinner.stop();
    
    if (forms.length === 0) {
      console.log(chalk.yellow('\nNo forms found to delete.\n'));
      return;
    }

    const formId = await prompts.selectForm(forms, 'Select form to delete:');
    const form = await db.getForm(formId);

    console.log(chalk.red(`\n⚠️  Warning: This will delete "${form.name}" permanently!\n`));
    
    const confirmed = await prompts.confirmAction('Are you sure you want to delete this form?');
    
    if (!confirmed) {
      console.log(chalk.gray('\nDeletion cancelled.\n'));
      return;
    }

    const deleteSpinner = ora('Deleting form...').start();
    await db.deleteForm(formId);
    deleteSpinner.succeed('Form deleted successfully!');
    
  } catch (error) {
    if (spinner && spinner.isSpinning) spinner.stop();
    if (error instanceof CancelledError) {
      console.log(chalk.gray('\nDeletion cancelled.\n'));
      return;
    }
    console.error(chalk.red(`\nError deleting form: ${error.message}\n`));
  }
}

export async function testWebhook() {
  const spinner = ora('Loading forms...').start();
  
  try {
    const forms = await db.getAllForms();
    spinner.stop();
    
    if (forms.length === 0) {
      console.log(chalk.yellow('\nNo forms found to test.\n'));
      return;
    }

    const formId = await prompts.selectForm(forms, 'Select form to test:');
    const form = await db.getForm(formId);

    console.log(chalk.cyan(`\nTesting webhook for: ${form.name}\n`));

    const testData = await prompts.promptTestData(form.fields);
    
    const webhookUrl = `${process.env.WORKER_URL}/webhook/${form.id}`;
    
    const testSpinner = ora('Sending test webhook...').start();
    
    try {
      const response = await fetch(webhookUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'User-Agent': 'Elementor-WhatsApp-Manager/Test',
        },
        body: JSON.stringify(testData),
      });

      const result = await response.json();
      
      if (response.ok) {
        testSpinner.succeed('Test webhook sent successfully!');
        console.log(chalk.green('\n✅ Response:'), JSON.stringify(result, null, 2));
      } else {
        testSpinner.fail('Test webhook failed!');
        console.log(chalk.red('\n❌ Error:'), JSON.stringify(result, null, 2));
      }
    } catch (error) {
      testSpinner.fail(`Failed to send test webhook: ${error.message}`);
    }
    
  } catch (error) {
    if (spinner && spinner.isSpinning) spinner.stop();
    if (error instanceof CancelledError) {
      console.log(chalk.gray('\nTest cancelled.\n'));
      return;
    }
    console.error(chalk.red(`\nError testing webhook: ${error.message}\n`));
  }
}

export async function exportConfiguration() {
  const spinner = ora('Loading forms...').start();
  
  try {
    const forms = await db.getAllForms();
    spinner.stop();
    
    if (forms.length === 0) {
      console.log(chalk.yellow('\nNo forms found to export.\n'));
      return;
    }

    const formId = await prompts.selectForm(forms, 'Select form to export:');
    const form = await db.getForm(formId);

    const exportData = {
      form: {
        id: form.id,
        name: form.name,
        description: form.description,
      },
      fields: form.fields.map(f => ({
        field_id: f.field_id,
        field_label: f.field_label,
      })),
      numbers: form.numbers.map(n => ({
        phone_number: n.phone_number,
        label: n.label,
      })),
      webhook_url: `${process.env.WORKER_URL}/webhook/${form.id}`,
    };

    console.log(chalk.cyan('\n📤 Form Configuration:\n'));
    console.log(JSON.stringify(exportData, null, 2));
    console.log();
    
  } catch (error) {
    if (spinner && spinner.isSpinning) spinner.stop();
    if (error instanceof CancelledError) {
      console.log(chalk.gray('\nExport cancelled.\n'));
      return;
    }
    console.error(chalk.red(`\nError exporting configuration: ${error.message}\n`));
  }
}