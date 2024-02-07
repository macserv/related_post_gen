package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

const topN = 5
const InitialTagMapSize = 100
const InitialPostsSliceCap = 0

type isize uint32

type Post struct {
	ID    string   `json:"_id"`
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

type RelatedPosts struct {
	ID      string      `json:"_id"`
	Tags    *[]string   `json:"tags"`
	Related [topN]*Post `json:"related"`
}

func main() {
	posts := getPosts()

	runtime.GC()

	start := time.Now()

	postsLen := len(posts)

	tagMap := make(map[string][]isize, InitialTagMapSize)

	for i, post := range posts {
		for _, tag := range post.Tags {
			tagMap[tag] = append(tagMap[tag], isize(i))
		}
	}

	allRelatedPosts := make([]RelatedPosts, postsLen)
	taggedPostCount := make([]byte, postsLen)

	resetSlice := func() {
		for i := range taggedPostCount {
			taggedPostCount[i] = 0
		}
	}

	for i := range postsLen {

		for _, tag := range posts[i].Tags {
			for _, otherPostIdx := range tagMap[tag] {
				taggedPostCount[otherPostIdx] += 1

			}
		}

		taggedPostCount[i] = 0 // Don't count self
		top5 := [topN * 2]int{}
		minTags := byte(0)

		for j := range postsLen {
			count := taggedPostCount[j]
			if count > minTags {
				upperBound := (topN - 2) * 2

				countInt := int(count)
				for upperBound >= 0 && countInt > top5[upperBound] {
					top5[upperBound+2] = top5[upperBound]
					top5[upperBound+3] = top5[upperBound+1]
					upperBound -= 2
				}

				insertPos := upperBound + 2
				top5[insertPos] = countInt
				top5[insertPos+1] = j

				minTags = byte(top5[topN*2-2])
			}
		}

		topPosts := [topN]*Post{}

		for j := 0; j < topN; j += 1 {
			index := top5[j*2+1]
			topPosts[j] = &posts[index]
		}

		allRelatedPosts[i] = RelatedPosts{
			ID:      posts[i].ID,
			Tags:    &posts[i].Tags,
			Related: topPosts,
		}

		resetSlice()
	}

	fmt.Println("Processing time (w/o IO):", time.Since(start))

	writeResults(allRelatedPosts)
}

func getPosts() []Post {
	file, _ := os.Open("../posts.json")
	var posts = make([]Post, 0, InitialPostsSliceCap)
	err := json.NewDecoder(file).Decode(&posts)
	if err != nil {
		fmt.Println(err)
	}

	return posts
}

func writeResults(allRelatedPosts []RelatedPosts) {
	file, _ := os.Create("../related_posts_go.json")
	_ = json.NewEncoder(file).Encode(allRelatedPosts)
}
