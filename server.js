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
  console.log('Received webhook:', req.body);
  
  try {
    const formData = req.body;
    
    const clientNumbers = [
      process.env.WHATSAPP_NUMBER_1,
      process.env.WHATSAPP_NUMBER_2
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
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

function formatMessage(formData) {
  const timestamp = new Date().toLocaleString('pt-BR', { timeZone: 'America/Sao_Paulo' });
  
  let message = `*Nova submissão de formulário*\n`;
  message += `Data/Hora: ${timestamp}\n\n`;
  
  const fields = {
    'name': 'Nome',
    'nome': 'Nome',
    'email': 'E-mail',
    'phone': 'Telefone',
    'telefone': 'Telefone',
    'message': 'Mensagem',
    'mensagem': 'Mensagem',
    'subject': 'Assunto',
    'assunto': 'Assunto'
  };
  
  for (const [key, label] of Object.entries(fields)) {
    if (formData[key]) {
      message += `*${label}:* ${formData[key]}\n`;
    }
  }
  
  const knownFields = Object.keys(fields);
  const customFields = Object.entries(formData)
    .filter(([key]) => !knownFields.includes(key.toLowerCase()));
  
  if (customFields.length > 0) {
    message += `\n*Outros campos:*\n`;
    customFields.forEach(([key, value]) => {
      message += `${key}: ${value}\n`;
    });
  }
  
  return message;
}

app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});