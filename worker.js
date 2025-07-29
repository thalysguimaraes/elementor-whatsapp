export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const startTime = Date.now();
    
    // Enhanced structured logging
    const logRequest = {
      timestamp: new Date().toISOString(),
      method: request.method,
      path: url.pathname,
      headers: {
        contentType: request.headers.get('content-type'),
        userAgent: request.headers.get('user-agent'),
        origin: request.headers.get('origin'),
        contentLength: request.headers.get('content-length')
      },
      cf: request.cf ? {
        country: request.cf.country,
        city: request.cf.city,
        asn: request.cf.asn
      } : null
    };
    
    console.log(JSON.stringify({
      type: 'request_received',
      ...logRequest
    }));
    
    // Health check endpoint
    if (url.pathname === '/health' && request.method === 'GET') {
      const healthCheck = await performHealthCheck(env);
      return new Response(JSON.stringify(healthCheck), {
        status: healthCheck.status === 'healthy' ? 200 : 503,
        headers: { 
          'Content-Type': 'application/json',
          'Access-Control-Allow-Origin': '*'
        }
      });
    }
    
    // Test email endpoint (temporary - remove in production)
    if (url.pathname === '/test-email' && request.method === 'GET') {
      try {
        await sendAlertEmail(env, {
          subject: '🧪 Teste de Email - Sistema de Monitoramento',
          issue: 'Este é um email de teste',
          details: {
            message: 'Se você recebeu este email, o sistema de alertas está funcionando corretamente!',
            timestamp: new Date().toISOString(),
            zapiStatus: await checkZAPIStatus(env)
          }
        });
        
        return new Response(JSON.stringify({
          success: true,
          message: 'Test email sent to ' + env.ALERT_EMAIL
        }), {
          headers: { 
            'Content-Type': 'application/json',
            'Access-Control-Allow-Origin': '*'
          }
        });
      } catch (error) {
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
    
    // Dynamic webhook endpoint with form ID
    const webhookMatch = url.pathname.match(/^\/webhook\/(.+)$/);
    if (webhookMatch && request.method === 'POST') {
      const formId = webhookMatch[1];
      
      try {
        // Get raw body for logging
        const rawBody = await request.text();
        console.log(JSON.stringify({
          type: 'webhook_raw_body',
          timestamp: new Date().toISOString(),
          formId,
          size: rawBody.length,
          preview: rawBody.substring(0, 500)
        }));
        
        // Parse the body
        let data;
        try {
          data = JSON.parse(rawBody);
        } catch (e) {
          console.log(JSON.stringify({
            type: 'parse_attempt_url_encoded',
            timestamp: new Date().toISOString()
          }));
          // Try URL-encoded format
          const params = new URLSearchParams(rawBody);
          data = Object.fromEntries(params);
        }
        
        console.log(JSON.stringify({
          type: 'parsed_data',
          timestamp: new Date().toISOString(),
          formId,
          dataKeys: Object.keys(data),
          dataStructure: JSON.stringify(data, null, 2).substring(0, 1000)
        }));
        
        // Validate required environment variables
        const validationResult = validateEnvironment(env);
        if (!validationResult.valid) {
          console.error(JSON.stringify({
            type: 'environment_validation_failed',
            timestamp: new Date().toISOString(),
            errors: validationResult.errors
          }));
          
          return new Response(JSON.stringify({
            success: false,
            error: 'Configuration error',
            details: validationResult.errors
          }), {
            status: 500,
            headers: { 
              'Content-Type': 'application/json',
              'Access-Control-Allow-Origin': '*'
            }
          });
        }
        
        // Get form configuration from database
        const formConfig = await getFormConfiguration(env, formId);
        
        if (!formConfig) {
          console.log(JSON.stringify({
            type: 'form_not_found',
            timestamp: new Date().toISOString(),
            formId
          }));
          
          return new Response(JSON.stringify({
            success: false,
            error: 'Form not found',
            formId
          }), {
            status: 404,
            headers: { 
              'Content-Type': 'application/json',
              'Access-Control-Allow-Origin': '*'
            }
          });
        }
        
        console.log(JSON.stringify({
          type: 'form_config_loaded',
          timestamp: new Date().toISOString(),
          formId,
          formName: formConfig.name,
          fieldCount: formConfig.fields.length,
          numberCount: formConfig.numbers.length
        }));
        
        // Extract and validate form data
        const extractedFields = extractFormFields(data, formConfig.fields);
        const validation = validateFormData(extractedFields, formConfig.fields);
        
        if (!validation.valid) {
          console.log(JSON.stringify({
            type: 'form_validation_failed',
            timestamp: new Date().toISOString(),
            formId,
            errors: validation.errors,
            receivedFields: extractedFields
          }));
          
          return new Response(JSON.stringify({
            success: false,
            error: 'Invalid form data',
            details: validation.errors,
            receivedFields: Object.keys(extractedFields)
          }), {
            status: 400,
            headers: { 
              'Content-Type': 'application/json',
              'Access-Control-Allow-Origin': '*'
            }
          });
        }
        
        // Format message
        const message = formatWhatsAppMessage(extractedFields, formConfig.fields);
        
        console.log(JSON.stringify({
          type: 'message_formatted',
          timestamp: new Date().toISOString(),
          formId,
          messageLength: message.length,
          fields: Object.keys(extractedFields)
        }));
        
        // Get numbers from form configuration
        const numbers = formConfig.numbers.map(n => n.phone_number);
        
        // Send messages
        const zapiUrl = `https://api.z-api.io/instances/${env.ZAPI_INSTANCE_ID}/token/${env.ZAPI_INSTANCE_TOKEN}/send-text`;
        
        const results = await Promise.all(
          numbers.map(async (phone) => {
            const sendStart = Date.now();
            try {
              const response = await fetch(zapiUrl, {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                  'Client-Token': env.ZAPI_CLIENT_TOKEN
                },
                body: JSON.stringify({ phone, message })
              });
              
              const responseText = await response.text();
              let result;
              try {
                result = JSON.parse(responseText);
              } catch (e) {
                result = { raw: responseText };
              }
              
              const sendDuration = Date.now() - sendStart;
              
              console.log(JSON.stringify({
                type: 'whatsapp_send_result',
                timestamp: new Date().toISOString(),
                formId,
                phone,
                success: response.ok,
                statusCode: response.status,
                duration: sendDuration,
                response: result
              }));
              
              return { 
                phone, 
                success: response.ok, 
                statusCode: response.status,
                duration: sendDuration,
                result 
              };
            } catch (error) {
              const sendDuration = Date.now() - sendStart;
              console.error(JSON.stringify({
                type: 'whatsapp_send_error',
                timestamp: new Date().toISOString(),
                formId,
                phone,
                error: error.message,
                stack: error.stack,
                duration: sendDuration
              }));
              
              return { 
                phone, 
                success: false, 
                error: error.message,
                errorType: error.name,
                duration: sendDuration
              };
            }
          })
        );
        
        const successful = results.filter(r => r.success).length;
        const failed = results.filter(r => !r.success).length;
        const totalDuration = Date.now() - startTime;
        
        const responseLog = {
          type: 'webhook_completed',
          timestamp: new Date().toISOString(),
          formId,
          formName: formConfig.name,
          duration: totalDuration,
          successful,
          failed,
          totalNumbers: numbers.length,
          formFields: Object.keys(extractedFields),
          results: results
        };
        
        console.log(JSON.stringify(responseLog));
        
        return new Response(JSON.stringify({
          success: failed === 0,
          form: formConfig.name,
          message: `Mensagens enviadas: ${successful} sucesso, ${failed} falhas`,
          duration: totalDuration,
          results
        }), {
          status: failed === 0 ? 200 : 207,
          headers: { 
            'Content-Type': 'application/json',
            'Access-Control-Allow-Origin': '*',
            'X-Processing-Time': totalDuration.toString()
          }
        });
        
      } catch (error) {
        const duration = Date.now() - startTime;
        console.error(JSON.stringify({
          type: 'webhook_error',
          timestamp: new Date().toISOString(),
          formId,
          duration,
          error: error.message,
          stack: error.stack,
          errorType: error.name
        }));
        
        return new Response(JSON.stringify({
          success: false,
          error: error.message,
          errorType: error.name,
          duration
        }), {
          status: 500,
          headers: { 
            'Content-Type': 'application/json',
            'Access-Control-Allow-Origin': '*',
            'X-Processing-Time': duration.toString()
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
          'Access-Control-Max-Age': '86400'
        }
      });
    }
    
    // Root endpoint
    if (url.pathname === '/' && request.method === 'GET') {
      return new Response(JSON.stringify({
        status: 'ok',
        service: 'Elementor WhatsApp Webhook',
        version: '3.2.0',
        endpoints: {
          webhook: '/webhook/:formId',
          health: '/health',
          testEmail: '/test-email'
        },
        monitoring: {
          enabled: env.MONITORING_ENABLED === 'true',
          interval: '15 minutes'
        },
        documentation: 'https://github.com/thalysguimaraes/elementor-whatsapp'
      }), {
        headers: { 
          'Content-Type': 'application/json',
          'Access-Control-Allow-Origin': '*'
        }
      });
    }
    
    // 404 for unknown routes
    console.log(JSON.stringify({
      type: 'route_not_found',
      timestamp: new Date().toISOString(),
      method: request.method,
      path: url.pathname
    }));
    
    return new Response(JSON.stringify({
      error: 'Not Found',
      path: url.pathname,
      method: request.method
    }), { 
      status: 404,
      headers: { 
        'Content-Type': 'application/json',
        'Access-Control-Allow-Origin': '*' 
      }
    });
  },

  // Scheduled handler for cron monitoring
  async scheduled(event, env, ctx) {
    if (env.MONITORING_ENABLED !== 'true') {
      console.log('Monitoring is disabled');
      return;
    }

    console.log(JSON.stringify({
      type: 'monitoring_check_start',
      timestamp: new Date().toISOString(),
      cron: event.cron
    }));

    try {
      // Check Z-API status
      const zapiStatus = await checkZAPIStatus(env);
      
      // Get previous state from KV
      const previousState = await env.MONITOR_STATE.get('zapi-status', { type: 'json' }) || {};
      
      console.log(JSON.stringify({
        type: 'monitoring_status',
        timestamp: new Date().toISOString(),
        current: zapiStatus,
        previous: previousState
      }));

      // Check if status changed from connected to disconnected
      if (previousState.connected === true && zapiStatus.connected === false) {
        console.log('Z-API disconnection detected, sending alert...');
        await sendAlertEmail(env, {
          subject: '🚨 Z-API WhatsApp Desconectado',
          issue: 'WhatsApp desconectado',
          details: zapiStatus
        });
      }

      // Check if status changed from disconnected to connected
      if (previousState.connected === false && zapiStatus.connected === true) {
        console.log('Z-API reconnection detected, sending recovery notification...');
        await sendAlertEmail(env, {
          subject: '✅ Z-API WhatsApp Reconectado',
          issue: 'WhatsApp reconectado',
          details: zapiStatus
        });
      }

      // Store current state
      await env.MONITOR_STATE.put('zapi-status', JSON.stringify({
        ...zapiStatus,
        lastChecked: new Date().toISOString()
      }));

      // Store monitoring history
      const history = await env.MONITOR_STATE.get('monitoring-history', { type: 'json' }) || [];
      history.push({
        timestamp: new Date().toISOString(),
        connected: zapiStatus.connected,
        session: zapiStatus.session
      });
      
      // Keep only last 100 entries
      if (history.length > 100) {
        history.splice(0, history.length - 100);
      }
      
      await env.MONITOR_STATE.put('monitoring-history', JSON.stringify(history));

    } catch (error) {
      console.error(JSON.stringify({
        type: 'monitoring_error',
        timestamp: new Date().toISOString(),
        error: error.message,
        stack: error.stack
      }));

      // Send error alert
      await sendAlertEmail(env, {
        subject: '⚠️ Erro no Monitoramento Z-API',
        issue: 'Erro ao verificar status',
        details: { error: error.message }
      });
    }
  }
};

