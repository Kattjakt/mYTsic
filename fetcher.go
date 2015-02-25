package main

import (
	"errors"
	"fmt"
	json "github.com/likexian/simplejson-go"
	"github.com/oliamb/cutter"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

type Metadata struct {
	title  string
	artist string
	album  string // will be the playlist name (optional?)
}

type Video struct {
	id    string
	title string
	meta  Metadata
}

type Playlist struct {
	videos []Video
	title  string
	album  string
	items  int
}

type Fetcher struct {
	method   string
	playlist Playlist
}

const (
	ITEMS_PER_PAGE = 50
	MAX_ITERATIONS = 100
)

func (f *Fetcher) getUseruploads(username string) (err error) {
	// as the "uploaded" list doesn't have a name, set it to the username
	f.playlist.title = username

	// we get a 4XX if we specify page 0
	for page := 1; page < ITEMS_PER_PAGE*MAX_ITERATIONS; page += ITEMS_PER_PAGE {
		res, err := http.Get("http://gdata.youtube.com/feeds/api/users/" +
			username + "/uploads?alt=json&v=2&max-results=50&start-index=" +
			strconv.Itoa(page))

		if err != nil {
			log.Fatal(err)
		}

		defer res.Body.Close()
		if err != nil || res.StatusCode != 200 {
			log.Fatal("Unable to read json data!")
		}

		// translate the data into a parsable json format
		body, _ := ioutil.ReadAll(res.Body)
		parsed, _ := json.Loads(string(body))

		items, _ := parsed.Get("feed").Get("entry").Array()
		fmt.Println("Got JSON response with", len(items), "items!")

		for i, _ := range items {
			// make some scary 4am casts to get a readable title and id
			title_raw := items[i].(map[string]interface{})["title"]
			title := title_raw.(map[string]interface{})["$t"]

			id_raw := items[i].(map[string]interface{})["id"]
			id := id_raw.(map[string]interface{})["$t"]

			// add one slot to our playlist and append the title
			f.playlist.videos = append(f.playlist.videos, Video{})
			f.playlist.videos[f.playlist.items].title = title.(string)
			f.playlist.videos[f.playlist.items].id = id.(string)[27:]

			f.playlist.items++
		}

		totalitems, _ := parsed.Get("feed").Get("openSearch$totalResults").Get("$t").Int()
		if page+ITEMS_PER_PAGE > totalitems {
			fmt.Println("Done parsing JSON!")
			break
		}
	}
	return
}

func (f *Fetcher) getPlaylist(id string) (err error) {
	// we get a 4XX if we specify page 0
	for page := 1; page < ITEMS_PER_PAGE*MAX_ITERATIONS; page += ITEMS_PER_PAGE {
		res, err := http.Get("http://gdata.youtube.com/feeds/api/playlists/" +
			id + "?v=2&alt=jsonc&max-results=50&start-index=" +
			strconv.Itoa(page))

		if err != nil {
			log.Fatal(err)
		}

		defer res.Body.Close()
		if err != nil || res.StatusCode != 200 {
			log.Fatal("Unable to read json data!")
		}

		// translate the data into a parsable json format
		body, _ := ioutil.ReadAll(res.Body)
		parsed, _ := json.Loads(string(body))

		// set the playlist title to the playlist name
		// probably shouldn't be done in a loop but whatever
		f.playlist.title, _ = parsed.Get("data").Get("title").String()

		items, _ := parsed.Get("data").Get("items").Array()

		fmt.Println("Got JSON response with", len(items), "items!")
		for i, _ := range items {
			video_raw := items[i].(map[string]interface{})
			video := video_raw["video"].(map[string]interface{})

			f.playlist.videos = append(f.playlist.videos, Video{})
			f.playlist.videos[f.playlist.items].title = video["title"].(string)
			f.playlist.videos[f.playlist.items].id = video["id"].(string)

			f.playlist.items++
		}

		totalitems, _ := parsed.Get("data").Get("totalItems").Int()
		if page+ITEMS_PER_PAGE > totalitems {
			//f.playlist.items = totalitems
			fmt.Println("Done parsing JSON!")
			break
		}
	}
	return
}

func (f *Fetcher) getVideo(video Video, directory string, album string) (err error) {
	if video.id == "" || video.title == "" {
		return errors.New("Corrupt song")
	}

	// sanitize our shit so we don't get completely fucked over
	video.title = sanitize(video.title)

	// seperate video title and artist
	exp, _ := regexp.Compile(`([^-]+) - ([^-]+)-? ?([^-]+)?`)
	out := exp.FindStringSubmatch(video.title)
	if len(out) != 4 {
		return errors.New("Couldn't parse video title")
	} else {
		if out[3] == "" {
			video.meta.artist = out[1]
			video.meta.title = out[2]
		} else { // very much retarded, yes
			return errors.New("Couldn't parse metadata")
		}
	}

	// create some directories for our files if the don't already exist
	dirs := []string{"tmp", "converted", "converted/" + directory}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.Mkdir(dir, 0777)
		}
	}

	// download .mp4 file from youtube
	cmd := exec.Command("youtube-dl", video.id, "-o", "tmp/"+video.title+".mp4")
	cmd.Run()

	// some videos may be blocked in our country
	// or removed, and doesn't throw an error
	if _, err := os.Stat("tmp/" + video.title + ".mp4"); os.IsNotExist(err) {
		return errors.New("Unable to download song (Country Restrictions/removed)")
	}

	// download cover and crop it to a square
	err = getCover(video, "tmp/cover.jpg")
	if err != nil { // this is retarded, we should have a fallback photo
		return errors.New("No suitable photo found")
	}

	// convert video to .mp3 and add appropriate tags and cover image
	cmd = exec.Command(
		"ffmpeg",
		"-i", "tmp/"+video.title+".mp4",
		"-i", "tmp/cover.jpg",
		"-map", "0:1",
		"-map", "1:0",
		"-q:a", "0", // enable variable bitrate
		"-id3v2_version", "3",
		"-metadata", "comment=Cover (Front)",
		"-metadata", "title="+video.meta.title,
		"-metadata", "artist="+video.meta.artist,
		"-metadata", "album="+album,
		"-y", // always overwrite files
		"converted/"+directory+"/"+video.title+".mp3")

	// detailed error output
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}

	// just remove the temporary folder
	os.RemoveAll("tmp")

	return nil
}

func getCover(video Video, output string) (err error) {
	raw, err := http.Get("http://img.youtube.com/vi/" + video.id + "/maxresdefault.jpg")
	if err != nil {
		fmt.Printf("%s", err)
		return err
	} else {
		defer raw.Body.Close()

		// crop the image to a square
		cover, _ := jpeg.Decode(raw.Body)
		cropped, err := cutter.Crop(cover, cutter.Config{
			Width:   3, // aspect ratio
			Height:  3,
			Mode:    cutter.Centered,
			Options: cutter.Ratio,
		})

		// youtube may return a sample image sometimes, and we don't want that
		if cropped.Bounds().Max.X < 600 && cropped.Bounds().Max.Y < 600 {
			return errors.New("Cover image too small")
		}

		out, err := os.Create(output)
		if err != nil {
			fmt.Println(err)
		}
		defer out.Close()

		// write the file if everything went good
		jpeg.Encode(out, cropped, nil)
	}
	return err
}

func sanitize(query string) string {
	exp := regexp.MustCompile(`\/|\\|:|\*|\?|>|<|\"|\||~`)
	return exp.ReplaceAllLiteralString(query, "")
}
