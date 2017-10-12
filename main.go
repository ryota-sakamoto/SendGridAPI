package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Config struct {
	Key string
}

type MailData struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Content string `json:"content"`
}

func main() {
	key := getConfig().Key
	fmt.Printf("api key: %s\n", key)

	http.HandleFunc("/api/sendmail", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			w.Header().Add("Content-Type", "application/json")

			var mailData MailData
			{
				body, _ := ioutil.ReadAll(r.Body)
				json.Unmarshal(body, &mailData)
			}

			req, _ := http.NewRequest(
				"POST",
				"https://api.sendgrid.com/v3/mail/send",
				bytes.NewBuffer(createMailRequest(mailData)),
			)
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer " + key)

			client := &http.Client{}
			resp, _ := client.Do(req)
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)

			w.Write([]byte(body))
		default:
		}
	})
	http.ListenAndServe(":8080", nil)
}

func createMailRequest(data MailData) []byte {
	request := `
	{
  "personalizations": [
    {
      "to": [
        {
          "email": "%s"
        }
      ],
      "subject": "%s"
    }
  ],
  "from": {
    "email": "%s"
  },
  "content": [
    {
      "type": "text/plain",
      "value": "%s"
    }
  ]
}
	`
	request = fmt.Sprintf(request, data.To, data.Subject, data.From, data.Content)

	return []byte(request)
}

func getConfig() Config {
	file, _ := os.Open("config.json")
	defer file.Close()

	decoder := json.NewDecoder(file)
	var config Config
	decoder.Decode(&config)

	return config
}
