import chalk from 'chalk';
import Table from 'cli-table3';
import ora from 'ora';
import { D1Client } from './database.js';
import * as prompts from './prompts.js';
import { CancelledError } from './prompts.js';

const db = new D1Client();

export async function listContacts() {
  const spinner = ora('Loading contacts...').start();
  
  try {
    const contacts = await db.getAllContacts();
    spinner.stop();
    
    if (contacts.length === 0) {
      console.log(chalk.yellow('\nNo contacts found. Add your first contact!\n'));
      return;
    }

    const table = new Table({
      head: ['ID', 'Name', 'Role', 'Company', 'Phone', 'Forms', 'Created'],
      style: { head: ['cyan'] },
      colWidths: [5, 20, 15, 15, 17, 7, 12],
    });

    for (const contact of contacts) {
      const formCount = await db.getContactFormCount(contact.id);
      table.push([
        contact.id.toString(),
        contact.name,
        contact.role || '-',
        contact.company || '-',
        contact.phone_number,
        formCount.toString(),
        new Date(contact.created_at).toLocaleDateString(),
      ]);
    }

    console.log('\n' + table.toString() + '\n');
  } catch (error) {
    spinner.fail(`Failed to load contacts: ${error.message}`);
  }
}

export async function createContact() {
  try {
    const contactDetails = await prompts.promptContactDetails();
    
    const spinner = ora('Creating contact...').start();
    
    await db.createContact(contactDetails);
    spinner.succeed('Contact created successfully!');
    
    console.log(chalk.green('\n✅ Contact added to your contact list!\n'));
    
  } catch (error) {
    if (error instanceof CancelledError) {
      console.log(chalk.gray('\nContact creation cancelled.\n'));
      return;
    }
    console.error(chalk.red(`\nError creating contact: ${error.message}\n`));
  }
}

export async function editContact() {
  const spinner = ora('Loading contacts...').start();
  
  try {
    const contacts = await db.getAllContacts();
    spinner.stop();
    
    if (contacts.length === 0) {
      console.log(chalk.yellow('\nNo contacts found to edit.\n'));
      return;
    }

    const contactId = await prompts.selectContact(contacts);
    const contact = await db.getContact(contactId);

    console.log(chalk.cyan(`\nEditing contact: ${contact.name}\n`));

    const updatedDetails = await prompts.promptContactDetails(contact);

    const updateSpinner = ora('Updating contact...').start();

    await db.updateContact(contactId, updatedDetails);

    updateSpinner.succeed('Contact updated successfully!');
    
    // Check if we need to update form_numbers
    if (updatedDetails.phone_number !== contact.phone_number) {
      const formCount = await db.getContactFormCount(contactId);
      if (formCount > 0) {
        console.log(chalk.yellow(`\n⚠️  Phone number changed. Updating ${formCount} form(s)...\n`));
      }
    }

  } catch (error) {
    if (error instanceof CancelledError) {
      console.log(chalk.gray('\nEdit cancelled.\n'));
      return;
    }
    console.error(chalk.red(`\nError editing contact: ${error.message}\n`));
  }
}

export async function deleteContact() {
  const spinner = ora('Loading contacts...').start();
  
  try {
    const contacts = await db.getAllContacts();
    spinner.stop();
    
    if (contacts.length === 0) {
      console.log(chalk.yellow('\nNo contacts found to delete.\n'));
      return;
    }

    const contactId = await prompts.selectContact(contacts, 'Select contact to delete:');
    const contact = await db.getContact(contactId);
    const formCount = await db.getContactFormCount(contactId);

    if (formCount > 0) {
      console.log(chalk.yellow(`\n⚠️  This contact is used in ${formCount} form(s).\n`));
      const formsUsingContact = await db.getFormsUsingContact(contactId);
      console.log(chalk.gray('Forms using this contact:'));
      formsUsingContact.forEach(form => {
        console.log(chalk.gray(`  - ${form.name} (${form.id})`));
      });
    }

    const confirmDelete = await prompts.confirmAction(
      `Are you sure you want to delete ${contact.name}?`
    );

    if (!confirmDelete) {
      console.log(chalk.gray('\nDeletion cancelled.\n'));
      return;
    }

    const deleteSpinner = ora('Deleting contact...').start();

    await db.deleteContact(contactId);

    deleteSpinner.succeed('Contact deleted successfully!');

  } catch (error) {
    if (error instanceof CancelledError) {
      console.log(chalk.gray('\nDeletion cancelled.\n'));
      return;
    }
    console.error(chalk.red(`\nError deleting contact: ${error.message}\n`));
  }
}

export async function viewContactDetails() {
  const spinner = ora('Loading contacts...').start();
  
  try {
    const contacts = await db.getAllContacts();
    spinner.stop();
    
    if (contacts.length === 0) {
      console.log(chalk.yellow('\nNo contacts found.\n'));
      return;
    }

    const contactId = await prompts.selectContact(contacts, 'Select contact to view:');
    const contact = await db.getContact(contactId);
    const formsUsingContact = await db.getFormsUsingContact(contactId);

    console.log(chalk.cyan('\n📇 Contact Details\n'));
    console.log(chalk.gray('Name:       ') + contact.name);
    console.log(chalk.gray('Phone:      ') + contact.phone_number);
    console.log(chalk.gray('Role:       ') + (contact.role || '-'));
    console.log(chalk.gray('Company:    ') + (contact.company || '-'));
    console.log(chalk.gray('Notes:      ') + (contact.notes || '-'));
    console.log(chalk.gray('Created:    ') + new Date(contact.created_at).toLocaleString());
    console.log(chalk.gray('Updated:    ') + new Date(contact.updated_at).toLocaleString());

    if (formsUsingContact.length > 0) {
      console.log(chalk.cyan('\n📋 Used in Forms:\n'));
      formsUsingContact.forEach(form => {
        console.log(chalk.gray(`  - ${form.name} (${form.id})`));
      });
    } else {
      console.log(chalk.yellow('\n⚠️  Not used in any forms yet.\n'));
    }

  } catch (error) {
    if (error instanceof CancelledError) {
      console.log(chalk.gray('\nView cancelled.\n'));
      return;
    }
    console.error(chalk.red(`\nError viewing contact: ${error.message}\n`));
  }
}