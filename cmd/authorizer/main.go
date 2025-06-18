// Package main provides a command-line tool for obtaining Yandex Music API access token.
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Kud1nov/yamusic-dl/internal/logger"
	"github.com/google/uuid"
)

// API constants
const (
	// Client configuration
	ClientID    = "97fe03033fa34407ac9bcf91d5afed5b"
	UserAgent   = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) YandexMusic/5.56.0 Chrome/128.0.6613.162 Electron/32.1.2 Safari/537.36"
	RedirectURI = "music-application://desktop/oauth?redirectUri=&language=ru"
	Origin      = "music_desktop"
	Language    = "ru"

	// API endpoints
	URLPassportAuth    = "https://passport.yandex.ru/auth"
	URLAuthStart       = "https://passport.yandex.ru/registration-validations/auth/multi_step/start"
	URLCommitPassword  = "https://passport.yandex.ru/registration-validations/auth/multi_step/commit_password"
	URLChallengeSubmit = "https://passport.yandex.ru/registration-validations/auth/challenge/submit"
	URLSendPush        = "https://passport.yandex.ru/registration-validations/auth/challenge/send_push"
	URLChallengeCommit = "https://passport.yandex.ru/registration-validations/auth/challenge/commit"
)

// AuthSession holds authentication session data
type AuthSession struct {
	client               *http.Client
	csrfToken            string
	trackID              string
	availableAuthMethods []string
	state                string
	log                  *logger.Logger
}

// Response models for API parsing
type AuthStartResponse struct {
	Status              string   `json:"status"`
	AuthMethods         []string `json:"auth_methods"`
	CsrfToken           string   `json:"csrf_token"`
	TrackID             string   `json:"track_id"`
	PreferredAuthMethod string   `json:"preferred_auth_method"`
	SecurePhoneNumber   struct {
		MaskedE164          string `json:"masked_e164"`
		MaskedInternational string `json:"masked_international"`
	} `json:"secure_phone_number"`
}

type AuthPasswordResponse struct {
	Status      string `json:"status"`
	State       string `json:"state"`
	RedirectURL string `json:"redirect_url"`
}

type ChallengeResponse struct {
	Status    string `json:"status"`
	Challenge struct {
		ChallengeType string   `json:"challengeType"`
		Hint          []string `json:"hint,omitempty"`
		PhoneHint     string   `json:"phone_hint,omitempty"`
	} `json:"challenge"`
}

type ChallengePushResponse struct {
	Status       string `json:"status"`
	IsPushSilent bool   `json:"is_push_silent"`
}

type ChallengeCommitResponse struct {
	Status  string `json:"status"`
	Retpath string `json:"retpath"`
}

// NewAuthSession creates and initializes a new authentication session
func NewAuthSession(log *logger.Logger) (*AuthSession, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	session := &AuthSession{
		client: &http.Client{
			Jar: jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Don't follow redirects automatically
				return http.ErrUseLastResponse
			},
		},
		state: generateOAuthState(),
		log:   log,
	}

	return session, nil
}

// GetRetpathURL returns the full OAuth redirect URL
func (s *AuthSession) GetRetpathURL() string {
	return fmt.Sprintf("https://oauth.yandex.ru/authorize?response_type=token&display=popup&scope=music%%3Acontent&scope=music%%3Aread&scope=music%%3Awrite&client_id=%s&redirect_uri=%s&state=%s&origin=%s&language=%s",
		ClientID, url.QueryEscape(RedirectURI), s.state, Origin, Language)
}

// GetStandardHeaders returns common HTTP headers for requests
func (s *AuthSession) GetStandardHeaders() map[string]string {
	return map[string]string{
		"Content-Type":     "application/x-www-form-urlencoded; charset=UTF-8",
		"User-Agent":       UserAgent,
		"Accept":           "application/json, text/javascript, */*; q=0.01",
		"Accept-Language":  "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3",
		"Accept-Encoding":  "gzip, deflate, br",
		"X-Requested-With": "XMLHttpRequest",
		"Connection":       "keep-alive",
		"Origin":           "https://passport.yandex.ru",
		"Referer":          "https://passport.yandex.ru/",
		"Sec-Fetch-Dest":   "empty",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Site":   "same-origin",
	}
}

// AddStandardHeaders adds common headers to a request
func (s *AuthSession) AddStandardHeaders(req *http.Request) {
	headers := s.GetStandardHeaders()
	for key, value := range headers {
		req.Header.Add(key, value)
	}
}

