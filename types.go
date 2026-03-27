package main

import "time"

type SpotifyData struct {
	IsPlaying 	bool  					`json:"isPlaying"`
	IsLocal   	bool  					`json:"isLocal"`
	Progress  	int64 					`json:"progress"`
	Duration  	int64 					`json:"duration"`
	Album     	SpotifyDataAlbum 		`json:"album"`
	Artists 	[]SpotifyDataHrefName 	`json:"artists"`
	Track 		SpotifyDataHrefName 	`json:"track"`
}

type SpotifyDataAlbum struct {
	Href  *string `json:"href"`
	Image *string `json:"image"`
	Name  string  `json:"name"`
}

type SpotifyDataHrefName struct {
	Href *string `json:"href"`
	Name string  `json:"name"`
}

type SpotifyToken struct {
	AccessToken string
	Date        time.Time
}

type SpotifyResponse struct {
	IsPlaying  bool  			   `json:"is_playing"`
	ProgressMS int64 			   `json:"progress_ms"`
	Item       SpotifyResponseItem `json:"item"`
}

type SpotifyResponseItem struct {
	IsLocal    	bool   						`json:"is_local"`
	DurationMS 	int64  						`json:"duration_ms"`
	Name       	string 						`json:"name"`
	ID         	string 						`json:"id"`
	Album      	SpotifyResponseItemAlbum 	`json:"album"`
	Artists 	[]SpotifyResponseItemArtist `json:"artists"`
}

type SpotifyResponseItemAlbum struct {
	ID     string 							`json:"id"`
	Name   string 							`json:"name"`
	Images []SpotifyResponseItemAlbumImage  `json:"images"`
}

type SpotifyResponseItemAlbumImage struct {
	URL string `json:"url"`
}

type SpotifyResponseItemArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}