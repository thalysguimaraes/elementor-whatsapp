# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Elementor WhatsApp Webhook - Contexto do Projeto

## Visão Geral
Serviço webhook hospedado no Cloudflare Workers que recebe submissões de formulários do Elementor e envia mensagens WhatsApp automaticamente usando Z-API.

## Arquitetura
- **Plataforma**: Cloudflare Workers (serverless)
- **API WhatsApp**: Z-API (serviço gerenciado)
- **Banco de Dados**: Cloudflare D1 (SQLite)
- **URL do Webhook**: https://elementor-whatsapp.thalys.workers.dev/webhook/:formId
- **Linguagem**: JavaScript (Worker API)
- **Gerenciador CLI**: Node.js com Inquirer.js

## Configuração Z-API
As credenciais Z-API são armazenadas como variáveis de ambiente no `wrangler.toml`:
- `ZAPI_INSTANCE_ID`
- `ZAPI_INSTANCE_TOKEN`
- `ZAPI_CLIENT_TOKEN`

## Sistema de Números WhatsApp
Os números WhatsApp são gerenciados exclusivamente através do banco de dados D1:
- Configurados por formulário no gerenciador CLI (`npm run panel`)
- Armazenados na tabela `form_numbers` com relacionamento com `contacts`
- Não há mais números hardcoded no código ou variáveis de ambiente
- Formulários sem configuração no banco retornarão erro

## Mapeamento de Campos do Elementor
**IMPORTANTE**: O webhook suporta 3 formatos diferentes que o Elementor pode enviar!

```javascript
const fieldMapping = {
  // Formato direto JSON (campos com nomes em português)
  'nome': 'Nome',
  'empresa': 'Empresa',
  'site': 'Site',
  'telefone': 'Telefone',
  'e-mail': 'E-mail',
  'quer adiantar alguma informação? (opcional)': 'Mensagem',
  
  // Formato URL-encoded (campos com IDs genéricos)
  'name': 'Nome',
  'message': 'Site',
  'field_cef3ba0': 'Telefone',
  'field_389b567': 'E-mail',
  'field_69b2d23': 'Mensagem'
};
```

## Estrutura de Dados do Elementor
O webhook detecta automaticamente e suporta 3 formatos diferentes:

### Formato 1: JSON Aninhado
```json
{
  "form": { "id": "0ae175f", "name": "New Form" },
  "fields": {
    "nome": { "value": "Nome do cliente", ... },
    "empresa": { "value": "Nome da empresa", ... }
  }
}
```

### Formato 2: URL-Encoded Achatado
```
fields[name][value]=Nome do cliente
fields[empresa][value]=Nome da empresa
fields[field_cef3ba0][value]=(34) 99999-9999
```

### Formato 3: JSON Direto
```json
{
  "nome": "Nome do cliente",
  "empresa": "Nome da empresa",
  "site": "site.com.br",
  "telefone": "(34) 99999-9999",
  "e-mail": "email@empresa.com"
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

## Sistema de Gerenciamento de Contatos

### Visão Geral
O sistema agora inclui um gerenciador de contatos centralizado que permite reutilizar números WhatsApp entre diferentes formulários.

### Funcionalidades
- **Gerenciamento Centralizado**: Todos os contatos em um só lugar
- **Reutilização**: Use os mesmos contatos em múltiplos formulários
- **Informações Detalhadas**: Nome, empresa, cargo e notas para cada contato
- **Seleção Fácil**: Interface de checkbox para selecionar múltiplos contatos ao criar/editar formulários

### Como Usar
1. **Acessar Gerenciador**: `npm run panel` → Escolha "📞 Manage Contacts"
2. **Adicionar Contato**: Forneça nome, número WhatsApp, cargo (opcional), empresa (opcional)
3. **Criar Formulário**: Ao criar um formulário, selecione os contatos que receberão as mensagens
4. **Editar Contatos**: Atualizações nos contatos refletem automaticamente em todos os formulários

### Estrutura do Banco de Dados
```sql
contacts (
  id INTEGER PRIMARY KEY,
  phone_number TEXT UNIQUE,
  name TEXT NOT NULL,
  company TEXT,
  role TEXT,
  notes TEXT,
  created_at, updated_at
)

