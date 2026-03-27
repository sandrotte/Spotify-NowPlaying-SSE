package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	spotfiyMutex 	sync.Mutex
	token   *SpotifyToken
)

func getSpotifyData(context context.Context) (*SpotifyData, error) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	refreshToken := os.Getenv("SPOTIFY_REFRESH_TOKEN")

	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return nil, fmt.Errorf("missing env vars: SPOTIFY_CLIENT_ID, SPOTIFY_CLIENT_SECRET, SPOTIFY_REFRESH_TOKEN")
	}

	accessToken, err := getAccessToken(context, clientID, clientSecret, refreshToken)

	if err != nil {
		return nil, err
	}

	if accessToken == "" {
		return nil, nil
	}

	req, err := http.NewRequestWithContext(context, http.MethodGet, "https://api.spotify.com/v1/me/player/currently-playing", nil)

	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer " + accessToken)

	res, err := http.DefaultClient.Do(req)
	
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("spotify currently-playing failed: %s", res.Status)
	}

	var nowPlaying SpotifyResponse
	
	if err := json.NewDecoder(res.Body).Decode(&nowPlaying); err != nil {
		return nil, err
	}

	data := &SpotifyData{
		IsPlaying: nowPlaying.IsPlaying,
		IsLocal:   nowPlaying.Item.IsLocal,
		Progress:  nowPlaying.ProgressMS,
		Duration:  nowPlaying.Item.DurationMS,
	}

	data.Album.Href = spotifyAlbumHref(nowPlaying.Item.Album.ID)
	data.Album.Name = nowPlaying.Item.Album.Name

	if len(nowPlaying.Item.Album.Images) > 0 && nowPlaying.Item.Album.Images[0].URL != "" {
		image := nowPlaying.Item.Album.Images[0].URL
		data.Album.Image = &image
	}

	for _, artist := range nowPlaying.Item.Artists {
		data.Artists = append(data.Artists, SpotifyDataHrefName{
			Href: spotifyArtistHref(artist.ID),
			Name: artist.Name,
		})
	}

	data.Track.Href = spotifyTrackHref(nowPlaying.Item.ID)
	data.Track.Name = nowPlaying.Item.Name

	return data, nil
}

func getAccessToken(context context.Context, clientID, clientSecret, refreshToken string) (string, error) {
	spotfiyMutex.Lock()

	defer spotfiyMutex.Unlock()

	if token != nil && time.Since(token.Date) < 1*time.Hour {
		return token.AccessToken, nil
	}

	form := []byte("grant_type=refresh_token&refresh_token=" + urlEncode(refreshToken))

	req, err := http.NewRequestWithContext(context, http.MethodPost, "https://accounts.spotify.com/api/token", bytes.NewReader(form))

	if err != nil {
		return "", err
	}

	basic := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	
	req.Header.Set("Authorization", "Basic " + basic)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	
	if err != nil {
		token = nil
		return "", err
	}
	
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		token = nil
	
		return "", fmt.Errorf("spotify token refresh failed: %s", res.Status)
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	
	if err := json.NewDecoder(res.Body).Decode(&tokenResponse); err != nil {
		token = nil
	
		return "", err
	}

	if tokenResponse.AccessToken == "" {
		token = nil
	
		return "", nil
	}

	token = &SpotifyToken{
		AccessToken: tokenResponse.AccessToken,
		Date:        time.Now(),
	}
	
	return token.AccessToken, nil
}