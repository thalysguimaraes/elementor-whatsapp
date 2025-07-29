#!/usr/bin/env node

import dotenv from 'dotenv';
import chalk from 'chalk';
import { mainMenu, contactsMenu } from './lib/prompts.js';
import * as forms from './lib/forms.js';
import * as contacts from './lib/contacts.js';

dotenv.config();

// Check required environment variables
function checkEnv() {
  const required = ['CLOUDFLARE_ACCOUNT_ID', 'CLOUDFLARE_API_TOKEN', 'DATABASE_ID', 'WORKER_URL'];
  const missing = required.filter(key => !process.env[key]);
  
  if (missing.length > 0) {
    console.error(chalk.red('\n❌ Missing required environment variables:'));
    missing.forEach(key => console.error(chalk.red(`  - ${key}`)));
    console.error(chalk.gray('\nCopy .env.example to .env and fill in the values.\n'));
    process.exit(1);
  }
}

async function main() {
  checkEnv();
  
  let running = true;
  
  while (running) {
    let actionCompleted = false;
    
    try {
      const action = await mainMenu();
      
      switch (action) {
        case 'list':
          await forms.listForms();
          actionCompleted = true;
          break;
          
        case 'create':
          await forms.createForm();
          actionCompleted = true;
          break;
          
        case 'edit':
          await forms.editForm();
          actionCompleted = true;
          break;
          
        case 'delete':
          await forms.deleteForm();
          actionCompleted = true;
          break;
          
        case 'test':
          await forms.testWebhook();
          actionCompleted = true;
          break;
          
        case 'export':
          await forms.exportConfiguration();
          actionCompleted = true;
          break;
          
        case 'contacts':
          await manageContacts();
          actionCompleted = true;
          break;
          
        case 'exit':
          running = false;
          console.log(chalk.gray('\nGoodbye! 👋\n'));
          break;
      }
      
      if (running && actionCompleted) {
        // Pause before returning to menu
        await new Promise(resolve => {
          console.log(chalk.gray('\nPress Enter to continue...'));
          process.stdin.once('data', resolve);
        });
      }
      
    } catch (error) {
      if (error.name !== 'CancelledError') {
        console.error(chalk.red(`\nError: ${error.message}\n`));
        // Still pause on real errors
        await new Promise(resolve => {
          console.log(chalk.gray('\nPress Enter to continue...'));
          process.stdin.once('data', resolve);
        });
      }
      // For cancelled operations, just go back to menu immediately
    }
  }
}

async function manageContacts() {
  let inContactsMenu = true;
  
  while (inContactsMenu) {
    let actionCompleted = false;
    
    try {
      const action = await contactsMenu();
      
      switch (action) {
        case 'list':
          await contacts.listContacts();
          actionCompleted = true;
          break;
          
        case 'create':
          await contacts.createContact();
          actionCompleted = true;
          break;
          
        case 'edit':
          await contacts.editContact();
          actionCompleted = true;
          break;
          
        case 'delete':
          await contacts.deleteContact();
          actionCompleted = true;
          break;
          
        case 'view':
          await contacts.viewContactDetails();
          actionCompleted = true;
          break;
          
        case 'back':
          inContactsMenu = false;
          break;
      }
      
      if (inContactsMenu && actionCompleted) {
        // Pause before returning to contacts menu
        await new Promise(resolve => {
          console.log(chalk.gray('\nPress Enter to continue...'));
          process.stdin.once('data', resolve);
        });
      }
      
    } catch (error) {
      if (error.name !== 'CancelledError') {
        console.error(chalk.red(`\nError: ${error.message}\n`));
        // Still pause on real errors
        await new Promise(resolve => {
          console.log(chalk.gray('\nPress Enter to continue...'));
          process.stdin.once('data', resolve);
        });
      }
      // For cancelled operations, just go back to menu immediately
    }
  }
}

// Run the CLI
main().catch(error => {
  console.error(chalk.red(`\nFatal error: ${error.message}\n`));
  process.exit(1);
});