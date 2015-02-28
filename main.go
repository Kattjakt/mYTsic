package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
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

	var url string
	fmt.Print("Enter youtube URL: ")
	fmt.Scan(&url)

	choice, link := parseURL(url)
	f := new(Fetcher)

	switch choice {
	case CHANNEL:
		fmt.Println("Found user!")
		f.getUseruploads(link)
		break
	case PLAYLIST:
		fmt.Println("Found playlist!")
		f.getPlaylist(link)
		break
	}

	for i, video := range f.playlist.videos {
		fmt.Printf("(%d/%d)	%s\n", i+1, f.playlist.items, video.title)
		f.getVideo(video, f.playlist.title, "")
	}

	fmt.Println("Finished!")
}
