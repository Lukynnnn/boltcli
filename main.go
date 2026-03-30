package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

const (
	baseURL    = "https://deliveryuser.live.boltsvc.net"
	version    = "FI.1.106"
	language   = "en-US"
	country    = "cz"
	deviceID   = "9AB772DB-DF16-41CA-9C99-3E7432CA36A7"
	deviceType = "ios"
)

type Config struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	AuthToken    string `json:"auth_token"`
	SessionID    string `json:"session_id"`
	Phone        string `json:"phone"`
	CityID       int    `json:"city_id"`
}

var cfg Config
var cfgPath string

func configFilePath() string {
	if cfgPath != "" {
		return cfgPath
	}
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "boltcli", "config.json")
}

func loadConfig() {
	data, err := os.ReadFile(configFilePath())
	if err != nil {
		return
	}
	json.Unmarshal(data, &cfg)
}

func saveConfig() error {
	p := configFilePath()
	os.MkdirAll(filepath.Dir(p), 0700)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}

func commonParams() url.Values {
	q := url.Values{}
	q.Set("version", version)
	q.Set("language", language)
	q.Set("country", country)
	q.Set("session_id", cfg.SessionID)
	q.Set("deviceId", deviceID)
	q.Set("deviceType", deviceType)
	q.Set("device_name", "iPhone15,2")
	q.Set("device_os_version", "26.3.1")
	q.Set("distinct_id", "delivery-179310078")
	return q
}

func doRequest(method, path string, queryParams url.Values, body any) (map[string]any, error) {
	fullURL := baseURL + path
	if queryParams != nil {
		fullURL += "?" + queryParams.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Bolt/FI.1.106 iOS/26.3.1")
	if cfg.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

var rootCmd = &cobra.Command{
	Use:   "boltcli",
	Short: "Bolt Food CLI",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Sign in via phone number + SMS OTP",
	RunE: func(cmd *cobra.Command, args []string) error {
		phone, _ := cmd.Flags().GetString("phone")
		if phone == "" {
			return fmt.Errorf("provide --phone +XXXXXXXXXXX")
		}

		cfg.SessionID = deviceID + "eater" + fmt.Sprintf("%d", time.Now().Unix())

		q := commonParams()
		body := map[string]any{
			"type":             "phone",
			"phone_uuid":       "b8a8e0fc-ee3e-418c-adf5-e5df76096baa",
			"phone_number":     phone,
			"method":           "sms",
			"last_known_state": map[string]any{},
		}

		fmt.Println("Sending OTP to", phone, "...")
		resp, err := doRequest("POST", "/profile/verification/start", q, body)
		if err != nil {
			return err
		}
		if code, ok := resp["code"].(float64); ok && code != 0 {
			return fmt.Errorf("error: %v", resp["message"])
		}
		fmt.Print("SMS sent. Enter OTP code: ")

		var otp string
		fmt.Scan(&otp)

		confirmBody := map[string]any{
			"phone_number":     phone,
			"phone_uuid":       "b8a8e0fc-ee3e-418c-adf5-e5df76096baa",
			"code":             otp,
			"type":             "phone",
			"last_known_state": map[string]any{},
		}

		resp, err = doRequest("POST", "/profile/verification/confirm", q, confirmBody)
		if err != nil {
			return err
		}
		if code, ok := resp["code"].(float64); ok && code != 0 {
			return fmt.Errorf("error: %v", resp["message"])
		}

		data := resp["data"].(map[string]any)
		auth := data["auth"].(map[string]any)

		cfg.Phone = phone
		cfg.AuthToken = auth["auth_token"].(string)
		cfg.AccessToken = auth["access_token"].(string)
		cfg.RefreshToken = auth["refresh_token"].(string)
		if cityID, ok := auth["city_id"].(float64); ok {
			cfg.CityID = int(cityID)
		} else {
			cfg.CityID = 459
		}

		if err := saveConfig(); err != nil {
			return err
		}

		firstName := auth["first_name"].(string)
		lastName := auth["last_name"].(string)
		fmt.Printf("Logged in as %s %s\n", firstName, lastName)
		return nil
	},
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "List past orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.AccessToken == "" {
			return fmt.Errorf("not logged in, run: boltcli login")
		}
		limit, _ := cmd.Flags().GetInt("limit")

		q := commonParams()
		q.Set("page_size", fmt.Sprintf("%d", limit))

		resp, err := doRequest("GET", "/deliveryClient/getOrderHistory", q, nil)
		if err != nil {
			return err
		}

		data, ok := resp["data"].(map[string]any)
		if !ok {
			return fmt.Errorf("unexpected response")
		}
		orders, ok := data["orders"].([]any)
		if !ok || len(orders) == 0 {
			fmt.Println("No orders found.")
			return nil
		}

		for _, o := range orders {
			order := o.(map[string]any)
			orderID := int64(order["order_id"].(float64))
			state := order["order_state"].(string)
			ts := int64(order["order_created_timestamp"].(float64))
			t := time.Unix(ts, 0)
			providerName := order["provider_name"].(map[string]any)["value"].(string)
			fmt.Printf("%d\t%s\t%s\t%s\n", orderID, providerName, state, t.Format("2006-01-02 15:04"))
		}
		return nil
	},
}

