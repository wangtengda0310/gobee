package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
)

var token = flag.String("token", "c1be2fb9af823da8a1d5d5b7a616a6cad3646b39b86ce0c170fbd387936689a7", "机器人token")
var secret = flag.String("secret", "SECbdc49e29fba225a5c4ba50e4786e81664e5058d88765c8fba3ed54779a04d6b1", "机器人secret")

var dingding = flag.Bool("dingding", false, "是否发送钉钉消息")

func alarmJson(data *JSONData) {
	// limit alarm message count to 10
	if len(data.Content.New100PercentFiles) > 10 {
		data.Content.New100PercentFiles = data.Content.New100PercentFiles[:10]
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

var alarmUrl = flag.String("alarmUrl", "http://alarm.iwgame.com/alarm/dingtalk/sendTemplate", "报警url")

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