// GetInitialCSRFToken requests the initial CSRF token needed for authentication
func (s *AuthSession) GetInitialCSRFToken() error {
	authURL := URLPassportAuth + "?noreturn=1&origin=" + Origin + "&language=" + Language + "&retpath=" + url.QueryEscape(s.GetRetpathURL())

	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("User-Agent", UserAgent)
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Add("Accept-Language", "ru")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Sec-Fetch-Dest", "document")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-Site", "none")
	req.Header.Add("Sec-Fetch-User", "?1")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.log.Errorf("Error closing response body: %v", err)
		}
	}()

	// Check for captcha
	if resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")
		if strings.Contains(location, "showcaptcha") {
			s.log.Info("\n⚠️  CAPTCHA required!")
			s.log.Info("Open the following link in your browser, complete the CAPTCHA and finish the authorization process:")
			s.log.Infof("\n%s\n", location)
			s.log.Info("After passing the CAPTCHA, open Developer Tools (F12), find the access_token in the page source or run in the browser console:")
			s.log.Info("console.log((document.documentElement.innerHTML.match(/access_token=([a-zA-Z0-9_-]+)/) || [])[1] || 'Not found');")
			s.log.Fatal("CAPTCHA required. See instructions above.")
		}
	}

	// Handle gzip encoding
	var bodyReader io.ReadCloser
	if resp.Header.Get("Content-Encoding") == "gzip" {
		bodyReader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("gzip decoding error: %w", err)
		}
		defer func() {
			if err := bodyReader.Close(); err != nil {
				s.log.Errorf("Error closing gzip reader: %v", err)
			}
		}()
	} else {
		bodyReader = resp.Body
	}

	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	bodyStr := string(body)

	// Search patterns for CSRF token
	patterns := []struct {
		name   string
		search string
		offset int
	}{
		{"input-field", `name="csrf_token" value="`, 24},
		{"json-field", `"csrf_token":"`, 14},
		{"data-attribute", `data-csrf="`, 11},
		{"redux-store", `"csrf":"`, 8},
		{"common-store", `"common":{"csrf":"`, 16},
		{"csrf-script", `csrf_token = "`, 13},
		{"csrf-var", `var csrf_token = "`, 17},
		{"csrf-const", `const csrf_token = "`, 19},
		{"csrf-direct", `csrf_token: "`, 13},
	}

	// Try to find CSRF token using various patterns
	for _, pattern := range patterns {
		csrfStart := strings.Index(bodyStr, pattern.search)
		if csrfStart != -1 {
			csrfStart += pattern.offset
			csrfEnd := strings.Index(bodyStr[csrfStart:], "\"")
			if csrfEnd != -1 {
				s.csrfToken = bodyStr[csrfStart : csrfStart+csrfEnd]
				if len(s.csrfToken) > 0 {
					s.log.Info("✓ CSRF token found")
					return nil
				}
			}
		}
	}

	// Try regex patterns if standard patterns fail
	regexPatterns := []struct {
		name  string
		regex *regexp.Regexp
	}{
		{"csrf-standard", regexp.MustCompile(`csrf_token[=:]["']([a-zA-Z0-9:._-]+)["']`)},
		{"hexadecimal-with-colon", regexp.MustCompile(`[a-f0-9]{32}:[0-9]+`)},
		{"form-input", regexp.MustCompile(`<input[^>]*name=["']csrf_token["'][^>]*value=["']([^"']+)["']`)},
	}

	for _, pattern := range regexPatterns {
		matches := pattern.regex.FindStringSubmatch(bodyStr)
		if len(matches) > 1 {
			s.csrfToken = matches[1]
			s.log.Info("✓ CSRF token found via regex")
			return nil
		} else if len(matches) == 1 {
			s.csrfToken = matches[0]
			s.log.Info("✓ CSRF token found via regex (full match)")
			return nil
		}
	}

	// Manual token entry if automatic methods fail
	reader := bufio.NewReader(os.Stdin)
	s.log.Info("❌ CSRF token not found automatically")
	s.log.Info("Please enter the CSRF token manually (or press Enter to search for potential tokens): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input != "" {
		s.csrfToken = input
		return nil
	}

	// Find potential tokens
	potentialTokens := regexp.MustCompile(`[a-f0-9]{32}[.:][a-f0-9]+`).FindAllString(bodyStr, -1)
	if len(potentialTokens) > 0 {
		s.log.Info("Potential tokens found:")
		for i, token := range potentialTokens[:minInt(5, len(potentialTokens))] {
			s.log.Infof("  %d. %s\n", i+1, token)
		}

		s.log.Info("Select a token number (or 0 to skip): ")
		var choice int
		if _, err := fmt.Scanf("%d", &choice); err != nil {
			s.log.Errorf("Error reading choice: %v", err)
			return fmt.Errorf("failed to read choice")
		}
		if choice > 0 && choice <= len(potentialTokens) {
			s.csrfToken = potentialTokens[choice-1]
			return nil
		}
	}

	return fmt.Errorf("failed to find CSRF token")
}

