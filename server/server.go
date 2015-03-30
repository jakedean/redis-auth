/*
  This is the package that will manage the authentication
  with a redis server for the clients that are making HTTP
  calls to it.

  For every call that is made we will make sure the user
  is logged in, if they are we will carry out thier request
  if not we will tell them they need to log in.
  The request and the response will be made with JSON.

  The possible requests are:
  - /log_in
  - /log_out
  - /create_user
  - /about_me
  - /projects

  The structure we will have on out redis server will be as
  follows:
    - User hash map that will map a user_id to all of the
      auth info about that user => user_id, user_email,
      user_password and user_cookie.
    - Users hash map that will map user_email to user_id.
      This will be useful when the user is trying to login
      and all we have is the user email and we need to map
      the email to the rest of the user info.
    - Cookies hash map that will map user_cookie to user_id
      so when we want to check if the user is logged in we
      can grab the cookie from the request and check other
      user info if the cookie key exists in the Cookies hash.
 */

package main

import (
      "net/http"
      "io"
      "github.com/garyburd/redigo/redis"
      "fmt"
      "strconv"
      "time"
)

var redisCn redis.Conn

// Show some text for the about me and projects page.
func ShowStaticContent(w http.ResponseWriter, req *http.Request) {
  if ok := userLoggedIn(req); ok {
    switch req.URL.Path {
    case "/about_me":
      io.WriteString(w, "Jake Dean is a cool cat working at Wayfair.")
    case "/projects":
      io.WriteString(w, "You can see his projects at http://theuinversalcatalyst.herokuapp.com")
    default:
      io.WriteString(w, "We could not find the page "+req.URL.Path+", please try again.")
    }
  } else {
    io.WriteString(w, "Whoa, you need to be logged in to see that, login here: /login --email=your_email --password=your_pw")
  }
}

// Function to see if this user is logged in, we will check their
// cookie that they sent with the http request against the cookie
// hashmap.
func userLoggedIn(req *http.Request) bool {
  if req.Header["X-User-Cookie"] != nil {
    // We will go into the cookies hashmap and try to find this
    // cookie, if we can find it they are logged in, if not then no.
    cook, err := redis.String(redisCn.Do("HGET", "cookies", req.Header["X-User-Cookie"][0]))
    if cook == "" || err != nil {
      return false
    } else {
      return true
    }
  } else {
    // Dont even have a cookie, nope no good.
    return false
  }
}

// This function will create a new user for us if they have
// given us all of the info that we need.
func CreateUser(w http.ResponseWriter, req *http.Request) {
  // We need to make sure they have an email and a password.
  email := req.PostFormValue("email")
  pw := req.PostFormValue("password")
  fmt.Println("email: ", email)
  fmt.Println("pw: ", pw)
  fmt.Println(req)
  if email == "" || pw == "" {
    io.WriteString(w, "You must provide an email and password")
  } else {
    // Add this user to the user:user_id hash and the users hash.
    uid, _ := redis.Int(redisCn.Do("INCR", "next_user_id"))
    _, _ = redisCn.Do("HMSET", "user:"+strconv.Itoa(uid), "email", email, "password", pw, "cookie", strconv.FormatInt(time.Now().Unix(), 30))
    _, _ = redisCn.Do("HSET", "users", "email", email, "user_id", uid)
    io.WriteString(w, "Your account has been created please login to use it.")
  }

}

func main() {
  // Get the connection to redis running locally on the
  // default port 6397.
  cn, err := redis.Dial("tcp", ":6379")
  if err != nil {
    panic("Oh no we could not get a connection to redis!")
  }
  defer cn.Close()

  // Toss a pointer to this redis connectoin up to the global
  // scope so we can have access to it in all of the package funcs.
  redisCn = cn

  // Just init our user count here to 0 so the first user who creates
  // a new account will get user number 1.
  //_, err = cn.Do("SET", "next_user_id", "0")
  //if err != nil {
    //panic("We were not able to init the user id count, oh no.")
  //}
  fmt.Println("About to get the next user id")
  res, _ := redis.Int(cn.Do("GET", "next_user_id"))
  fmt.Println(res)

  // Now we will define our handler functions and start
  // up the http server.
  http.HandleFunc("/about_me", ShowStaticContent)
  http.HandleFunc("/projects", ShowStaticContent)
  //http.HandleFunc("/log_in", LogUserIn)
  //http.HandleFunc("/log_out", LogUserOut)
  http.HandleFunc("/create_user", CreateUser)
  err = http.ListenAndServe("localhost:3000", nil)
  if err != nil {
    panic(err)
  }
}