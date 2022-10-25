package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
)

const (
	API_URL  = "https://api.spotify.com"
	AUTH_URL = "https://accounts.spotify.com"
)

func (s *Spotify) Init() {
	s.client = resty.New()
}

type Spotify struct {
	client *resty.Client
}

type UserPlaylists struct {
	Href string `json:"href"`

	Items []struct {
		Id     string `json:"id"`
		Name   string `json:"name"`
		Tracks struct {
			Href  string `json:"href"`
			Total int    `json:"total"`
		} `json:"tracks"`
	} `json:"items"`

	Limit    int    `json:"limit"`
	Next     string `json:"next"`
	Offset   int    `json:"offset"`
	Previous string `json:"previous"`
	Total    int    `json:"total"`
}

type UserPlaylist struct {
	Href string `json:"href"`

	Items []struct {
		Id    string `json:"id"`
		Track *struct {
			Id      string `json:"id"`
			Name    string `json:"name"`
			Artists []struct {
				Id   string `json:"id"`
				Name string `json:"name"`
			}
		} `json:"track"`
	} `json:"items"`

	Limit    int     `json:"limit"`
	Next     *string `json:"next"`
	Offset   int     `json:"offset"`
	Previous string  `json:"previous"`
	Total    int     `json:"total"`
}

func (s *Spotify) Authenticate(clientId string, clientSecret string) {
	s.client.SetBasicAuth(clientId, clientSecret)
	s.client.SetHostURL(AUTH_URL)

	s.client.SetFormData(map[string]string{
		"grant_type": "client_credentials",
	})

	resp, err := s.client.R().EnableTrace().Post("/api/token")

	if err != nil {
		panic(err)
	}

	var tokenData map[string]interface{}
	err = json.Unmarshal(resp.Body(), &tokenData)
	if err != nil {
		panic(err)
	}

	accessToken := tokenData["access_token"].(string)
	s.client.SetAuthToken(accessToken)
}

func (s *Spotify) SetAuthToken(accessToken string) {
	s.client.SetAuthToken(accessToken)
}

func (s *Spotify) GetPlaylists(username string) UserPlaylists {
	s.client.SetHostURL(API_URL)
	s.client.SetAuthScheme("Bearer")

	resp, err := s.client.R().EnableTrace().Get(fmt.Sprintf("/v1/users/%s/playlists", username))

	var userPlaylists UserPlaylists
	err = json.Unmarshal(resp.Body(), &userPlaylists)
	if err != nil {
		panic(err)
	}

	return userPlaylists
}

func (s *Spotify) GetTracksFromUrl(url string) *string {

	resp, err := s.client.R().EnableTrace().Get(url)

	var userPlaylist UserPlaylist
	err = json.Unmarshal(resp.Body(), &userPlaylist)
	if err != nil {
		panic(err)
	}

	for _, track := range userPlaylist.Items {
		var artists string

		if track.Track == nil {
			continue
		}

		for _, artist := range track.Track.Artists {
			if len(artists) > 0 {
				artists += ", "
			}
			artists += artist.Name
		}
		fmt.Printf("%s - %s\n", track.Track.Name, artists)
	}

	return userPlaylist.Next
}

func (s *Spotify) GetPlaylistTracks(userPlaylists *UserPlaylists) {
	s.client.SetHostURL("")
	s.client.SetAuthScheme("Bearer")

	for _, item := range userPlaylists.Items {
		var next *string = &item.Tracks.Href
		for next != nil {
			next = s.GetTracksFromUrl(*next)
		}
	}
}

func run(clientId string, clientSecret string, authToken string, username string) {
	spotify := new(Spotify)
	spotify.Init()

	if len(authToken) > 0 {
		spotify.SetAuthToken(authToken)
	} else {
		spotify.Authenticate(clientId, clientSecret)
	}

	userPlaylists := spotify.GetPlaylists(username)

	spotify.GetPlaylistTracks(&userPlaylists)
}

func main() {
	var clientId = os.Getenv("SPOTIFY_CLIENT_ID")
	var clientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
	var authToken = os.Getenv("SPOTIFY_AUTH_TOKEN")
	var username = os.Args[1]

	run(clientId, clientSecret, authToken, username)
}
