package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

const (
	authURL  = "https://twitter.com/i/oauth2/authorize"
	tokenURL = "https://api.twitter.com/2/oauth2/token"
	tweetURL = "https://api.twitter.com/2/tweets"
)

// Estrutura para resposta do token
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

func main() {
	// Carrega variáveis do .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar .env")
	}

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	redirectURI := os.Getenv("REDIRECT_URI")

	if clientID == "" || clientSecret == "" || redirectURI == "" {
		log.Fatal("CLIENT_ID, CLIENT_SECRET ou REDIRECT_URI ausente no .env")
	}

	// Gera state e code verifier para PKCE
	state := uuid.New().String()
	codeVerifier := generateCodeVerifier()
	codeChallenge := generateCodeChallenge(codeVerifier)

	// Monta URL de autorização
	authParams := url.Values{}
	authParams.Add("response_type", "code")
	authParams.Add("client_id", clientID)
	authParams.Add("redirect_uri", redirectURI)
	authParams.Add("scope", "tweet.write tweet.read users.read offline.access")
	authParams.Add("state", state)
	authParams.Add("code_challenge", codeChallenge)
	authParams.Add("code_challenge_method", "S256")

	authFullURL := fmt.Sprintf("%s?%s", authURL, authParams.Encode())

	fmt.Println("Abra essa URL no navegador para autorizar:")
	fmt.Println(authFullURL)

	// Tenta abrir navegador automaticamente (Linux, macOS, Windows)
	_ = exec.Command("xdg-open", authFullURL).Start()
	_ = exec.Command("open", authFullURL).Start()
	_ = exec.Command("start", authFullURL).Start()

	// Aguarda callback e captura código
	code := waitForCallback(state)
	if code == "" {
		log.Fatal("Não foi possível capturar o código de autorização")
	}

	// Troca código por tokens
	tokenRes := exchangeCodeForToken(clientID, clientSecret, codeVerifier, code, redirectURI)
	if tokenRes.AccessToken == "" {
		log.Fatal("Token não foi recebido")
	}

	// Lê todas as linhas do arquivo tweets.txt
	tweets, err := readAllLines("tweets.txt")
	if err != nil {
		log.Fatal("Erro ao ler tweets.txt:", err)
	}

	// Intervalo entre tweets (ajuste se quiser)
	interval := 30 * time.Minute

	// Posta tweets um a um com intervalo
	for i, tweet := range tweets {
		fmt.Printf("Postando tweet %d/%d: %s\n", i+1, len(tweets), tweet)
		ok := postTweet(tokenRes.AccessToken, tweet)
		if ok {
			fmt.Println("✅ Tweet postado com sucesso!")
		} else {
			fmt.Println("❌ Falha ao postar tweet.")
		}
		if i < len(tweets)-1 {
			fmt.Printf("Aguardando %v para o próximo tweet...\n", interval)
			time.Sleep(interval)
		}
	}

	fmt.Println("Todos os tweets foram postados.")
}

// Gera um código verificador PKCE aleatório
func generateCodeVerifier() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "") // pode ser mais complexo se quiser
}

// Gera o code challenge SHA256 base64url
func generateCodeChallenge(verifier string) string {
	h := sha256.New()
	h.Write([]byte(verifier))
	hash := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(hash)
}

// Espera o callback HTTP e captura o código
func waitForCallback(expectedState string) string {
	var code string
	server := &http.Server{Addr: ":8080"}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state != expectedState {
			http.Error(w, "State inválido", http.StatusForbidden)
			return
		}
		code = r.URL.Query().Get("code")
		fmt.Fprintln(w, "✅ Autorização recebida! Pode fechar esta janela.")
		go func() {
			time.Sleep(time.Second)
			server.Close()
		}()
	})

	fmt.Println("⏳ Aguardando callback em http://localhost:8080/callback ...")
	_ = server.ListenAndServe()
	return code
}

// Troca o código de autorização pelo access token (e refresh token)
func exchangeCodeForToken(clientID, clientSecret, codeVerifier, code, redirectURI string) tokenResponse {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("code_verifier", codeVerifier)
	data.Set("client_id", clientID)

	req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Erro na troca de token:", err)
	}
	defer resp.Body.Close()

	var res tokenResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		log.Fatal("Erro ao decodificar resposta do token:", err)
	}

	return res
}

// Lê todas as linhas de um arquivo txt
func readAllLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines, scanner.Err()
}

// Posta um tweet usando o token bearer
func postTweet(token, text string) bool {
	bodyMap := map[string]string{"text": text}
	bodyBytes, _ := json.Marshal(bodyMap)

	req, _ := http.NewRequest("POST", tweetURL, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Erro ao postar:", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		log.Printf("Falha ao postar tweet, status HTTP: %d\n", resp.StatusCode)
		return false
	}

	return true
}
