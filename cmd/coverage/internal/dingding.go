package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func AlarmJson(data *JSONData) {
	// limit alarm message count to 10
	if files, ok := data.Content["新达到100%覆盖率的文件"].([]string); ok && len(files) > 10 {
		data.Content["新达到100%覆盖率的文件"] = files[:10]
	}

	payload, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	if *dingding {
		alarm(payload)
	} else {
		fmt.Println(string(payload))
	}
}

func alarm(payload []byte) {
	//HTTP POST request
	req, err := http.NewRequest("POST", *alarmUrl, bytes.NewBuffer(payload))
	if err != nil {
		panic(err)

	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println("Response body:", string(all))

	fmt.Println("Response status:", resp.Status)
}
