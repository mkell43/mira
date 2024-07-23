package main

import (
	"github.com/thecsw/mira/v4"
)

func main() {
	r, _ := mira.Init(mira.ReadCredsFromFile("login.conf"))
	c, _ := r.Subreddit("all").StreamSubmissions()
	for {
		post := <-c
		r.Submission(post.GetId()).Save("hello there")
	}
}
