package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type YouTubeResponse struct {
	Video []Video `json:"items"`
}

type Video struct {
	Id Id `json:"id"`
}

type Id struct {
	VideoId string `json:"videoId"`
}

type ApiResponse struct {
	Videos []VideoResponse `json:"videos"`
}

type VideoResponse struct {
	ThumbnailUrl string `json:"thumbnailUrl"`
	VideoUrl     string `json:"videoUrl"`
}

func thumbnails(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("YT_API_KEY")
	if apiKey == "" {
		log.Fatal("Environment variable YT_API_KEY has not been set!")
	}

	channelId := r.URL.Query().Get("channelId")
	if channelId == "" {
		http.Error(w, "channelId is missing", 400)
		return
	}

	maxResults, err := strconv.Atoi(
		r.URL.Query().Get("maxResults"),
	)
	if err != nil {
		http.Error(w, "maxResults needs to be a valid Integer", 400)
		return
	}

	res, err := http.Get(
		buildYouTubeRequestUrl(
			channelId, maxResults, apiKey,
		),
	)
	if err != nil {
		http.Error(w, "An internal error occurred while making the request", 500)
		log.Print(err)
		return
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		http.Error(w, "YouTube API returned response code "+strconv.Itoa(res.StatusCode), 500)
		log.Printf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		return
	}
	if err != nil {
		http.Error(w, "An internal error occurred while reading the YouTube response", 500)
		log.Println(err)
		return
	}

	var youTubeResponse YouTubeResponse
	json.Unmarshal(body, &youTubeResponse)

	apiResponse := mapYouTubeResponseToJsonResponse(youTubeResponse)
	apiResponseJson, err := json.Marshal(apiResponse)
	if err != nil {
		http.Error(w, "An internal error occurred while generating the json response", 500)
		log.Print(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(apiResponseJson)
}

func mapYouTubeResponseToJsonResponse(response YouTubeResponse) ApiResponse {
	apiResponse := ApiResponse{}
	for _, video := range response.Video {
		videoResponse := VideoResponse{
			ThumbnailUrl: "https://img.youtube.com/vi/" + video.Id.VideoId + "/maxresdefault.jpg",
			VideoUrl:     "https://www.youtube.com/watch?v=" + video.Id.VideoId,
		}
		apiResponse.Videos = append(apiResponse.Videos, videoResponse)
	}
	return apiResponse
}

func buildYouTubeRequestUrl(channelId string, maxResults int, apiKey string) string {
	return "https://www.googleapis.com/youtube/v3/search?part=snippet" +
		"&channelId=" + channelId +
		"&maxResults=" + strconv.Itoa(maxResults) +
		"&order=date&type=video" +
		"&key=" + apiKey
}

func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Add("Access-Control-Allow-Methods", "GET")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func handleRequests() {
	http.HandleFunc("/thumbnails", CORS(thumbnails))
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func main() {
	handleRequests()
}
