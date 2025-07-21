const axios = require('axios');

class WhatsAppService {
  constructor() {
    this.instanceId = process.env.Z_API_INSTANCE_ID;
    this.token = process.env.Z_API_TOKEN;
    this.clientToken = process.env.Z_API_CLIENT_TOKEN;
    this.baseUrl = 'https://api.z-api.io/instances';
  }

  async sendMessage(phone, message) {
    try {
      const url = `${this.baseUrl}/${this.instanceId}/token/${this.token}/send-text`;
      
      const response = await axios.post(url, {
        phone: phone,
        message: message
      }, {
        headers: {
          'Client-Token': this.clientToken,
          'Content-Type': 'application/json'
        }
      });

      console.log(`Message sent to ${phone}:`, response.data);
      return response.data;
    } catch (error) {
      console.error(`Failed to send message to ${phone}:`, error.response?.data || error.message);
      throw error;
    }
  }

  async sendToMultipleNumbers(numbers, message) {
    const results = [];
    
    for (const number of numbers) {
      try {
        const result = await this.sendMessage(number, message);
        results.push({ number, success: true, data: result });
      } catch (error) {
        results.push({ 
          number, 
          success: false, 
          error: error.response?.data || error.message 
        });
      }
    }
    
    return results;
  }
}

module.exports = new WhatsAppService();