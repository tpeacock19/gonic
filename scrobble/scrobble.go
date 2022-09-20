package scrobble

import (
	"time"

	"github.com/tpeacock19/gonic/db"
)

type Scrobbler interface {
	Scrobble(user *db.User, track *db.Track, stamp time.Time, submission bool) error
}
