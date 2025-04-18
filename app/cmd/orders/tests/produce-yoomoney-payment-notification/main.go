package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/google/uuid"
)

func main() {
	env := cfg.AssertEnv(setup.EnvKeyYoomoneyNotificationSecret)

	apiUrl := "http://localhost:8080/api/v1/order/process-payment/yoomoney"

	notificationSecret := env[setup.EnvKeyYoomoneyNotificationSecret]

	orderId := "foobar"
	var amount float64 = 299.45

	notificationType := "card-incoming"
	operationId := uuid.NewString()
	amountStr := strconv.FormatFloat(amount, 'f', 3, 64)
	currency := "643"
	datetime := time.Now().Format(time.RFC3339)
	sender := "12345"
	codepro := "false"

	label := fmt.Sprintf("%s", orderId)

	integrityCheckString := strings.Join([]string{
		notificationType,
		operationId,
		amountStr,
		currency,
		datetime,
		sender,
		codepro,
		notificationSecret,
		label,
	}, "&")
	h := sha1.New()
	h.Write([]byte(integrityCheckString))
	hashBytes := h.Sum(nil)

	sha1Hash := hex.EncodeToString(hashBytes)

	formData := url.Values{}
	formData.Set("notification_type", notificationType)
	formData.Set("operation_id", operationId)
	formData.Set("amount", amountStr)
	formData.Set("currency", currency)
	formData.Set("datetime", datetime)
	formData.Set("sender", sender)
	formData.Set("codepro", codepro)
	formData.Set("label", label)

	formData.Set("sha1_hash", sha1Hash)

	_ = notificationSecret

	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request:", err)
	}

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatal("Error reading response:", err)
	}

	fmt.Print("response data: ")
	prettyData, err := json.MarshalIndent(&data, "", "  ")
	if err != nil {
		log.Fatal("Error encoding response to stdout: ", err)
	}

	io.Copy(os.Stdout, bytes.NewReader(prettyData))
}
