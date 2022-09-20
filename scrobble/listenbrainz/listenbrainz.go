package listenbrainz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/tpeacock19/gonic/db"
	"github.com/tpeacock19/gonic/scrobble"
)

const (
	BaseURL = "https://api.listenbrainz.org"

	submitPath           = "/1/submit-listens"
	listenTypeSingle     = "single"
	listenTypePlayingNow = "playing_now"
)

var (
	ErrListenBrainz = errors.New("listenbrainz error")
)

type AdditionalInfo struct {
	TrackNumber   int    `json:"tracknumber,omitempty"`
	TrackMBID     string `json:"track_mbid,omitempty"`
	RecordingMBID string `json:"recording_mbid,omitempty"`
	ReleaseMBID   string `json:"release_mbid,omitempty"`
	TrackLength   int    `json:"track_length,omitempty"`
}

type TrackMetadata struct {
	AdditionalInfo *AdditionalInfo `json:"additional_info"`
	ArtistName     string          `json:"artist_name,omitempty"`
	TrackName      string          `json:"track_name,omitempty"`
	ReleaseName    string          `json:"release_name,omitempty"`
}

type Payload struct {
	ListenedAt    int            `json:"listened_at,omitempty"`
	TrackMetadata *TrackMetadata `json:"track_metadata"`
}

type Scrobble struct {
	ListenType string     `json:"listen_type,omitempty"`
	Payload    []*Payload `json:"payload"`
}

type Scrobbler struct{}

func (s *Scrobbler) Scrobble(user *db.User, track *db.Track, stamp time.Time, submission bool) error {
	if user.ListenBrainzURL == "" || user.ListenBrainzToken == "" {
		return nil
	}
	payload := &Payload{
		TrackMetadata: &TrackMetadata{
			AdditionalInfo: &AdditionalInfo{
				TrackNumber:   track.TagTrackNumber,
				TrackMBID:     track.TagBrainzTrackID,
				RecordingMBID: track.TagBrainzRecordingID,
				ReleaseMBID:   track.TagBrainzReleaseID,
				TrackLength:   track.Length,
			},
			ArtistName:  track.TagTrackArtist,
			TrackName:   track.TagTitle,
			ReleaseName: track.Album.TagTitle,
		},
	}
	scrobble := Scrobble{
		Payload: []*Payload{payload},
	}
	if submission && len(scrobble.Payload) > 0 {
		scrobble.ListenType = listenTypeSingle
		scrobble.Payload[0].ListenedAt = int(stamp.Unix())
	} else {
		scrobble.ListenType = listenTypePlayingNow
	}

	var payloadBuf bytes.Buffer
	if err := json.NewEncoder(&payloadBuf).Encode(scrobble); err != nil {
		return err
	}
	submitURL := fmt.Sprintf("%s%s", user.ListenBrainzURL, submitPath)
	authHeader := fmt.Sprintf("Token %s", user.ListenBrainzToken)
	req, _ := http.NewRequest(http.MethodPost, submitURL, &payloadBuf)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", authHeader)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return fmt.Errorf("unathorized: %w", ErrListenBrainz)
	case resp.StatusCode >= 400:
		respBytes, _ := httputil.DumpResponse(resp, true)
		log.Printf("received bad listenbrainz response:\n%s", string(respBytes))
		return fmt.Errorf(">= 400: %d: %w", resp.StatusCode, ErrListenBrainz)
	}
	return nil
}

var _ scrobble.Scrobbler = (*Scrobbler)(nil)
