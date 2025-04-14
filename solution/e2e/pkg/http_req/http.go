package http_req

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type HTTPReq struct {
	BaseURL string
}

func (h *HTTPReq) GET(endpoint string, respParse any) (int, error) {
	res, err := http.Get(h.BaseURL + endpoint)
	if err != nil {
		log.Println(err)
		return 404, err
	}
	if respParse != nil && res.StatusCode == 200 {
		if err := parseBody(res, respParse); err != nil {
			log.Println(err)
			return 404, err
		}
	}
	return res.StatusCode, nil
}
func (h *HTTPReq) POST(endpoint string, body any, respParse any) (int, error) {
	json, _ := json.Marshal(body)
	res, err := http.Post(h.BaseURL+endpoint, "application/json", bytes.NewBuffer(json))
	if err != nil {
		log.Println(err)
		return 0, err
	}
	if res.StatusCode != 200 && res.StatusCode != 201 && res.StatusCode != 204 {
		body, _ := io.ReadAll(res.Body)
		log.Println(string(body))
	}
	if respParse != nil {
		if err := parseBody(res, respParse); err != nil {
			return 0, err
		}
	}
	return res.StatusCode, nil
}

func parseBody(res *http.Response, respParse any) error {
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bodyBytes, respParse)
	if err != nil {
		return err
	}
	return nil
}
