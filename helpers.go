package main

import "net/url"

func spotifyAlbumHref(id string) *string {
	if id == "" {
		return nil
	}

	href := "https://open.spotify.com/album/" + id

	return &href
}

func spotifyArtistHref(id string) *string {
	if id == "" {
		return nil
	}

	href := "https://open.spotify.com/artist/" + id

	return &href
}

func spotifyTrackHref(id string) *string {
	if id == "" {
		return nil
	}

	href := "https://open.spotify.com/track/" + id
	
	return &href
}

func urlEncode(value string) string {
	return url.QueryEscape(value)
}