// Helper functions (keeping all existing ones)

function validateEnvironment(env) {
  const errors = [];
  
  if (!env.ZAPI_INSTANCE_ID) errors.push('ZAPI_INSTANCE_ID not configured');
  if (!env.ZAPI_INSTANCE_TOKEN) errors.push('ZAPI_INSTANCE_TOKEN not configured');
  if (!env.ZAPI_CLIENT_TOKEN) errors.push('ZAPI_CLIENT_TOKEN not configured');
  
  return {
    valid: errors.length === 0,
    errors
  };
}

async function getFormConfiguration(env, formId) {
  try {
    // Special handling for legacy 'elementor' endpoint
    if (formId === 'elementor') {
      // Return default configuration for backward compatibility
      return {
        id: 'default',
        name: 'Default Form',
        fields: [
          { field_id: 'nome', field_label: 'Nome' },
          { field_id: 'empresa', field_label: 'Empresa' },
          { field_id: 'site', field_label: 'Site' },
          { field_id: 'telefone', field_label: 'Telefone' },
          { field_id: 'e-mail', field_label: 'E-mail' },
          { field_id: 'quer adiantar alguma informação? (opcional)', field_label: 'Mensagem' },
          { field_id: 'name', field_label: 'Nome' },
          { field_id: 'message', field_label: 'Site' },
          { field_id: 'field_cef3ba0', field_label: 'Telefone' },
          { field_id: 'field_389b567', field_label: 'E-mail' },
          { field_id: 'field_69b2d23', field_label: 'Mensagem' }
        ],
        numbers: [] // Legacy numbers removed - must be configured via CLI
      };
    }
    
    // Query D1 database for form configuration
    const { results: forms } = await env.DB.prepare(
      'SELECT * FROM forms WHERE id = ?'
    ).bind(formId).all();
    
    if (!forms || forms.length === 0) {
      return null;
    }
    
    const form = forms[0];
    
    // Get fields
    const { results: fields } = await env.DB.prepare(
      'SELECT * FROM form_fields WHERE form_id = ? ORDER BY field_order'
    ).bind(formId).all();
    
    // Get numbers
    const { results: numbers } = await env.DB.prepare(
      'SELECT * FROM form_numbers WHERE form_id = ? ORDER BY id'
    ).bind(formId).all();
    
    return {
      ...form,
      fields: fields || [],
      numbers: numbers || []
    };
  } catch (error) {
    console.error(JSON.stringify({
      type: 'database_error',
      timestamp: new Date().toISOString(),
      formId,
      error: error.message,
      stack: error.stack
    }));
    
    // Fallback to default configuration if database fails
    if (formId === 'default' || formId === 'elementor') {
      return getDefaultConfiguration(env);
    }
    
    return null;
  }
}

