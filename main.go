package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
)

func checkDependencies(dependencies ...string) error {
	for _, dep := range dependencies {
		cmd := exec.Command(dep)
		cmd.Run()
		if cmd.ProcessState == nil {
			return errors.New("Could not satisfy dependency: " + dep)
		}
	}
	return nil
}

func main() {
	err := checkDependencies("ffmpeg", "youtube-dl")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Enter youtube URL: ")
	var url string
	fmt.Scan(&url)

	exp, _ := regexp.Compile(`[&?](list)=?([A-Za-z0-9_-]+)|(\/(channel|user)\/([A-Za-z0-9_-]+))`)
	matches := exp.FindAllStringSubmatch(url, 2)

	// check for malformed url / no matches
	if len(matches) > 0 {
		f := new(Fetcher)

		// kinda hackish values, will fix later
		for _, match := range matches {
			if match[5] != "" {
				// begin download of user uploads
				fmt.Println("\nFound user:", match[5])
				f.getUseruploads(match[5])
			} else if match[1] == "list" {
				fmt.Println("\nPlaylist!")
				f.getPlaylist(match[2])
			}
		}

		// download each video and add appropriate tags/images
		for i, video := range f.playlist.videos {
			fmt.Printf("(%d/%d)	%s\n", i+1, f.playlist.items, video.title)
			f.getVideo(video, f.playlist.title, "")
		}

		fmt.Println("\nFinished!")

	} else {
		log.Fatal("Please enter a valid User or Playlist url")
	}
}
