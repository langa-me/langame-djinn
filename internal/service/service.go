package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/langa-me/langame-djinn/internal/config"
	"github.com/langa-me/langame-djinn/internal/djinn"
)

var (
	MODEL = "siebert/sentiment-roberta-large-english"
	API_URL = fmt.Sprintf("https://api-inference.huggingface.co/models/%s", MODEL)
)

func Query(payload string) (*djinn.MagnificationResponse_Sentiment, error) {
	values := map[string]string{"inputs": payload}
	json_data, err := json.Marshal(values)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", API_URL, bytes.NewReader(json_data))
	if err != nil {
		return nil, err
	}
	req.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{fmt.Sprintf("Bearer %s", config.Config.HuggingfaceKey)},
	}
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", resp.Status)
	}

	defer resp.Body.Close()
	var hfResp [][]huggingfaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&hfResp); err != nil {
		return nil, err
	}

	fmt.Printf("log %v", hfResp[0][0].Label)
	// Find label == NEGATIVE and label == POSITIVE values
	pos, neg := 0.0, 0.0
	for _, hf := range hfResp[0] {
		if strings.Contains(hf.Label, "NEGATIVE") {
			neg = hf.Score
		}
		if strings.Contains(hf.Label, "POSITIVE") {
			pos = hf.Score
		}
	}
	if pos == 0.0 || neg == 0.0 {
		return nil, fmt.Errorf("could not find POSITIVE or NEGATIVE label")
	}

	return &djinn.MagnificationResponse_Sentiment{
		Positive: pos,
		Negative: neg,
	}, nil
}

type huggingfaceResponse struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}