function getDefaultConfiguration(env) {
  return {
    id: 'default',
    name: 'Default Form',
    fields: [
      { field_id: 'nome', field_label: 'Nome' },
      { field_id: 'empresa', field_label: 'Empresa' },
      { field_id: 'site', field_label: 'Site' },
      { field_id: 'telefone', field_label: 'Telefone' },
      { field_id: 'e-mail', field_label: 'E-mail' },
      { field_id: 'quer adiantar alguma informação? (opcional)', field_label: 'Mensagem' },
      { field_id: 'name', field_label: 'Nome' },
      { field_id: 'message', field_label: 'Site' },
      { field_id: 'field_cef3ba0', field_label: 'Telefone' },
      { field_id: 'field_389b567', field_label: 'E-mail' },
      { field_id: 'field_69b2d23', field_label: 'Mensagem' }
    ],
    numbers: [] // Legacy numbers removed - must be configured via CLI
  };
}

function extractFormFields(data, fieldConfig) {
  const extractedFields = {};
  const fieldMap = {};
  
  // Create a map of field IDs for quick lookup
  fieldConfig.forEach(field => {
    fieldMap[field.field_id] = field.field_label;
  });
  
  // Check if we have nested JSON structure (data.fields exists and is an object)
  if (data.fields && typeof data.fields === 'object' && !Array.isArray(data.fields)) {
    console.log(JSON.stringify({
      type: 'extraction_nested_fields',
      timestamp: new Date().toISOString(),
      fieldKeys: Object.keys(data.fields)
    }));
    
    for (const [fieldId, fieldData] of Object.entries(data.fields)) {
      if (fieldData && typeof fieldData === 'object' && fieldData.value !== undefined) {
        if (fieldMap[fieldId]) {
          extractedFields[fieldId] = fieldData.value;
        }
      }
    }
  } else {
    // Check for direct field names first (simple JSON format)
    let foundDirectFields = false;
    for (const fieldId of Object.keys(fieldMap)) {
      if (data[fieldId] !== undefined) {
        extractedFields[fieldId] = data[fieldId];
        foundDirectFields = true;
      }
    }
    
    // If no direct fields found, check for flattened URL-encoded format
    if (!foundDirectFields) {
      console.log(JSON.stringify({
        type: 'extraction_flattened_format',
        timestamp: new Date().toISOString(),
        dataKeys: Object.keys(data)
      }));
      
      for (const [key, value] of Object.entries(data)) {
        const match = key.match(/^fields\[([^\]]+)\]\[value\]$/);
        if (match) {
          const fieldId = match[1];
          if (fieldMap[fieldId]) {
            extractedFields[fieldId] = value;
          }
        }
      }
    }
  }
  
  console.log(JSON.stringify({
    type: 'fields_extracted',
    timestamp: new Date().toISOString(),
    extractedCount: Object.keys(extractedFields).length,
    extractedKeys: Object.keys(extractedFields)
  }));
  
  return extractedFields;
}