var ordersCmd = &cobra.Command{
	Use:   "orders",
	Short: "List active orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.AccessToken == "" {
			return fmt.Errorf("not logged in, run: boltcli login")
		}

		q := commonParams()
		q.Set("reason", "default-polling-not-active")

		resp, err := doRequest("GET", "/deliveryClient/getActiveOrders", q, nil)
		if err != nil {
			return err
		}

		data, ok := resp["data"].(map[string]any)
		if !ok {
			return fmt.Errorf("unexpected response")
		}
		orders, ok := data["orders"].([]any)
		if !ok || len(orders) == 0 {
			fmt.Println("No active orders.")
			return nil
		}

		for _, o := range orders {
			order := o.(map[string]any)
			orderID := int64(order["order_id"].(float64))
			state := order["order_state"].(string)
			providerName := order["provider_name"].(map[string]any)["value"].(string)
			fmt.Printf("%d\t%s\t%s\n", orderID, providerName, state)
		}
		return nil
	},
}

var orderCmd = &cobra.Command{
	Use:   "order <order_id>",
	Short: "Show details for a single order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.AccessToken == "" {
			return fmt.Errorf("not logged in, run: boltcli login")
		}

		var orderID int64
		fmt.Sscanf(args[0], "%d", &orderID)

		q := commonParams()
		body := map[string]any{
			"order_id": orderID,
			"city_id":  cfg.CityID,
		}

		resp, err := doRequest("POST", "/deliveryClient/v2/getOrderDetails", q, body)
		if err != nil {
			return err
		}

		data, ok := resp["data"].(map[string]any)
		if !ok {
			return fmt.Errorf("unexpected response: %v", resp)
		}

		providerName := data["provider_name"].(map[string]any)["value"].(string)
		state := data["order_state"].(string)
		ref := data["order_reference_id"].(string)

		fmt.Printf("Order:      %s\n", ref)
		fmt.Printf("Restaurant: %s\n", providerName)
		fmt.Printf("Status:     %s\n", state)

		if baskets, ok := data["user_baskets"].([]any); ok && len(baskets) > 0 {
			basket := baskets[0].(map[string]any)
			if items, ok := basket["items"].([]any); ok {
				fmt.Println("Items:")
				for _, item := range items {
					i := item.(map[string]any)
					name := i["name"].(map[string]any)["value"].(string)
					qty := int(i["amount"].(float64))
					price := i["unit_item_price"].(map[string]any)["original_price"].(map[string]any)["value"].(float64)
					fmt.Printf("  %dx %s (%.0f)\n", qty, name, price)
				}
			}
		}

		if total, ok := data["total_price"].(map[string]any); ok {
			if v, ok := total["value"].(float64); ok {
				fmt.Printf("Total:      %.0f\n", v)
			}
		}

		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg = Config{}
		return saveConfig()
	},
}

func init() {
	loginCmd.Flags().String("phone", "", "Phone number (+XXXXXXXXXXX)")
	historyCmd.Flags().Int("limit", 20, "Number of orders to show")

	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "config file path")
	rootCmd.AddCommand(loginCmd, historyCmd, ordersCmd, orderCmd, logoutCmd)
}

func main() {
	loadConfig()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
