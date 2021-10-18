package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type CallBackMgr struct {
	HTTPSCallback     string
	HTTPSCallbackAuth string
}

type CallbackModel struct {
	ID     string `json:"id"`
	IP     string `json:"ip"`
	Action string `json:"action"`
}

func (p *CallBackMgr) callback(id, action, ip string) error {

	data := CallbackModel{
		ID:     id,
		IP:     ip,
		Action: action,
	}

	jdata, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", p.HTTPSCallback, bytes.NewBuffer(jdata))
	req.Header.Add("Authorization", p.HTTPSCallbackAuth)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}
	defer req.Body.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, error := client.Do(req)
	if error != nil {
		return error
	}
	defer resp.Body.Close()

	return err
}
