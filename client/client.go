/*
  This will be a command line client that will send request
  to the server over http to see info.  The server will be managing
  the session with Redis so this is just to navigate around and
  see the results of logging in/ logging out and creating a new
  user.

  The options the client has to call are:
  /about_me -> Display some info about me if user logged in.
  /projects -> Tell user about my projects if user logged in.
  /log_in --username=Jake --password=pw -> Log user in if can.
  /log_out -> Log this user out if they are currently logged in
  /create_user --username=BabyT --password=pw -> create new user.
 */

package main

import (
  		"net/http"
  		"bufio"
  		"fmt"
      "os"
      "io/ioutil"
      "strings"
      "regexp"
      "bytes"
)

// This struct will represent the client.  We will get the user
// cookie from the server when the user logs in.
type Client struct {
	UserEmail    string
	UserPassword string
	UserCookie   string
}

var ourClient *Client = &Client{}
var httpClient *http.Client = &http.Client{}

func main() {
  giveUserInstructions()
  watchUserInput()
}

// Just tell the user about the stuff
func giveUserInstructions() {
  fmt.Println(
    `Welcome to the Redis auth client in GO.

     THis is a simple tool that will let you log into
     the web server and naigate to the following links:
     - /about_me
     - /projects
     - /log_in --username=Jake --password=pw
     - /log_out
     - /create_user --username=Jake2 --password=pw2

     If you try to naigate to a link and you are not logged
     in we will make you log in, give it a shot!.`)
}

// We will watch the user input here to see when they enter
// stuff into the console. This will be running in a goroutine.
func watchUserInput() {
  for {
    reader := bufio.NewReader(os.Stdin)
    input, _ := reader.ReadString('\n')
    // Strip the new line char off the end of the byte slice
    // and get this info formated for an http request.
    requestPt, ok := formatRequestObj(strings.TrimRight(input, " \n"))
    if !ok {
      fmt.Println("Make sure you format your request correctly.")
    } else {
      makeReqAndShowRes(requestPt)
    }
  }
}

// We will format an http request here and return a pointer
// to the request.
func formatRequestObj(input string) (*http.Request, bool) {
  redisUrl := "http://localhost:3000"
  if input == "/about_me" || input == "/projects" {
    // We just need to make a get request to this url
    // and send the cookie with it as well.
    req, _ := http.NewRequest("GET", redisUrl+input, nil)
    req.Header.Set("X-User-Cookie", ourClient.UserCookie)
    req.Header.Set("Content-Type", "application/json")
    return req, true
  } else if match, _ := regexp.MatchString("/create_user*", input); match  {
    email, emOk := getUserInput(input, "--email=")
    pw, pwOk := getUserInput(input, "--password=")

    if !emOk || !pwOk {
      return nil, false
    }

    // We are all good so we will create the body.
    bodyStr := []byte(`{"email" : "`+email+`", "password" : "`+pw+`"`)
    req, _ := http.NewRequest("POST", redisUrl+"/create_user", bytes.NewBuffer(bodyStr))
    req.Header.Set("X-User-Cookie", ourClient.UserCookie)
    req.Header.Set("Content-Type", "application/json")
    return req, true
  } else {
    return nil, false
  }

}

// Get the user input from the user command.
func getUserInput(input, flag string) (string, bool) {
  // first parse out the email.
  splitRes := strings.Split(input, flag)
  if len(splitRes) == 1 {
    // THe string did not match at all no go
    return "", false
  }
  // THe flag has matched so lets call flags on the second string
  // to split on whitespace.
  return strings.Fields(splitRes[1])[0], true
}

// Make the http request and print the body to the console.
func makeReqAndShowRes(req *http.Request) {
  fmt.Println(*req)
  res, err := httpClient.Do(req)
  if err != nil {
    panic(err)
  }
  defer res.Body.Close()
  body, _ := ioutil.ReadAll(res.Body)
  fmt.Println(string(body))
}