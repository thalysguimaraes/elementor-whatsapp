# Elementor WhatsApp Webhook

Serviço webhook para receber dados de formulários do Elementor e enviar mensagens WhatsApp usando Z-API.

## Configuração

### 1. Clone o repositório
```bash
git clone <seu-repositorio>
cd elementor-whatsapp
```

### 2. Instale as dependências
```bash
npm install
```

### 3. Configure as variáveis de ambiente
Copie o arquivo `.env.example` para `.env` e preencha com suas credenciais:

```bash
cp .env.example .env
```

Edite o arquivo `.env` com:
- `Z_API_INSTANCE_ID`: ID da sua instância Z-API
- `Z_API_TOKEN`: Token da sua instância Z-API
- `Z_API_CLIENT_TOKEN`: Client Token da sua conta Z-API
- `WHATSAPP_NUMBER_1`: Primeiro número WhatsApp (formato: 5511999999999)
- `WHATSAPP_NUMBER_2`: Segundo número WhatsApp (formato: 5511999999999)

### 4. Execute localmente
```bash
npm start
```

O servidor estará rodando em `http://localhost:3000`

## Deploy no Render

### 1. Crie uma conta no Render
Acesse [render.com](https://render.com) e crie uma conta gratuita.

### 2. Conecte seu repositório
- Faça push do código para GitHub/GitLab
- No Render, clique em "New +" → "Web Service"
- Conecte seu repositório

### 3. Configure o serviço
O arquivo `render.yaml` já está configurado. O Render detectará automaticamente.

### 4. Configure as variáveis de ambiente
No painel do Render, adicione as seguintes variáveis:
- `Z_API_INSTANCE_ID`
- `Z_API_TOKEN`
- `Z_API_CLIENT_TOKEN`
- `WHATSAPP_NUMBER_1`
- `WHATSAPP_NUMBER_2`

### 5. Deploy
O deploy será feito automaticamente. Você receberá uma URL como:
`https://seu-servico.onrender.com`

## Configuração no Elementor

### 1. No seu formulário Elementor
- Vá para as configurações do formulário
- Adicione uma nova ação após o envio
- Escolha "Webhook"

### 2. Configure o webhook
- **URL**: `https://seu-servico.onrender.com/webhook/elementor`
- **Método**: POST
- **Formato**: JSON

### 3. Mapeamento de campos
O webhook reconhece automaticamente os seguintes campos:
- `name` ou `nome`
- `email`
- `phone` ou `telefone`
- `message` ou `mensagem`
- `subject` ou `assunto`

Outros campos serão incluídos como "Outros campos" na mensagem.

## Formato da mensagem WhatsApp

As mensagens são formatadas assim:

```
*Nova submissão de formulário*
Data/Hora: 21/07/2025 15:30:00

*Nome:* João Silva
*E-mail:* joao@email.com
*Telefone:* 11999999999
*Mensagem:* Olá, gostaria de mais informações

*Outros campos:*
empresa: Empresa XYZ
```

## Teste local

Para testar localmente, você pode usar curl:

```bash
curl -X POST http://localhost:3000/webhook/elementor \
  -H "Content-Type: application/json" \
  -d '{
    "nome": "Teste",
    "email": "teste@email.com",
    "mensagem": "Mensagem de teste"
  }'
```

## Suporte

Em caso de problemas:
1. Verifique os logs no painel do Render
2. Confirme que as variáveis de ambiente estão corretas
3. Verifique se sua instância Z-API está ativa