function validateFormData(fields, fieldConfig) {
  const errors = [];
  
  // Check if we have at least some fields
  if (Object.keys(fields).length === 0) {
    errors.push('No recognized form fields found');
  }
  
  // Optional: Add specific field validations here
  // For example, validate email format, phone format, etc.
  
  return {
    valid: errors.length === 0,
    errors
  };
}

function formatWhatsAppMessage(fields, fieldConfig) {
  const now = new Date();
  const dateStr = now.toLocaleDateString('pt-BR');
  const timeStr = now.toLocaleTimeString('pt-BR');
  
  let message = `*Nova submissão de formulário*\nData/Hora: ${dateStr}, ${timeStr}\n\n`;
  
  // Create a map for deduplication
  const addedLabels = new Set();
  
  // Add fields in the order they're configured
  fieldConfig.forEach(config => {
    const value = fields[config.field_id];
    if (value && !addedLabels.has(config.field_label)) {
      message += `*${config.field_label}:* ${value}\n`;
      addedLabels.add(config.field_label);
    }
  });
  
  return message;
}

async function performHealthCheck(env) {
  const checks = {
    service: 'healthy',
    configuration: 'unknown',
    database: 'unknown',
    zapi: 'unknown'
  };
  
  // Check configuration
  const configValidation = validateEnvironment(env);
  checks.configuration = configValidation.valid ? 'healthy' : 'unhealthy';
  
  // Check database connectivity
  try {
    const { results } = await env.DB.prepare('SELECT COUNT(*) as count FROM forms').all();
    checks.database = 'healthy';
  } catch (error) {
    checks.database = 'unhealthy';
  }
  
  // Check Z-API connectivity
  const zapiStatus = await checkZAPIStatus(env);
  checks.zapi = zapiStatus.connected ? 'healthy' : 'unhealthy';
  
  const overallStatus = Object.values(checks).every(status => status === 'healthy') ? 'healthy' : 'degraded';
  
  return {
    status: overallStatus,
    timestamp: new Date().toISOString(),
    version: '3.1.0',
    checks,
    errors: configValidation.errors,
    zapiDetails: zapiStatus
  };
}