// StartAuth initiates the authentication process
func (s *AuthSession) StartAuth(login string) (*AuthStartResponse, error) {
	data := url.Values{}
	data.Set("csrf_token", s.csrfToken)
	data.Set("login", login)
	data.Set("process_uuid", uuid.NewString())
	data.Set("retpath", s.GetRetpathURL())
	data.Set("origin", Origin)
	data.Set("check_for_xtokens_for_pictures", "1")
	data.Set("force_check_for_protocols", "true")

	req, err := http.NewRequest("POST", URLAuthStart, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	s.AddStandardHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.log.Errorf("Error closing response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle empty response
	if len(body) == 0 || string(body) == "{}" {
		return nil, fmt.Errorf("empty response received, possible CSRF token issue")
	}

	var authResponse AuthStartResponse
	err = json.Unmarshal(body, &authResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Update track ID for subsequent requests
	if authResponse.TrackID != "" {
		s.trackID = authResponse.TrackID
	}

	// Store available auth methods
	if len(authResponse.AuthMethods) > 0 {
		s.availableAuthMethods = authResponse.AuthMethods
	}

	return &authResponse, nil
}

// SubmitPassword sends the password for authentication
func (s *AuthSession) SubmitPassword(password string) (*AuthPasswordResponse, error) {
	data := url.Values{}
	data.Set("csrf_token", s.csrfToken)
	data.Set("track_id", s.trackID)
	data.Set("password", password)
	data.Set("retpath", s.GetRetpathURL())
	data.Set("lang", Language)

	req, err := http.NewRequest("POST", URLCommitPassword, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	s.AddStandardHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.log.Errorf("Error closing response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle empty response
	if len(body) == 0 || string(body) == "{}" {
		return nil, fmt.Errorf("empty response received, possible password error")
	}

	var authResponse AuthPasswordResponse
	err = json.Unmarshal(body, &authResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &authResponse, nil
}

// SubmitChallenge requests 2FA challenge information
func (s *AuthSession) SubmitChallenge() (*ChallengeResponse, error) {
	data := url.Values{}
	data.Set("csrf_token", s.csrfToken)
	data.Set("track_id", s.trackID)

	req, err := http.NewRequest("POST", URLChallengeSubmit, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	s.AddStandardHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.log.Errorf("Error closing response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var challengeResponse ChallengeResponse
	err = json.Unmarshal(body, &challengeResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &challengeResponse, nil
}

// SendPush sends a push notification for 2FA
func (s *AuthSession) SendPush() (*ChallengePushResponse, error) {
	data := url.Values{}
	data.Set("csrf_token", s.csrfToken)
	data.Set("track_id", s.trackID)

	req, err := http.NewRequest("POST", URLSendPush, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	s.AddStandardHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.log.Errorf("Error closing response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var pushResponse ChallengePushResponse
	err = json.Unmarshal(body, &pushResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &pushResponse, nil
}

// CommitChallenge validates the 2FA code
func (s *AuthSession) CommitChallenge(code string) (*ChallengeCommitResponse, error) {
	data := url.Values{}
	data.Set("csrf_token", s.csrfToken)
	data.Set("track_id", s.trackID)
	data.Set("challenge", "push_2fa")
	data.Set("answer", code)

	req, err := http.NewRequest("POST", URLChallengeCommit, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	s.AddStandardHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.log.Errorf("Error closing response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var commitResponse ChallengeCommitResponse
	err = json.Unmarshal(body, &commitResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &commitResponse, nil
}

// GetToken follows redirects to obtain the access token
func (s *AuthSession) GetToken(retpath string) (string, error) {
	req, err := http.NewRequest("GET", retpath, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("User-Agent", UserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.log.Errorf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")

		// Check if URL contains the token
		if strings.Contains(location, "access_token=") {
			u, err := url.Parse(location)
			if err != nil {
				return "", fmt.Errorf("failed to parse URL: %w", err)
			}

			fragment := u.Fragment
			values, err := url.ParseQuery(fragment)
			if err != nil {
				return "", fmt.Errorf("failed to parse fragment: %w", err)
			}

			return values.Get("access_token"), nil
		}

		// Follow redirect if token not found
		return s.GetToken(location)
	}

	// Check page content for token if not found in redirect
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	bodyStr := string(body)
	if strings.Contains(bodyStr, "access_token=") {
		startIdx := strings.Index(bodyStr, "access_token=")
		endIdx := strings.Index(bodyStr[startIdx:], "&")
		if endIdx == -1 {
			endIdx = len(bodyStr) - startIdx
		}

		token := bodyStr[startIdx+13 : startIdx+endIdx]
		return token, nil
	}

	return "", fmt.Errorf("access token not found")
}

// Helper functions

// generateOAuthState creates a random state for OAuth
func generateOAuthState() string {
	randomNum := rand.New(rand.NewSource(time.Now().UnixNano())).Float64() * 1e11
	return fmt.Sprintf("%x", int64(randomNum))
}

// promptForLogin asks user to input their Yandex login
func promptForLogin(log *logger.Logger) string {
	reader := bufio.NewReader(os.Stdin)
	log.Info("Enter Yandex login: ")
	login, _ := reader.ReadString('\n')
	return strings.TrimSpace(login)
}

// promptForPassword asks user to input their password
func promptForPassword(log *logger.Logger) string {
	reader := bufio.NewReader(os.Stdin)
	log.Info("Enter password: ")
	password, _ := reader.ReadString('\n')
	return strings.TrimSpace(password)
}

// promptForCode asks user to input 2FA code
func promptForCode(log *logger.Logger) string {
	reader := bufio.NewReader(os.Stdin)
	log.Info("Enter code from push notification: ")
	code, _ := reader.ReadString('\n')
	return strings.TrimSpace(code)
}

// minInt returns the smaller of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	verbose := flag.Bool("verbose", false, "Output debug messages")
	flag.Parse()
	log := logger.New(*verbose)

	log.Info("Yandex Music Authorization Tool")
	log.Info("==============================")

	// Create authentication session
	session, err := NewAuthSession(log)
	if err != nil {
		log.Fatal("Error initializing session: %v", err)
	}

	// Get CSRF token
	log.Info("Requesting CSRF token...")
	err = session.GetInitialCSRFToken()
	if err != nil {
		log.Fatal("Error getting CSRF token: %v", err)
	}

	// Get login from user
	login := promptForLogin(log)

	// Start authentication
	log.Info("Starting authentication...")
	authStartResp, err := session.StartAuth(login)
	if err != nil {
		log.Fatal("Authentication start error: %v", err)
	}

	if authStartResp.Status != "ok" {
		log.Fatal("Failed to start authentication. Check your login.")
	}

	// Get password from user
	password := promptForPassword(log)

	// Submit password
	log.Info("Submitting password...")
	authPassResp, err := session.SubmitPassword(password)
	if err != nil {
		log.Fatal("Password submission error: %v", err)
	}

	if authPassResp.Status != "ok" {
		log.Fatal("Incorrect password or authentication error.")
	}

	// Check if 2FA is required
	if authPassResp.State == "auth_challenge" {
		// Get 2FA type
		log.Info("Two-factor authentication required.")
		challengeResp, err := session.SubmitChallenge()
		if err != nil {
			log.Fatal("Challenge request error: %v", err)
		}

		if challengeResp.Status != "ok" {
			log.Fatal("Error requesting two-factor authentication.")
		}

		// Handle push notification 2FA
		if challengeResp.Challenge.ChallengeType == "push_2fa" {
			log.Info("Push notification confirmation required.")

			// Send push notification
			pushResp, err := session.SendPush()
			if err != nil {
				log.Fatal("Push notification error: %v", err)
			}

			if pushResp.Status != "ok" {
				log.Fatal("Error sending push notification.")
			}

			log.Info("Push notification sent to your device.")

			// Get confirmation code
			code := promptForCode(log)

			// Submit code
			log.Info("Submitting confirmation code...")
			commitResp, err := session.CommitChallenge(code)
			if err != nil {
				log.Fatal("Code submission error: %v", err)
			}

			if commitResp.Status != "ok" {
				log.Fatal("Incorrect code or authentication error.")
			}

			// Get access token
			log.Info("Getting access token...")
			token, err := session.GetToken(commitResp.Retpath)
			if err != nil {
				log.Fatal("Error getting access token: %v", err)
			}

			log.Info("\nAuthentication successful!")
			log.Info("==========================")
			log.Infof("Access Token: %s\n", token)
		} else {
			log.Fatalf("Unsupported two-factor authentication type: %s\n", challengeResp.Challenge.ChallengeType)
		}
	} else {
		// No 2FA required, get token directly
		log.Info("Getting access token...")
		token, err := session.GetToken(authPassResp.RedirectURL)
		if err != nil {
			log.Fatal("Error getting access token: %v", err)
		}

		log.Info("\nAuthentication successful!")
		log.Info("==========================")
		log.Infof("Access Token: %s\n", token)
	}
}
