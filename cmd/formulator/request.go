package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"

	uuid "github.com/satori/go.uuid"

	"github.com/fletaio/fleta_testnet/service/apiserver"
)

func DoRequest(c *websocket.Conn, Method string, Params []interface{}) (interface{}, error) {
	id := uuid.NewV1().String()
	req := &apiserver.JRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  Method,
		Params:  Params,
	}
	bs, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if err := c.WriteMessage(websocket.TextMessage, bs); err != nil {
		return nil, err
	}

	_, data, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}

	var res apiserver.JRPCResponse
	if json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	if res.Error != nil {
		return nil, errors.New(fmt.Sprint(res.Error))
	} else {
		return res.Result, nil
	}
}