form_numbers (
  ... campos existentes ...,
  contact_id INTEGER REFERENCES contacts(id)
)
```

## Gerenciamento de Formulários Multi-Form

### Funcionalidades
- **Múltiplos Formulários**: Crie quantos formulários precisar, cada um com seu próprio webhook
- **Gerenciamento Dinâmico**: Todos os números são configurados via banco de dados, não há mais números hardcoded
- **Sistema de Contatos**: Contatos centralizados que podem ser reutilizados entre formulários
- **URLs Únicas**: Cada formulário tem sua URL: `/webhook/{formId}`
- **CLI Manager**: Interface interativa (`npm run panel`) para gerenciar formulários e contatos

### Comandos do Manager
```bash
# Instalar dependências do manager
npm run manager:install

# Executar o painel de gerenciamento
npm run panel

# Criar banco de dados D1
npm run db:create

# Inicializar esquema do banco
npm run db:init

# Consultar banco de dados
npm run db:query
```

## Sistema de Monitoramento Z-API

### Funcionalidades
- **Verificação Automática**: Cron job a cada 15 minutos
- **Alertas por Email**: Notificações via Resend quando WhatsApp desconecta
- **Estado Persistente (D1)**: Armazenado em tabelas do D1, escreve apenas quando o status muda
- **Histórico**: Mantém as últimas 100 mudanças de status em D1

### Configuração
As configurações de monitoramento são definidas em `wrangler.toml`:
- **Cron**: Executa a cada 15 minutos
- **Email de Alerta**: Configurado via `ALERT_EMAIL`
- **API de Email**: Resend API para notificações
- **Estado**: Armazenado no D1 (`monitoring_state`, `monitoring_history`)

Nota: Para reduzir o consumo e sair do KV, o Worker agora:
- Lê o estado atual do D1 e só grava quando o campo `connected` muda.
- Atualiza o `monitoring_history` apenas em mudanças de status (mantém últimas 100 entradas por chave).
- Removeu a dependência do KV `MONITOR_STATE`.

## Notas de Configuração

### Dependências do Projeto
- **whatsapp-cloudflare-workers**: Instalado durante testes mas NÃO é compatível com Workers
- **qrcode**: Instalado para testes, pode ser removido com `npm uninstall qrcode whatsapp-cloudflare-workers`

### Arquivos de Teste
- Vários scripts de teste foram criados: `test-webhook.sh`, `test-monitoring.sh`, etc.
- Úteis para debugging e verificação de funcionamento

## Comandos de Desenvolvimento

### Instalação
```bash
npm install
npm run setup  # Instala dependências do manager também
```

### Desenvolvimento Local
```bash
npm run dev    # Inicia servidor de desenvolvimento local
wrangler dev   # Alternativa direta
```

### Deploy para Cloudflare Workers
```bash
npm run deploy # Deploy para produção
wrangler deploy # Alternativa direta
```

### Monitoramento e Logs
```bash
npm run tail              # Ver logs em tempo real
wrangler tail             # Alternativa direta
wrangler tail --format pretty  # Logs formatados
```

### Banco de Dados
```bash
# Executar migrações (contatos + monitoramento)
wrangler d1 execute elementor-whatsapp-forms --file=./migrations/001_add_contacts_safe.sql --remote
wrangler d1 execute elementor-whatsapp-forms --file=./migrations/002_add_monitoring.sql --remote
```

### Testar webhook
```bash
npm test          # Executa script de teste
./test-webhook.sh # Teste formato URL-encoded

# Teste formato JSON direto
curl -X POST https://elementor-whatsapp.thalys.workers.dev/webhook/elementor \
  -H "Content-Type: application/json" \
  -d '{
    "nome": "Teste",
    "empresa": "Empresa Teste",
    "site": "site.com",
    "telefone": "(34) 99999-9999",
    "e-mail": "teste@email.com"
  }'
