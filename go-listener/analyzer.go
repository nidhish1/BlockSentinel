package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func sendToAIAnalyzer(analyzerURL string, txData map[string]interface{}) error {
	jsonData, err := json.Marshal(txData)
	if err != nil {
		return err
	}

	resp, err := http.Post(analyzerURL+"/analyze", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AI analyzer error: %s", string(body))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	log.Printf("Risk Analysis: %+v", result)

	return nil
}
