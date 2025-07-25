export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    
    // Log all requests for debugging
    console.log(`${request.method} ${url.pathname}`);
    console.log('Headers:', Object.fromEntries(request.headers));
    
    if (url.pathname === '/webhook/elementor' && request.method === 'POST') {
      try {
        // Get raw body for logging
        const rawBody = await request.text();
        console.log('Raw body received:', rawBody);
        
        // Parse the body
        let data;
        try {
          data = JSON.parse(rawBody);
        } catch (e) {
          console.error('Failed to parse JSON:', e);
          // Try URL-encoded format
          const params = new URLSearchParams(rawBody);
          data = Object.fromEntries(params);
        }
        
        console.log('Parsed data:', JSON.stringify(data, null, 2));
        
        // Try different possible field locations
        const fields = data.fields || data.form_fields || data;
        console.log('Fields object:', JSON.stringify(fields, null, 2));
        
        const fieldMapping = {
          'nome': 'Nome',
          'empresa': 'Empresa',
          'site': 'Site',
          'telefone': 'Telefone',
          'e-mail': 'E-mail',
          'quer adiantar alguma informação? (opcional)': 'Mensagem'
        };
        
        const now = new Date();
        const dateStr = now.toLocaleDateString('pt-BR');
        const timeStr = now.toLocaleTimeString('pt-BR');
        
        let message = `*Nova submissão de formulário*\nData/Hora: ${dateStr}, ${timeStr}\n\n`;
        let hasFields = false;
        
        for (const [fieldId, label] of Object.entries(fieldMapping)) {
          // Try different ways to access field values
          let value;
          if (fields[fieldId]?.value !== undefined) {
            value = fields[fieldId].value;
          } else if (fields[fieldId] !== undefined && typeof fields[fieldId] === 'string') {
            value = fields[fieldId];
          } else if (data[fieldId] !== undefined) {
            value = data[fieldId];
          }
          
          if (value) {
            message += `*${label}:* ${value}\n`;
            hasFields = true;
          }
        }
        
        // If no fields were found, log all available data
        if (!hasFields) {
          console.log('No fields found in expected format. All data:', JSON.stringify(data, null, 2));
          message += '\nDados brutos:\n' + JSON.stringify(data, null, 2);
        }
        
        console.log('Mensagem formatada:', message);
        
        const numbers = [
          env.WHATSAPP_NUMBER_1,
          env.WHATSAPP_NUMBER_2,
          env.WHATSAPP_NUMBER_3
        ].filter(Boolean);
        
        const zapiUrl = `https://api.z-api.io/instances/${env.ZAPI_INSTANCE_ID}/token/${env.ZAPI_INSTANCE_TOKEN}/send-text`;
        
        const results = await Promise.all(
          numbers.map(async (phone) => {
            try {
              const response = await fetch(zapiUrl, {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                  'Client-Token': env.ZAPI_CLIENT_TOKEN
                },
                body: JSON.stringify({ phone, message })
              });
              
              const result = await response.json();
              console.log(`Mensagem enviada para ${phone}:`, result);
              return { phone, success: true, result };
            } catch (error) {
              console.error(`Erro ao enviar para ${phone}:`, error);
              return { phone, success: false, error: error.message };
            }
          })
        );
        
        const successful = results.filter(r => r.success).length;
        const failed = results.filter(r => !r.success).length;
        
        return new Response(JSON.stringify({
          success: true,
          message: `Mensagens enviadas: ${successful} sucesso, ${failed} falhas`,
          results
        }), {
          status: 200,
          headers: { 
            'Content-Type': 'application/json',
            'Access-Control-Allow-Origin': '*'
          }
        });
        
      } catch (error) {
        console.error('Erro ao processar webhook:', error);
        return new Response(JSON.stringify({
          success: false,
          error: error.message
        }), {
          status: 500,
          headers: { 
            'Content-Type': 'application/json',
            'Access-Control-Allow-Origin': '*'
          }
        });
      }
    }
    
    // Handle CORS preflight
    if (request.method === 'OPTIONS') {
      return new Response(null, {
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
          'Access-Control-Allow-Headers': 'Content-Type',
        }
      });
    }
    
    if (url.pathname === '/' && request.method === 'GET') {
      return new Response(JSON.stringify({
        status: 'ok',
        service: 'Elementor WhatsApp Webhook',
        version: '2.1.0'
      }), {
        headers: { 
          'Content-Type': 'application/json',
          'Access-Control-Allow-Origin': '*'
        }
      });
    }
    
    return new Response('Not Found', { 
      status: 404,
      headers: { 'Access-Control-Allow-Origin': '*' }
    });
  }
};