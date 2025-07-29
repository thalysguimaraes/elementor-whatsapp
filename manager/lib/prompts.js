import inquirer from 'inquirer';
import chalk from 'chalk';
import { nanoid } from 'nanoid';

// Custom error class for cancelled operations
export class CancelledError extends Error {
  constructor() {
    super('Operation cancelled');
    this.name = 'CancelledError';
  }
}

export async function mainMenu() {
  console.clear();
  console.log(chalk.cyan.bold('\n📋 Elementor WhatsApp Manager\n'));
  
  const { action } = await inquirer.prompt([
    {
      type: 'list',
      name: 'action',
      message: 'What would you like to do?',
      choices: [
        { name: '📜 List Forms', value: 'list' },
        { name: '➕ Create New Form', value: 'create' },
        { name: '✏️  Edit Form', value: 'edit' },
        { name: '🗑️  Delete Form', value: 'delete' },
        { name: '🧪 Test Webhook', value: 'test' },
        { name: '📤 Export Configuration', value: 'export' },
        new inquirer.Separator(),
        { name: '📞 Manage Contacts', value: 'contacts' },
        new inquirer.Separator(),
        { name: '❌ Exit', value: 'exit' },
      ],
    },
  ]);

  return action;
}

export async function contactsMenu() {
  console.clear();
  console.log(chalk.gray('Elementor WhatsApp Manager > ') + chalk.cyan.bold('Contact Management'));
  console.log(chalk.gray('─'.repeat(50)));
  
  const { action } = await inquirer.prompt([
    {
      type: 'list',
      name: 'action',
      message: 'What would you like to do?',
      choices: [
        { name: '📋 List All Contacts', value: 'list' },
        { name: '➕ Add New Contact', value: 'create' },
        { name: '✏️  Edit Contact', value: 'edit' },
        { name: '🗑️  Delete Contact', value: 'delete' },
        { name: '👁️  View Contact Details', value: 'view' },
        new inquirer.Separator(),
        { name: '⬅️  Back to Main Menu', value: 'back' },
      ],
    },
  ]);

  return action;
}

export async function promptFormDetails(existingForm = null) {
  const action = existingForm ? 'Edit Form' : 'Create Form';
  console.log(chalk.gray('\nElementor WhatsApp Manager > Forms > ') + chalk.yellow(action));
  console.log(chalk.gray('─'.repeat(50)));
  console.log(chalk.yellow('\n📝 Form Details\n'));
  
  const formDetails = await inquirer.prompt([
    {
      type: 'input',
      name: 'name',
      message: 'Form Name:',
      default: existingForm?.name,
      validate: (input) => input.trim() !== '' || 'Form name is required',
    },
    {
      type: 'input',
      name: 'description',
      message: 'Form Description:',
      default: existingForm?.description,
    },
    {
      type: 'input',
      name: 'id',
      message: 'Form ID (leave blank to auto-generate):',
      default: existingForm?.id,
      filter: (input) => {
        if (!input.trim() && !existingForm) {
          return nanoid(10).toLowerCase();
        }
        return input.trim().toLowerCase().replace(/[^a-z0-9-_]/g, '-');
      },
    },
  ]);

  return formDetails;
}

export async function promptFields(existingFields = []) {
  console.log(chalk.yellow('\n🏷️  Form Fields Mapping\n'));
  console.log(chalk.gray('Map Elementor field IDs to friendly labels'));
  
  const fields = [];
  let addMore = true;

  // Show existing fields if editing
  if (existingFields.length > 0) {
    console.log(chalk.gray('\nExisting fields:'));
    existingFields.forEach(field => {
      console.log(chalk.gray(`  ${field.field_label}: ${field.field_id}`));
    });
    
    const { keepExisting } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'keepExisting',
        message: 'Keep existing fields?',
        default: true,
      },
    ]);

    if (keepExisting) {
      fields.push(...existingFields);
      const { addNew } = await inquirer.prompt([
        {
          type: 'confirm',
          name: 'addNew',
          message: 'Add new fields?',
          default: false,
        },
      ]);
      addMore = addNew;
    }
  }

  while (addMore) {
    const field = await inquirer.prompt([
      {
        type: 'input',
        name: 'field_label',
        message: 'Field Label (e.g., "Nome"):',
        validate: (input) => input.trim() !== '' || 'Field label is required',
      },
      {
        type: 'input',
        name: 'field_id',
        message: 'Field ID from Elementor (e.g., "nome" or "field_cef3ba0"):',
        validate: (input) => input.trim() !== '' || 'Field ID is required',
      },
    ]);

    fields.push(field);

    const { continue: cont } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'continue',
        message: 'Add another field?',
        default: fields.length < 6,
      },
    ]);

    addMore = cont;
  }

  return fields;
}

export async function promptWhatsAppNumbers(existingNumbers = []) {
  console.log(chalk.yellow('\n📱 WhatsApp Numbers\n'));
  
  const numbers = [];
  let addMore = true;

  // Show existing numbers if editing
  if (existingNumbers.length > 0) {
    console.log(chalk.gray('\nExisting numbers:'));
    existingNumbers.forEach(num => {
      console.log(chalk.gray(`  ${num.phone_number}${num.label ? ` (${num.label})` : ''}`));
    });
    
    const { keepExisting } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'keepExisting',
        message: 'Keep existing numbers?',
        default: true,
      },
    ]);

    if (keepExisting) {
      numbers.push(...existingNumbers);
      const { addNew } = await inquirer.prompt([
        {
          type: 'confirm',
          name: 'addNew',
          message: 'Add new numbers?',
          default: false,
        },
      ]);
      addMore = addNew;
    }
  }

  while (addMore) {
    const number = await inquirer.prompt([
      {
        type: 'input',
        name: 'phone_number',
        message: 'WhatsApp Number (with country code):',
        validate: (input) => {
          const cleaned = input.replace(/\D/g, '');
          return cleaned.length >= 10 || 'Please enter a valid phone number';
        },
        filter: (input) => input.replace(/\D/g, ''),
      },
      {
        type: 'input',
        name: 'label',
        message: 'Label (optional, e.g., "Sales Manager"):',
      },
    ]);

    numbers.push(number);

    const { continue: cont } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'continue',
        message: 'Add another number?',
        default: numbers.length < 4,
      },
    ]);

    addMore = cont;
  }

  return numbers;
}

