# Elementor WhatsApp Webhook - Contexto do Projeto

## Visão Geral
Serviço webhook que recebe submissões de formulários do Elementor e envia mensagens WhatsApp automaticamente usando Z-API.

## Arquitetura
- **Framework**: Node.js + Express
- **Hospedagem**: Render.com (deploy automático via GitHub)
- **API WhatsApp**: Z-API (usando instância existente do cliente)
- **URL do Webhook**: https://elementor-whatsapp.onrender.com/webhook/elementor

## Credenciais Z-API (já configuradas)
- Instance ID: ***REMOVED***
- Instance Token: ***REMOVED***
- Client Token: ***REMOVED***

## Números WhatsApp Configurados
- WHATSAPP_NUMBER_1: 5534991517110
- WHATSAPP_NUMBER_2: 5511888888888 (atualizar conforme necessário)

## Mapeamento de Campos do Elementor
**IMPORTANTE**: Os IDs dos campos não correspondem ao seu significado semântico!

```javascript
const fieldMapping = {
  'name': 'Nome',           // Campo nome real
  'email': 'Empresa',       // ATENÇÃO: Este campo recebe o nome da empresa
  'message': 'Site',        // ATENÇÃO: Este campo recebe o site
  'field_cef3ba0': 'Telefone',
  'field_389b567': 'E-mail', // Este é o e-mail real
  'field_69b2d23': 'Mensagem' // Este é a mensagem real (opcional)
};
```

## Estrutura de Dados do Elementor
O Elementor envia os dados neste formato:
```json
{
  "form": { "id": "0ae175f", "name": "New Form" },
  "fields": {
    "name": { "value": "Nome do cliente", ... },
    "email": { "value": "Nome da empresa", ... },
    "message": { "value": "Site da empresa", ... },
    "field_cef3ba0": { "value": "Telefone", ... },
    "field_389b567": { "value": "email@real.com", ... },
    "field_69b2d23": { "value": "Mensagem opcional", ... }
  },
  "meta": { ... }
}
```

## Formato da Mensagem WhatsApp
```
*Nova submissão de formulário*
Data/Hora: 21/07/2025, 12:01:43

*Nome:* João Silva
*Empresa:* Empresa XYZ
*Site:* empresa.com.br
*Telefone:* (34) 99999-9999
*E-mail:* joao@empresa.com
*Mensagem:* Texto opcional
```

## Repositório GitHub
- URL: https://github.com/thalysguimaraes/elementor-whatsapp
- Branch: main
- Deploy automático no Render ao fazer push

## Comandos Úteis

### Testar localmente
```bash
npm install
npm start

# Teste com curl
curl -X POST http://localhost:3000/webhook/elementor \
  -H "Content-Type: application/json" \
  -d '{
    "fields": {
      "name": {"value": "Teste"},
      "email": {"value": "Empresa Teste"},
      "message": {"value": "site.com"},
      "field_cef3ba0": {"value": "11999999999"},
      "field_389b567": {"value": "teste@email.com"}
    }
  }'
```

### Deploy
```bash
git add -A
git commit -m "Sua mensagem"
git push origin main
# Render faz deploy automático
```

## Logs e Debugging
- Logs disponíveis no painel do Render
- Console.log do webhook mostra estrutura completa recebida
- Verificar se as variáveis de ambiente estão configuradas no Render

## Notas Importantes
1. **Não validamos números**: Z-API permite enviar para qualquer número
2. **Sem limite de mensagens**: Z-API não tem limites de envio
3. **Formato dos números**: Usar formato completo (DDI + DDD + número)
4. **Deploy automático**: Qualquer push para main dispara novo deploy

## Próximos Passos Potenciais
- [ ] Adicionar mais números de destino (WHATSAPP_NUMBER_3, etc)
- [ ] Implementar template de mensagem customizável
- [ ] Adicionar webhook de confirmação de entrega
- [ ] Implementar fila de retry para falhas