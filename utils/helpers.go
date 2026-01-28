package utils

import (
	"crypto/rand"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func GenerateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func ValidateEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func SanitizeHTML(input string) string {
	// Remove potentially dangerous tags
	re := regexp.MustCompile(`<script.*?>.*?</script>`)
	input = re.ReplaceAllString(input, "")

	re = regexp.MustCompile(`on\w+=".*?"`)
	input = re.ReplaceAllString(input, "")

	re = regexp.MustCompile(`javascript:`)
	input = re.ReplaceAllString(input, "")

	return input
}

func ExtractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func MaskEmail(email string) string {
	if len(email) < 5 {
		return email
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	username := parts[0]
	domain := parts[1]

	if len(username) > 2 {
		maskedUsername := string(username[0]) + "***" + string(username[len(username)-1])
		return maskedUsername + "@" + domain
	}

	return username + "@" + domain
}
func GetBaseURL(host, port string) string {

	// 1️⃣ Explicit override (recommended)
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		return baseURL
	}

	// 2️⃣ Vercel
	if vercelURL := os.Getenv("VERCEL_URL"); vercelURL != "" {
		return "https://" + vercelURL
	}

	// 3️⃣ Render
	if renderURL := os.Getenv("RENDER_EXTERNAL_URL"); renderURL != "" {
		return renderURL
	}

	// 4️⃣ Fallback → local/dev
	return fmt.Sprintf("http://%s:%s", host, port)
}