// New monitoring functions

async function checkZAPIStatus(env) {
  try {
    const zapiUrl = `https://api.z-api.io/instances/${env.ZAPI_INSTANCE_ID}/token/${env.ZAPI_INSTANCE_TOKEN}/status`;
    const response = await fetch(zapiUrl, {
      method: 'GET',
      headers: {
        'Client-Token': env.ZAPI_CLIENT_TOKEN
      }
    });
    
    if (!response.ok) {
      return {
        connected: false,
        session: false,
        error: `API Error: ${response.status}`,
        timestamp: new Date().toISOString()
      };
    }
    
    const status = await response.json();
    return {
      ...status,
      timestamp: new Date().toISOString()
    };
  } catch (error) {
    return {
      connected: false,
      session: false,
      error: error.message,
      timestamp: new Date().toISOString()
    };
  }
}

async function sendAlertEmail(env, alert) {
  if (!env.ALERT_EMAIL) {
    console.warn('ALERT_EMAIL not configured, skipping email notification');
    return;
  }

  const emailBody = `
Alerta do Sistema de Webhook WhatsApp

${alert.issue}

Detalhes:
${JSON.stringify(alert.details, null, 2)}

Timestamp: ${new Date().toISOString()}
Instance ID: ${env.ZAPI_INSTANCE_ID}

---
Este é um alerta automático do sistema de monitoramento.
  `.trim();

  // Log the alert
  console.log(JSON.stringify({
    type: 'email_alert',
    timestamp: new Date().toISOString(),
    to: env.ALERT_EMAIL,
    subject: alert.subject,
    body: emailBody
  }));

  // Send email via Resend
  if (env.RESEND_API_KEY) {
    try {
      const response = await fetch('https://api.resend.com/emails', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${env.RESEND_API_KEY}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          from: env.RESEND_FROM_EMAIL || 'onboarding@resend.dev',
          to: [env.ALERT_EMAIL],
          subject: alert.subject,
          text: emailBody,
          html: `<pre>${emailBody.replace(/\n/g, '<br>')}</pre>`
        })
      });
      
      if (!response.ok) {
        const errorData = await response.text();
        console.error('Failed to send email via Resend:', errorData);
      } else {
        const result = await response.json();
        console.log('Email sent successfully via Resend:', result);
      }
    } catch (error) {
      console.error('Error sending email via Resend:', error);
    }
  }

  // Backup: Send notification to WhatsApp if Z-API is working
  if (env.MONITOR_WHATSAPP_NUMBER && alert.details?.connected !== false) {
    try {
      const message = `🚨 *Alerta de Monitoramento*\n\n${alert.subject}\n\n${alert.issue}\n\nVerifique os logs para mais detalhes.`;
      const zapiUrl = `https://api.z-api.io/instances/${env.ZAPI_INSTANCE_ID}/token/${env.ZAPI_INSTANCE_TOKEN}/send-text`;
      
      const response = await fetch(zapiUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Client-Token': env.ZAPI_CLIENT_TOKEN
        },
        body: JSON.stringify({ 
          phone: env.MONITOR_WHATSAPP_NUMBER, 
          message 
        })
      });
      
      if (response.ok) {
        console.log('WhatsApp backup alert sent successfully');
      } else {
        console.error('Failed to send WhatsApp backup alert:', await response.text());
      }
    } catch (error) {
      console.error('Error sending WhatsApp backup alert:', error);
    }
  }
}