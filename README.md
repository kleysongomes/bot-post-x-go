
# 🐦 bot-post-x-go

Um bot em Go que:

✅ Busca tópicos no Google via [SerpAPI](https://serpapi.com/) e gera sugestões de tweets  
✅ Salva os tweets gerados no arquivo `tweets.txt`  
✅ Lê o arquivo `tweets.txt` e posta automaticamente no X (Twitter) via API v2 com OAuth 2.0 PKCE  

---

## 📂 Estrutura do Projeto

```
bot-post-x-go/
├── main.go         // Controla a execução: busca ou postagem
├── search.go       // Faz busca no SerpAPI e cria o tweets.txt
├── post.go         // Lê tweets.txt e posta no X/Twitter com OAuth 2.0 PKCE
├── tweets.txt      // Arquivo de saída com os tweets gerados (um por linha)
├── .env            // Variáveis de ambiente (não subir isso para o GitHub)
└── README.md       // Este arquivo
```

---

## ✅ Pré-requisitos

- Go instalado: [https://go.dev/doc/install](https://go.dev/doc/install)
- Conta no [SerpAPI](https://serpapi.com/) (para buscar no Google)
- Conta de desenvolvedor no [Twitter Developer Portal (X.com)](https://developer.x.com/)

---

## 📌 Configuração do `.env`

Crie um arquivo chamado `.env` na raiz do projeto com o seguinte conteúdo:

```
# Chave da SerpAPI
SERPAPI_KEY=your_serpapi_key

# Configurações do Twitter/X (OAuth2)
CLIENT_ID=your_twitter_client_id
CLIENT_SECRET=your_twitter_client_secret
REDIRECT_URI=http://localhost:8080/callback
```

> ⚠️ **Atenção:** Nunca suba o arquivo `.env` para repositórios públicos.

---

## 🚀 Como usar

### 🔎 1. Gerar tweets a partir de uma busca no Google (via SerpAPI)

No arquivo `main.go`, **descomente** a linha:

```go
err := GenerateTweetsFileFromSearch("golang programação", serpAPIKey)
```

E **comente** a linha de postagem.

Depois execute:

```bash
go run main.go
```

Isso irá criar um arquivo chamado `tweets.txt` com os tweets sugeridos (um por linha).

---

### 🐦 2. Postar os tweets no X (Twitter)

Coloque as linhas que deseja postar dentro do arquivo `tweets.txt`.

No `main.go`, **comente** a linha de busca e **descomente** a de postagem:

```go
err := PostTweetsFromFile()
```

Depois execute:

```bash
go run main.go
```

O app abrirá o navegador para você autorizar o app no Twitter/X.  
Após autorizar, ele começará a postar os tweets com intervalo de segurança.

---

## ⏱️ Intervalo entre tweets

O intervalo entre cada postagem é de **10 minutos**, configurado dentro do `post.go`:

```go
interval := 10 * time.Minute
```

Você pode ajustar esse valor conforme desejar.

---

## ⚠️ Atenção

- Nunca execute as duas funções ao mesmo tempo no `main.go`.  
  Sempre deixe **uma ativa por vez** (geração **OU** postagem).

- Respeite os limites de uso da API do Twitter e da SerpAPI para evitar bloqueios.

- Nunca publique tokens ou credenciais no repositório.

---

## ✅ Próximos passos (opcional)

- Criar um `Makefile` para automatizar as execuções
- Criar um `go.mod` para gerenciar as dependências
- Separar os arquivos em pacotes (`/internal`, `/cmd`, etc.)

---

## 📄 Licença

MIT License.
