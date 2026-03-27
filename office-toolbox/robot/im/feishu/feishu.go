package feishu

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func Send(robot string, fMsg *MsgData) {
	postData, err := json.Marshal(fMsg)
	if err != nil {
		return
	}

	request, err := http.NewRequestWithContext(context.Background(), "POST", "https://open.feishu.cn/open-apis/bot/v2/hook/"+robot, bytes.NewBuffer(postData))
	if err != nil {
		return
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpClient := &http.Client{
		Timeout: time.Second * time.Duration(30),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	fmt.Println(string(contents))
}
