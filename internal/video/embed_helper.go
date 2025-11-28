package video

import (
	"strings"
)

type EmbedType int

const (
	EmbedTypeNone EmbedType = iota
	EmbedTypeYouTube
	EmbedTypeVideo
	EmbedTypeIframe
)

type EmbedInfo struct {
	Type EmbedType
	URL  string
}

func GetEmbedInfo(link *string) EmbedInfo {
	if link == nil || *link == "" {
		return EmbedInfo{Type: EmbedTypeNone}
	}

	l := *link

	// Check for YouTube links
	if strings.Contains(l, "youtube.com") || strings.Contains(l, "youtu.be") {
		videoID := ""
		if strings.Contains(l, "youtube.com/watch?v=") {
			parts := strings.Split(l, "v=")
			if len(parts) > 1 {
				videoID = parts[1]
				// Clean up youtube link parameters, probably won't catch everything
				if idx := strings.Index(videoID, "&"); idx != -1 {
					// Go slices are so cool
					videoID = videoID[:idx]
				}
			}
		} else if strings.Contains(l, "youtu.be/") {
			parts := strings.Split(l, "youtu.be/")
			if len(parts) > 1 {
				videoID = parts[1]
				// Clean up youtube link parameters
				if idx := strings.Index(videoID, "?"); idx != -1 {
					videoID = videoID[:idx]
				}
			}
		} else if strings.Contains(l, "youtube.com/embed/") {
			return EmbedInfo{Type: EmbedTypeYouTube, URL: l}
		}

		if videoID != "" {
			return EmbedInfo{Type: EmbedTypeYouTube, URL: "https://www.youtube.com/embed/" + videoID}
		}
	}

	// Check for regular video files
	lower := strings.ToLower(l)
	if strings.HasSuffix(lower, ".mp4") || strings.HasSuffix(lower, ".webm") || strings.HasSuffix(lower, ".ogg") || strings.HasSuffix(lower, ".mov") {
		return EmbedInfo{Type: EmbedTypeVideo, URL: l}
	}

	// Default to generic iframe and hope for the best
	return EmbedInfo{Type: EmbedTypeIframe, URL: l}
}
