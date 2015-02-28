package main

import (
	"log"
	"regexp"
)

const (
	PLAYLIST = 1 << iota // 1 (i.e. 1 << 0)
	CHANNEL              // 2 (i.e. 1 << 1)
)

func parseURL(url string) (id int, link string) {
	// regex for matching playlists and channels
	exp, _ := regexp.Compile(`[&?](list)=?([A-Za-z0-9_-]+)|(\/(channel|user)\/([A-Za-z0-9_-]+))`)
	matches := exp.FindAllStringSubmatch(url, 2)

	// check for malformed url / no matches
	if len(matches) > 0 {
		for _, match := range matches {
			if match[5] != "" {
				return CHANNEL, match[5]
			} else if match[1] == "list" {
				return PLAYLIST, match[2]
			}
		}
	} else {
		log.Fatal("Please enter a valid User or Playlist url")
	}
	return
}
