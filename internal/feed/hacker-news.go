package feed

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type hackerNewsPostResponseJson struct {
	Id           int    `json:"id"`
	Score        int    `json:"score"`
	Title        string `json:"title"`
	TargetUrl    string `json:"url,omitempty"`
	CommentCount int    `json:"descendants"`
	TimePosted   int64  `json:"time"`
}

func getHackerNewsTopPostIds() ([]int, error) {
	request, _ := http.NewRequest("GET", "https://hacker-news.firebaseio.com/v0/topstories.json", nil)
	response, err := decodeJsonFromRequest[[]int](defaultClient, request)

	if err != nil {
		return nil, fmt.Errorf("%w: could not fetch list of post IDs", ErrNoContent)
	}

	return response, nil
}

func getHackerNewsPostsFromIds(postIds []int) (ForumPosts, error) {
	requests := make([]*http.Request, len(postIds))

	for i, id := range postIds {
		request, _ := http.NewRequest("GET", fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id), nil)
		requests[i] = request
	}

	task := decodeJsonFromRequestTask[hackerNewsPostResponseJson](defaultClient)
	job := newJob(task, requests).withWorkers(30)
	results, errs, err := workerPoolDo(job)

	if err != nil {
		return nil, err
	}

	posts := make(ForumPosts, 0, len(postIds))

	for i := range results {
		if errs[i] != nil {
			slog.Error("Failed to fetch or parse hacker news post", "error", errs[i], "url", requests[i].URL)
			continue
		}

		posts = append(posts, ForumPost{
			Title:           results[i].Title,
			DiscussionUrl:   "https://news.ycombinator.com/item?id=" + strconv.Itoa(results[i].Id),
			TargetUrl:       results[i].TargetUrl,
			TargetUrlDomain: extractDomainFromUrl(results[i].TargetUrl),
			CommentCount:    results[i].CommentCount,
			Score:           results[i].Score,
			TimePosted:      time.Unix(results[i].TimePosted, 0),
		})
	}

	if len(posts) == 0 {
		return nil, ErrNoContent
	}

	if len(posts) != len(postIds) {
		return posts, fmt.Errorf("%w could not fetch some hacker news posts", ErrPartialContent)
	}

	return posts, nil
}

func FetchHackerNewsTopPosts(limit int) (ForumPosts, error) {
	postIds, err := getHackerNewsTopPostIds()

	if err != nil {
		return nil, err
	}

	if len(postIds) > limit {
		postIds = postIds[:limit]
	}

	return getHackerNewsPostsFromIds(postIds)
}
