package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

const AUTHORIZE = "%s/v1/authorize?client_id=%s&response_type=code&response_mode=query&" +
	"scope=openid+profile+email&redirect_uri=%s&state=%s&" +
	"nonce=85582b03-f422-4742-b515-eedefe373ae2&sessionToken=%s"

var config Configuration
var states map[string]int64

func main() {
	loadConfig(&config)
	states = make(map[string]int64)
	http.HandleFunc("/authorize", handleAuthorize)
	http.HandleFunc("/code", handleCode)
	fmt.Println("starting...")

	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("Server startup failed: %s\n", err)
	}
}

/*
 * Handle call from client with session token passed
 * Generate state/nonce and redirect to Auth Server /authorize
 */
func handleAuthorize(res http.ResponseWriter, req *http.Request) {
	sessionToken := req.URL.Query().Get("session_token")
	if len(sessionToken) == 0 {
		fmt.Printf("Error no Session Token passed to /authorize")
		s := buildReply(`'{"error": "No SessionToken Present"}'`)
		res.WriteHeader(200)
		res.Write([]byte(s))
	} else {
		uuid := uuid.NewString()
		states[uuid] = time.Now().UnixNano()
		url := fmt.Sprintf(AUTHORIZE, config.ISSUER, config.CLIENT_ID, config.REDIRECT_URI, uuid, sessionToken)
		fmt.Printf("\nauthorize URL: %s\n", url)
		http.Redirect(res, req, url, http.StatusMovedPermanently)
	}
}

/*
 * Handle redirect from Auth Server /authorize call
 * exchange code for tokens
 */
func handleCode(res http.ResponseWriter, req *http.Request) {
	code := req.URL.Query().Get("code")
	state := req.URL.Query().Get("state")
	errorDescription := req.URL.Query().Get("error_description")
	err := req.URL.Query().Get("error")

	if len(code) == 0 {
		fmt.Println("Error returned from /authorize")
		var s string
		if len(errorDescription) != 0 {
			s = buildReply(fmt.Sprintf(`'{"error": "%s"}'`, errorDescription))
		} else if len(err) != 0 {
			s = buildReply(fmt.Sprintf(`'{"error": "%s"}'`, err))
		} else {
			s = buildReply(`'{"error": "No code/state returned"}'`)
		}

		res.WriteHeader(200)
		res.Write([]byte(s))
	} else {
		if x, found := states[state]; found {
			fmt.Printf("should verify state is less then 30 seconds %v\n", time.Now().UnixNano()-x)
			delete(states, state)
		} else {
			fmt.Printf("Error in /code call, state not found: %s\n", state)
			s := buildReply(fmt.Sprintf(`'{"error": "state not found: %s"}'`, state))
			res.WriteHeader(200)
			res.Write([]byte(s))
			return
		}

		client := http.Client{}
		v := url.Values{
			"grant_type":   {"authorization_code"},
			"redirect_uri": {config.REDIRECT_URI},
			"code":         {code},
		}
		req, err := http.NewRequest("POST", config.ISSUER+"/v1/token", strings.NewReader(v.Encode()))
		if err != nil {
			fmt.Printf("Error in setting up /token call: %s\n", err.Error())
			s := buildReply(fmt.Sprintf(`'{"error": "%s"}'`, err.Error()))
			res.WriteHeader(200)
			res.Write([]byte(s))
			return
		}

		req.SetBasicAuth(config.CLIENT_ID, config.CLIENT_SECRET)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error in /token call: %s\n", err.Error())
			s := buildReply(fmt.Sprintf(`'{"error": "%s"}'`, err.Error()))
			res.WriteHeader(200)
			res.Write([]byte(s))
			return
		}
		if resp.StatusCode > 299 {
			err, _ := io.ReadAll(resp.Body)
			fmt.Printf("Error in /token call Return: %s\n", err)
			s := buildReply(fmt.Sprintf(`'{"error": "%s"}'`, err))
			res.WriteHeader(200)
			res.Write([]byte(s))
			return
		}

		defer resp.Body.Close()
		tokens := struct {
			Token_type   string `json:"token_type"`
			Expires_in   int    `json:"expires_in"`
			Access_token string `json:"access_token"`
			Scope        string `json:"scope"`
			Id_token     string `json:"id_token"`
		}{}

		fmt.Printf("\n/Token Response:\nStatus Code: %v\n%v\n", resp.StatusCode, resp)
		r, _ := io.ReadAll(resp.Body)
		json.Unmarshal(r, &tokens)
		toks, _ := json.MarshalIndent(tokens, "", "  ")
		fmt.Printf("\nTokens Received:\n%s\n", toks)
		decoded, _ := base64.RawStdEncoding.DecodeString(strings.Split(tokens.Id_token, ".")[1])
		fmt.Println(string(decoded))
		partialId := struct {
			Name      string `json:"name"`
			Auth_time int    `json:"auth_time"`
		}{}
		_ = json.Unmarshal(decoded, &partialId)
		s := buildReply(fmt.Sprintf(`'{"name": "%s", "auth_time": "%v"}'`, partialId.Name, partialId.Auth_time))
		fmt.Printf("Returning:\n%s\n", s)
		res.WriteHeader(200)
		res.Write([]byte(s))
	}
}

/*
 * Build response with Post Message to return to client
 */
func buildReply(content string) string {
	return fmt.Sprintf(
		`<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Document</title>
	</head>
	<body>
		<script type="text/javascript">
		  window.onload = () => {
			if (window.opener !== null) {
				window.opener.postMessage(%s, "*");
			}
		  }
  
		</script>
	</body>
	</html> `, content)
}

/*
 * Configuration bits
 */
type Okta_app struct {
	CLIENT_ID     string
	CLIENT_SECRET string
	ISSUER        string
	REDIRECT_URI  string
}

type Configuration struct {
	Okta_app
}

func loadConfig(c *Configuration) {
	buf, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("No Configuration file exists yet: %v\n", err)
		buf = []byte{}
	}

	err = yaml.Unmarshal(buf, c)
	if err != nil {
		log.Fatal(err)
	}
}
