package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"flag"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"io/ioutil"
)

type Config struct {
	Key string `json:"key"`
}

type MailData struct {
	From    string `json:"from" binding:"required"`
	To      string `json:"to" binding:"required"`
	Subject string `json:"subject" binding:"required"`
	Content string `json:"content" binding:"required"`
}

var (
	port_opt = flag.Int("p", 8080, "Run Application Port")
)

func main() {
	config := getConfig()
	fmt.Printf("api key: %s\n", config.Key)

	flag.Parse()

	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/sendmail", sendMail)
	}

	r.Run(fmt.Sprintf(":%d", *port_opt))
}

func sendMail(c *gin.Context) {
	var mailData MailData
	if c.BindJSON(&mailData) == nil {
		resp, e := sendRequest(
			"POST",
			"https://api.sendgrid.com/v3/mail/send",
			createMailRequest(mailData),
		)
		if e != nil {
			fmt.Printf("%s\n", e)
			c.Status(500)
		} else {
			if (string(resp) != "") {
				fmt.Printf("%s\n", string(resp))
				c.Status(400)
			} else {
				c.JSON(200, gin.H{
					"Success": true,
				})
			}
		}
	}
}

func sendRequest(method string, url string, body []byte) ([]byte, error) {
	key := getConfig().Key

	req, e := http.NewRequest(
		method,
		url,
		bytes.NewBuffer(body),
	)
	if e != nil {
		return nil, e
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer " + key)

	client := &http.Client{}
	resp, e := client.Do(req)
	defer resp.Body.Close()
	if e != nil {
		return nil, e
	}

	resp_body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, e
	}

	return resp_body, nil
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