```

## Estrutura do Código

### worker.js
- Implementação principal do Cloudflare Worker
- Suporta 3 formatos de dados do Elementor
- Processa requisições de forma assíncrona
- Envia mensagens para múltiplos números em paralelo
- Logging estruturado em JSON para observabilidade
- Tratamento de erros robusto com categorização

### Endpoints
- `GET /` - Status do serviço e versão
- `POST /webhook/elementor` - Recebe dados do Elementor
- `GET /health` - Verificação de saúde e conectividade Z-API
- `OPTIONS /*` - Suporte CORS para requisições preflight

### Observabilidade
- Logs estruturados em JSON com metadados completos
- Rastreamento de erros por categoria (validação, API, rede)
- Métricas de performance por requisição
- Suporte para Cloudflare Workers Logs (dashboard)
- Configuração em `wrangler.toml` com `[observability] enabled = true`

### Tratamento de Erros
- Validação de campos obrigatórios
- Verificação de formato de dados
- Respostas HTTP apropriadas (200, 400, 500)
- Logs detalhados de falhas com contexto
- Fallback para dados brutos quando formato é desconhecido

## Repositório GitHub
- URL: https://github.com/thalysguimaraes/elementor-whatsapp
- Branch: main

## Notas Importantes

### Limitações do Elementor
1. **Metadados ausentes**: Elementor não envia data/hora por padrão a menos que outra ação (como email) inclua campos meta
2. **IDs inconsistentes**: Campos personalizados podem ter IDs diferentes entre formulários
3. **Sem retry automático**: Elementor não tenta reenviar webhooks falhos
4. **Content-Type variável**: Pode enviar como JSON ou form-encoded

### Boas Práticas
1. **Sempre validar entrada**: Verificar campos obrigatórios antes de processar
2. **Logs estruturados**: Usar JSON para facilitar queries no dashboard
3. **Status codes corretos**: 200 para sucesso, 4xx para erros do cliente, 5xx para erros do servidor
4. **Não expor segredos**: Credenciais apenas em wrangler.toml ou secrets do Cloudflare

### Solução de Problemas
1. **Mensagens não enviadas**: Verificar logs com `wrangler tail` para detalhes
2. **Formato não reconhecido**: Webhook loga dados brutos para análise
3. **Erros Z-API**: Verificar status da instância e credenciais
4. **Rate limits**: Adicionar delays se necessário entre envios

## Alternativas Testadas e Descartadas

### whatsapp-cloudflare-workers (Janeiro 2025)
- **Tentativa**: Migrar de Z-API para solução independente usando o pacote `whatsapp-cloudflare-workers`
- **Problema**: Apesar do nome, o pacote não é compatível com Cloudflare Workers devido a:
  - Uso de APIs Node.js específicas (ex: `setInterval().unref()`)
  - Dependências profundas de módulos Node.js não disponíveis em Workers
  - WebSocket implementation incompatível
- **Decisão**: Manter Z-API como solução confiável e estável
- **Branch de teste**: `feature/whatsapp-independent` (deletada após testes)

## Histórico de Mudanças
1. **v3.2.0** (29/01/2025): Melhorias na navegação do CLI e sistema de contatos
   - Sistema completo de gerenciamento de contatos centralizado
   - Navegação aprimorada com opções de cancelamento em todas as operações
   - Breadcrumb navigation mostrando hierarquia de menus
   - Fluxo contínuo sem saída automática após ações
   - Comando renomeado de `npm run manage` para `npm run panel`
   - Integração de contatos com formulários (seleção por checkbox)
   - Monitoramento Z-API com alertas via Resend API
   - D1 storage para estado/histórico de monitoramento (migração do KV)
   - Migrações de banco de dados para adicionar tabela `contacts`
2. **v3.1.0**: Sistema de gerenciamento de contatos e monitoramento Z-API
   - Tabela `contacts` para gerenciamento centralizado
   - CLI interativo para gerenciar contatos
   - Reutilização de contatos entre formulários
   - Monitoramento automático com alertas por email
3. **v3.0.0**: Sistema multi-formulários com D1 database
   - Suporte para múltiplos formulários com URLs únicas
   - Gerenciador CLI com Inquirer.js
   - Banco de dados D1 para configurações
4. **v2.0.0**: Removido suporte Node/Express, apenas Cloudflare Workers
5. **v1.x**: Migração Render → Cloudflare Workers para melhor performance
6. **Suporte multi-formato**: Detecta automaticamente 3 formatos do Elementor
7. **Observabilidade aprimorada**: Logs estruturados e monitoramento nativo
