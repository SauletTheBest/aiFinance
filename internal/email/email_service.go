package email

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// EmailService represents our Gmail API Postman service
type EmailService struct {
	clientID     string
	clientSecret string
	refreshToken string
	sender       string
}

// NewEmailService creates a new instance of the Gmail API service
func NewEmailService(clientID, clientSecret, refreshToken, sender string) *EmailService {
	return &EmailService{
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		sender:       sender,
	}
}

// TokenResponse represents the OAuth2 response from Google
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// getAccessToken requests a new temporary Access Token using our Refresh Token
func (s *EmailService) getAccessToken() (string, error) {
	data := url.Values{}
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)
	data.Set("refresh_token", s.refreshToken)
	data.Set("grant_type", "refresh_token")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return "", fmt.Errorf("failed to make OAuth token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("oauth token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// encodeRFC2047 converts any UTF-8 string (like Russian text) into safe standard email header format
func encodeRFC2047(input string) string {
	// Format: =?utf-8?B?[Base64_Encoded_Text]?=
	return fmt.Sprintf("=?utf-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(input)))
}

// sendGmail base64-encodes the MIME email and posts it to the Gmail API
func (s *EmailService) sendGmail(toEmail, subject, htmlBody string) error {
	fmt.Printf("[Gmail Service] Starting email send process to: %s\n", toEmail)

	fmt.Println("[Gmail Service] Attempting to refresh access token...")
	accessToken, err := s.getAccessToken()
	if err != nil {
		fmt.Printf("[Gmail Service] ❌ Failed to get access token: %v\n", err)
		return fmt.Errorf("token error: %w", err)
	}
	fmt.Println("[Gmail Service] 🟢 Access token refreshed successfully!")

	// 1. Construct standard MIME headers & body
	// We use the encodeRFC2047 helper here to make the Subject safe for Russian characters!
	mimeMessage := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=\"utf-8\"\r\n\r\n"+
		"%s", s.sender, toEmail, encodeRFC2047(subject), htmlBody)

	// 2. Encode to Base64 URL Encoding
	rawEncoded := base64.URLEncoding.EncodeToString([]byte(mimeMessage))

	// 3. Prepare the JSON request body
	requestPayload := map[string]string{
		"raw": rawEncoded,
	}
	jsonBody, err := json.Marshal(requestPayload)
	if err != nil {
		fmt.Printf("[Gmail Service] ❌ Failed to marshal JSON: %v\n", err)
		return fmt.Errorf("failed to marshal gmail payload: %w", err)
	}

	// 4. Build the POST request
	apiURL := "https://gmail.googleapis.com/gmail/v1/users/me/messages/send"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("[Gmail Service] ❌ Failed to build request: %v\n", err)
		return fmt.Errorf("failed to build gmail request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	fmt.Println("[Gmail Service] Sending request to Google Gmail API...")

	// 5. Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[Gmail Service] ❌ Network request failed: %v\n", err)
		return fmt.Errorf("failed to execute gmail request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("[Gmail Service] ❌ Gmail API returned error status %d: %s\n", resp.StatusCode, string(body))
		return fmt.Errorf("gmail api returned status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Println("[Gmail Service] 🎉 Success! Email successfully sent.")
	return nil
}

// SendVerificationEmail sends a 4-digit code in Russian to verify the user's email
func (s *EmailService) SendVerificationEmail(toEmail string, code string) error {
	subject := "Подтверждение почты - AI Finance App"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 400px; margin: auto; padding: 20px; border: 1px solid #e2e8f0; border-radius: 10px;">
			<h2 style="color: #1e293b;">Добро пожаловать в AI Finance! 🎉</h2>
			<p style="color: #475569;">Ваш код подтверждения:</p>
			<h1 style="letter-spacing: 8px; color: #4F46E5; background-color: #f1f5f9; padding: 10px; text-align: center; border-radius: 5px;">%s</h1>
			<p style="color: #94a3b8; font-size: 12px;">Этот код истекает через 15 минут.</p>
		</div>
	`, code)

	return s.sendGmail(toEmail, subject, body)
}

// SendPasswordResetEmail sends a 4-digit code in Russian to reset the user's password
func (s *EmailService) SendPasswordResetEmail(toEmail string, code string) error {
	subject := "Сброс пароля - AI Finance App"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 400px; margin: auto; padding: 20px; border: 1px solid #e2e8f0; border-radius: 10px;">
			<h2 style="color: #1e293b;">Запрос на сброс пароля 🔐</h2>
			<p style="color: #475569;">Ваш код для сброса пароля:</p>
			<h1 style="letter-spacing: 8px; color: #DC2626; background-color: #fef2f2; padding: 10px; text-align: center; border-radius: 5px;">%s</h1>
			<p style="color: #94a3b8; font-size: 12px;">Этот код истекает через 15 минут. Если вы не запрашивали сброс пароля, просто проигнорируйте это письмо.</p>
		</div>
	`, code)

	return s.sendGmail(toEmail, subject, body)
}
