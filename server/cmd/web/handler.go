package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/google/uuid"
)

func (app *App) GetLoginForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<a href="/api/v1/auth/github">LOGIN</a>`)
}

func (app *App) GithubLoginHandler(w http.ResponseWriter, r *http.Request) {
	githubClientID := os.Getenv("GITHUB_CLIENT_ID")
	log.Printf("CLIENT ID %s", githubClientID)
	redirectURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s",
		githubClientID,
		"http://localhost:8080/auth/callback",
	)

	http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
}

func (app *App) GithubCallBackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	githubAccessToken := GetGithubAccessToken(code)

	githubData := GetGithubData(githubAccessToken)

	LoggedInHandler(w, r, githubData)

}

func LoggedInHandler(w http.ResponseWriter, r *http.Request, githubData string) {
	if githubData == "" {
		// Unauthorized users get an unauthorized message
		fmt.Fprintf(w, "UNAUTHORIZED!")
		return
	}

	// Set return type JSON
	w.Header().Set("Content-type", "application/json")

	// Prettifying the json
	var prettyJSON bytes.Buffer
	// json.indent is a library utility function to prettify JSON indentation
	parserr := json.Indent(&prettyJSON, []byte(githubData), "", "\t")
	if parserr != nil {
		log.Panic("JSON parse error")
	}

	// Return the prettified JSON as a string
	fmt.Fprintf(w, string(prettyJSON.Bytes()))
}

func GetGithubAccessToken(code string) string {

	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	// Set us the request body as JSON
	requestBodyMap := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	}
	requestJSON, err := json.Marshal(requestBodyMap)
	if err != nil {
		log.Fatal("SOMETHING FAILED WHEN CONVERTING IT TO STRING")
	} // JSON.stringify()

	// POST request to set URL
	req, err := http.NewRequest(
		"POST",
		"https://github.com/login/oauth/access_token",
		bytes.NewBuffer(requestJSON),
	)
	if err != nil {
		log.Panic("Request creation failed")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Get the response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Panic("Request failed")
	}
	// <BUFFER > --> String
	// Response body converted to stringified JSON
	body, _ := io.ReadAll(resp.Body)

	// Represents the response received from Github
	type githubAccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	// Convert stringified JSON to a struct object of type githubAccessTokenResponse
	var TokenResponse githubAccessTokenResponse
	//TODO store this in the DB
	json.Unmarshal(body, &TokenResponse)

	// Return the access token (as the rest of the
	// details are relatively unnecessary for us)
	return TokenResponse.AccessToken
}

func GetGithubData(accessToken string) string {
	// Get request to a set URL
	req, err := http.NewRequest(
		"GET",
		"https://api.github.com/user",
		nil,
	)
	if err != nil {
		log.Panic("API Request creation failed")
	}

	// Set the Authorization header before sending the request
	// Authorization: token XXXXXXXXXXXXXXXXXXXXXXXXXXX
	authorizationHeaderValue := fmt.Sprintf("token %s", accessToken)
	req.Header.Set("Authorization", authorizationHeaderValue)

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Panic("Request failed")
	}

	// Read the response as a byte slice
	res, _ := io.ReadAll(resp.Body)

	// Convert byte slice to string and return
	return string(res)
}

func (app App) ExecuteCode(w http.ResponseWriter, r *http.Request) {
	/*

		?{
			? Code : string
			? Language : string
		?}

	*/
	type RequestBody struct {
		Code     string
		Language string
	}
	var requestBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	createFile := func(code string, language string) string {
		randomString := uuid.NewString()
		extention := ""
		switch language {
		case "js":
			extention = "js"
		case "python":
			extention = "py"
		case "go":
			extention = "go"
		}

		return fmt.Sprintf("%s.%s", randomString, extention)
	}
	executeTheProgram := func(code string, language string, fileName string) []byte {
		cmd := ""
		additional := " "
		switch language {
		case "js":
			cmd = "node"
		case "python":
			cmd = "python"
		case "go":
			cmd = "go"
			additional = "run"
		}
		var command *exec.Cmd
		if additional != "" {
			command = exec.Command(cmd, additional, fileName)

		} else {
			command = exec.Command(cmd, fileName)

		}

		stdout, err := command.Output()
		fmt.Printf("%s OUTPUT %s ERROR", stdout, err)
		if err != nil {
			fmt.Println(err.Error())
			return []byte(err.Error())
		}
		return stdout
	}
	fileName := createFile(requestBody.Code, requestBody.Language)
	// path, err := os.Getwd()
	fmt.Printf("GOT THIS FILENAME %s", fileName)
	file, err := os.Create(fileName)
	if err != nil {
		http.Error(w, fmt.Sprintf("FAILED TO WRITE THE FILE, %s", err.Error()), http.StatusInternalServerError)
	}
	file.WriteString(requestBody.Code)
	output := executeTheProgram(requestBody.Code, requestBody.Language, fileName)
	file.Close()
	err = os.Remove(fileName)
	if err != nil {
		fmt.Println(fmt.Errorf("ERROR DELETING THE FILE, %s", err.Error()))
	}
	w.Write(output)
}
