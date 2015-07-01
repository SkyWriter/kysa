package main

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Channel struct {
	Name string
	Url  string
}

type Channels []Channel

func (a Channels) Len() int           { return len(a) }
func (a Channels) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Channels) Less(i, j int) bool { return a[i].Name < a[j].Name }

func getUrl(url string) string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	response, err := client.Get(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	return string(contents)
}

func parseChannels(s string) Channels {
	lines := strings.Split(s, "\n")
	channels := make(Channels, 0, len(lines)/2)

	name := ""
	url := ""
	for _, line := range lines {
		if strings.Contains(line, "EXTINF") {
			parts := strings.SplitN(line, ",", 2)
			name = strings.TrimSpace(parts[1])
			continue
		}

		if strings.Contains(line, "#EXTGRP") {
			continue
		}

		if name != "" {
			url = strings.TrimSpace(line)

			channels = append(channels, Channel{name, url})
		}
	}

	return channels
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getLabelPath(channel Channel, extension string) string {
	return fmt.Sprintf("labels/%s.%s", getMD5Hash(channel.Name), extension)
}

func killProcessGroup(proc *os.Process) {
	pgid, err := syscall.Getpgid(proc.Pid)
	if err == nil {
		syscall.Kill(-pgid, 15)
	}
	proc.Kill()
	proc.Wait()

}

func makeLabels(channels Channels) {
	os.Mkdir("labels", 0755)
	for _, channel := range channels {
		filename := getLabelPath(channel, "gz")
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			fmt.Printf("  ... making label for %s\n", channel.Name)
			jpgFilename := getLabelPath(channel, "jpg")
			exec.Command("convert", "-background", "black", "-fill", "white",
				"-font", "Liberation-Sans-Bold", "-size", "768x",
				"-pointsize", "160", "", "-gravity", "center",
				fmt.Sprintf("caption:%s", channel.Name), jpgFilename).Output()

			fbiProc := exec.Command("fbi", "-noverbose", "-T", "1", "-1", jpgFilename)
			fbiProc.Start()
			time.Sleep(500 * time.Millisecond)

			exec.Command("/bin/sh", "-c", fmt.Sprintf("cat /dev/fb0 | gzip > %s", filename)).Output()
			exec.Command("/bin/sh", "-c", "pkill fbi").Output()

			os.Remove(jpgFilename)
		} else {
			fmt.Printf("  ... label for %s is present\n", channel.Name)
		}
	}
}

func switchChannel(channel Channel) *os.Process {
	exec.Command("/bin/sh", "-c", "setterm -blank off -powerdown off > /dev/tty0").Output()
	exec.Command("/bin/sh", "-c", "clear > /dev/tty0").Output()
	exec.Command("/bin/sh", "-c", "setterm -cursor off > /dev/tty0").Output()
	exec.Command("/bin/sh", "-c",
		fmt.Sprintf("zcat %s > /dev/fb0",
			getLabelPath(channel, "gz"))).Output()

	player := exec.Command("omxplayer", "-o", "hdmi", channel.Url)
	player.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := player.Start()
	if err != nil {
		panic(err)
	}

	return player.Process
}

func killChannel(proc *os.Process) {
	killProcessGroup(proc)
	exec.Command("/bin/sh", "-c", "clear > /dev/tty0")
	exec.Command("/bin/sh", "-c", "setterm -cursor on > /dev/tty0")
}

var channels Channels
var currentChannelNo = -1
var currentPlayerProcess *os.Process = nil

func turnOff() {
	if currentPlayerProcess != nil {
		killChannel(currentPlayerProcess)
		currentPlayerProcess = nil
	}
	currentChannelNo = -1
}

func changeChannel(channels Channels, no int) {
	if no >= len(channels) || no < 0 {
		return
	}

	turnOff()

	currentChannelNo = no
	currentPlayerProcess = switchChannel(channels[currentChannelNo])
}

func changeChannelHandler(w http.ResponseWriter, r *http.Request) {
	no, err := strconv.Atoi(r.FormValue("no"))
	fmt.Printf("Switching to %d\n", no)
	if err != nil {
		fmt.Fprintf(w, "ERROR")
	} else {
		changeChannel(channels, no)
		http.Redirect(w, r, "/", 301)
	}
}

func turnoffHandler(w http.ResponseWriter, r *http.Request) {
	turnOff()
	fmt.Printf("Turning off\n")
	http.Redirect(w, r, "/", 301)
}

func channelDirectoryHandler(w http.ResponseWriter, r *http.Request) {
	channelsJson := make([]string, 0, len(channels))
	for idx, channel := range channels {
		channelsJson = append(channelsJson, fmt.Sprintf("{ \"no\": %d, \"name\": \"%s\" }", idx, channel.Name))
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{ \"current\": %d, \"channels\": [ %s ] }", currentChannelNo, strings.Join(channelsJson, ", "))
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s _m3u_url_", os.Args[0])
		return
	}

	fmt.Println("Welcome to MyTV\n")

	channels = parseChannels(getUrl(os.Args[1]))
	sort.Sort(channels)

	fmt.Println("Making labels")
	makeLabels(channels)

	fmt.Println("Starting web server")

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.HandleFunc("/channels", channelDirectoryHandler)
	http.HandleFunc("/channel", changeChannelHandler)
	http.HandleFunc("/off", turnoffHandler)

	fmt.Println("Serving on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
