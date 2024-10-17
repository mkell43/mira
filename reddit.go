package mira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/thecsw/mira/v4/models"
)

// MiraRequest Reddit API is always developing and I can't implement all endpoints;
// It will be a bit of a bloat; Instead, you have accessto *Reddit.MiraRequest
// method that will let you to do any custom reddit api calls!
//
// Here is the signature:
//
//	func (c *Reddit) MiraRequest(method string, target string, payload map[string]string) ([]byte, error) {...}
//
// It is pretty straight-forward, the return is a slice of bytes; Parse it yourself.
func (c *Reddit) MiraRequest(method string, target string, payload map[string]string) ([]byte, error) {
	values := "?"
	for i, v := range payload {
		v = url.QueryEscape(v)
		values += fmt.Sprintf("%s=%s&", i, v)
	}
	values = values[:len(values)-1]
	r, err := http.NewRequest(method, target+values, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Set("User-Agent", c.Creds.UserAgent)
	r.Header.Set("Authorization", "Bearer "+c.Token)
	response, err := c.Client.Do(r)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Reddit returns integers for X-Ratelimit-Used and X-Ratelimit-Reset, but not for X-Ratelimit-Remaining
	if rateLimitUsed := response.Header.Get("X-Ratelimit-Used"); rateLimitUsed != "" {
		if used, err := strconv.Atoi(rateLimitUsed); err == nil {
			c.RateLimitUsed = used
		}
	}
	if rateLimitRemaining := response.Header.Get("X-Ratelimit-Remaining"); rateLimitRemaining != "" {
		if remaining, err := strconv.ParseFloat(rateLimitRemaining, 64); err == nil {
			c.RateLimitRemaining = remaining
		}
	}
	if rateLimitReset := response.Header.Get("X-Ratelimit-Reset"); rateLimitReset != "" {
		if reset, err := strconv.Atoi(rateLimitReset); err == nil {
			c.RateLimitReset = reset
		}
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	data := buf.Bytes()
	if err := findRedditError(data); err != nil {
		return nil, err
	}
	return data, nil
}

// Me pushes a new Redditor value.
func (c *Reddit) Me() *Reddit {
	return c.addQueue(c.Creds.Username, meType)
}

// Subreddit pushes a new subreddit value to the internal queue.
func (c *Reddit) Subreddit(name ...string) *Reddit {
	return c.addQueue(strings.Join(name, "+"), subredditType)
}

// Submission pushes a new submission value to the internal queue.
func (c *Reddit) Submission(name string) *Reddit {
	return c.addQueue(name, submissionType)
}

// Comment pushes a new comment value to the internal queue.
func (c *Reddit) Comment(name string) *Reddit {
	return c.addQueue(name, commentType)
}

// Redditor pushes a new redditor value to the internal queue.
func (c *Reddit) Redditor(name string) *Reddit {
	return c.addQueue(name, redditorType)
}

// Info returns MiraInterface of last pushed object.
func (c *Reddit) Info() (MiraInterface, error) {
	name, ttype := c.getQueue()
	switch ttype {
	case meType:
		return c.getMe()
	case submissionType:
		return c.getSubmission(name)
	case commentType:
		return c.getComment(name)
	case subredditType:
		return c.getSubreddit(name)
	case redditorType:
		return c.getUser(name)
	default:
		return nil, fmt.Errorf("returning type is not defined")
	}
}

func (c *Reddit) getMe() (models.Me, error) {
	target := RedditOauth + "/api/v1/me"
	ret := &models.Me{}
	ans, err := c.MiraRequest("GET", target, nil)
	if err != nil {
		return *ret, err
	}
	json.Unmarshal(ans, ret)
	return *ret, nil
}

func (c *Reddit) getSubmission(id string) (models.PostListingChild, error) {
	target := RedditOauth + "/api/info.json"
	ans, err := c.MiraRequest("GET", target, map[string]string{
		"id": id,
	})
	ret := &models.PostListing{}
	json.Unmarshal(ans, ret)
	if len(ret.GetChildren()) < 1 {
		return models.PostListingChild{}, fmt.Errorf("id not found")
	}
	return ret.GetChildren()[0], err
}

func (c *Reddit) getUser(name string) (models.Redditor, error) {
	target := RedditOauth + "/user/" + name + "/about"
	ans, err := c.MiraRequest(http.MethodGet, target, nil)
	ret := &models.Redditor{}
	json.Unmarshal(ans, ret)
	return *ret, err
}

func (c *Reddit) getSubreddit(name string) (models.Subreddit, error) {
	target := RedditOauth + "/r/" + name + "/about"
	ans, err := c.MiraRequest(http.MethodGet, target, nil)
	ret := &models.Subreddit{}
	json.Unmarshal(ans, ret)
	return *ret, err
}
