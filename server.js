const express = require('express');
const axios = require('axios');
const cors = require('cors');
require('dotenv').config();

const whatsappService = require('./services/whatsapp');

const app = express();
const PORT = process.env.PORT || 3000;

app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.get('/', (req, res) => {
  res.json({ 
    status: 'OK', 
    message: 'Elementor WhatsApp Webhook Service',
    endpoints: {
      webhook: '/webhook/elementor'
    }
  });
});

app.post('/webhook/elementor', async (req, res) => {
  console.log('Received webhook:', JSON.stringify(req.body, null, 2));
  
  try {
    const formData = req.body;
    
    console.log('Environment check:', {
      hasInstanceId: !!process.env.Z_API_INSTANCE_ID,
      hasToken: !!process.env.Z_API_TOKEN,
      hasClientToken: !!process.env.Z_API_CLIENT_TOKEN,
      whatsappNumbers: [process.env.WHATSAPP_NUMBER_1, process.env.WHATSAPP_NUMBER_2, process.env.WHATSAPP_NUMBER_3].filter(Boolean)
    });
    
    const clientNumbers = [
      process.env.WHATSAPP_NUMBER_1,
      process.env.WHATSAPP_NUMBER_2,
      process.env.WHATSAPP_NUMBER_3
    ].filter(Boolean);
    
    if (clientNumbers.length === 0) {
      throw new Error('No WhatsApp numbers configured');
    }
    
    const message = formatMessage(formData);
    
    const results = await whatsappService.sendToMultipleNumbers(clientNumbers, message);
    
    const allSuccess = results.every(r => r.success);
    
    res.status(allSuccess ? 200 : 207).json({
      success: allSuccess,
      message: allSuccess ? 'Messages sent successfully' : 'Some messages failed',
      results: results
    });
    
  } catch (error) {
    console.error('Webhook error:', error);
    console.error('Error stack:', error.stack);
    res.status(500).json({
      success: false,
      error: error.message,
      details: error.response?.data || error.stack
    });
  }
});

function formatMessage(formData) {
  console.log('formatMessage received:', JSON.stringify(formData, null, 2));
  
  const timestamp = new Date().toLocaleString('pt-BR', { timeZone: 'America/Sao_Paulo' });
  
  let message = `*Nova submissão de formulário*\n`;
  message += `Data/Hora: ${timestamp}\n\n`;
  
  const fieldMapping = {
    'name': 'Nome',
    'empresa': 'Empresa',
    'message': 'Site',
    'field_cef3ba0': 'Telefone',
    'field_389b567': 'E-mail',
    'field_69b2d23': 'Mensagem'
  };
  
  const extractedFields = {};
  
  if (formData.fields && typeof formData.fields === 'object') {
    console.log('Processing fields object:', Object.keys(formData.fields));
    for (const [fieldId, fieldData] of Object.entries(formData.fields)) {
      if (fieldData && typeof fieldData === 'object' && fieldData.value) {
        extractedFields[fieldId] = fieldData.value;
      }
    }
  } else {
    console.log('No fields object found, using formData directly');
    Object.assign(extractedFields, formData);
  }
  
  console.log('Extracted fields:', extractedFields);
  
  for (const [fieldId, label] of Object.entries(fieldMapping)) {
    if (extractedFields[fieldId]) {
      message += `*${label}:* ${extractedFields[fieldId]}\n`;
    }
  }
  
  const mappedFields = Object.keys(fieldMapping);
  const unmappedFields = Object.entries(extractedFields)
    .filter(([key]) => !mappedFields.includes(key));
  
  if (unmappedFields.length > 0) {
    message += `\n*Outros campos:*\n`;
    unmappedFields.forEach(([key, value]) => {
      if (value && typeof value !== 'object') {
        message += `${key}: ${value}\n`;
      }
    });
  }
  
  return message;
}

app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});