export async function selectForm(forms, message = 'Select a form:', allowCancel = true) {
  const choices = forms.map(form => ({
    name: `${form.name} (${form.id}) - ${form.field_count} fields, ${form.number_count} numbers`,
    value: form.id,
  }));
  
  if (allowCancel) {
    choices.push(new inquirer.Separator());
    choices.push({ name: '❌ Cancel', value: 'CANCEL' });
  }

  const { formId } = await inquirer.prompt([
    {
      type: 'list',
      name: 'formId',
      message,
      choices,
    },
  ]);

  if (formId === 'CANCEL') {
    throw new CancelledError();
  }

  return formId;
}

export async function confirmAction(message) {
  const { confirm } = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'confirm',
      message,
      default: false,
    },
  ]);

  return confirm;
}

export async function promptContactDetails(existingContact = null) {
  const action = existingContact ? 'Edit Contact' : 'Create Contact';
  console.log(chalk.gray('\nElementor WhatsApp Manager > Contacts > ') + chalk.yellow(action));
  console.log(chalk.gray('─'.repeat(50)));
  console.log(chalk.yellow('\n📇 Contact Details\n'));
  
  const contactDetails = await inquirer.prompt([
    {
      type: 'input',
      name: 'name',
      message: 'Contact Name:',
      default: existingContact?.name,
      validate: (input) => input.trim() !== '' || 'Contact name is required',
    },
    {
      type: 'input',
      name: 'phone_number',
      message: 'WhatsApp Number (with country code):',
      default: existingContact?.phone_number,
      validate: (input) => {
        const cleaned = input.replace(/\D/g, '');
        return cleaned.length >= 10 || 'Please enter a valid phone number';
      },
      filter: (input) => input.replace(/\D/g, ''),
    },
    {
      type: 'input',
      name: 'role',
      message: 'Role/Position (optional):',
      default: existingContact?.role,
    },
    {
      type: 'input',
      name: 'company',
      message: 'Company (optional):',
      default: existingContact?.company,
    },
    {
      type: 'input',
      name: 'notes',
      message: 'Notes (optional):',
      default: existingContact?.notes,
    },
  ]);

  return contactDetails;
}

export async function selectContact(contacts, message = 'Select a contact:', allowCancel = true) {
  const choices = contacts.map(contact => ({
    name: `${contact.name} - ${contact.role || 'No role'} (${contact.phone_number})`,
    value: contact.id,
  }));
  
  if (allowCancel) {
    choices.push(new inquirer.Separator());
    choices.push({ name: '❌ Cancel', value: 'CANCEL' });
  }

  const { contactId } = await inquirer.prompt([
    {
      type: 'list',
      name: 'contactId',
      message,
      choices,
    },
  ]);

  if (contactId === 'CANCEL') {
    throw new CancelledError();
  }

  return contactId;
}

export async function selectContactsForForm(allContacts, selectedIds = []) {
  console.log(chalk.yellow('\n📞 Select WhatsApp Recipients\n'));
  
  const choices = allContacts.map(contact => ({
    name: `${contact.name} - ${contact.role || 'No role'} (${contact.phone_number})`,
    value: contact.id,
    checked: selectedIds.includes(contact.id),
  }));
  
  choices.push(new inquirer.Separator());
  choices.push({ name: '➕ Add new contact...', value: 'NEW' });

  const { selected } = await inquirer.prompt([
    {
      type: 'checkbox',
      name: 'selected',
      message: 'Select contacts to receive messages from this form: (Space to select, Enter to confirm)',
      choices,
      validate: (answer) => {
        if (answer.length === 0 || (answer.length === 1 && answer[0] === 'NEW')) {
          return 'You must select at least one contact';
        }
        return true;
      },
    },
  ]);

  return selected;
}

export async function promptTestData(fields) {
  console.log(chalk.yellow('\n🧪 Test Data\n'));
  console.log(chalk.gray('Enter test values for each field:'));
  
  const testData = {};
  
  for (const field of fields) {
    const { value } = await inquirer.prompt([
      {
        type: 'input',
        name: 'value',
        message: `${field.field_label} (${field.field_id}):`,
        default: getDefaultTestValue(field.field_label),
      },
    ]);
    
    testData[field.field_id] = value;
  }

  return testData;
}

function getDefaultTestValue(label) {
  const lowerLabel = label.toLowerCase();
  
  if (lowerLabel.includes('nome') || lowerLabel.includes('name')) {
    return 'Teste CLI';
  } else if (lowerLabel.includes('empresa') || lowerLabel.includes('company')) {
    return 'Empresa Teste';
  } else if (lowerLabel.includes('email') || lowerLabel.includes('e-mail')) {
    return 'teste@example.com';
  } else if (lowerLabel.includes('telefone') || lowerLabel.includes('phone')) {
    return '(34) 99999-9999';
  } else if (lowerLabel.includes('site') || lowerLabel.includes('website')) {
    return 'www.example.com';
  } else if (lowerLabel.includes('mensagem') || lowerLabel.includes('message')) {
    return 'Mensagem de teste do CLI';
  }
  
  return '';
}