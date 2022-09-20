package spec

import "github.com/tpeacock19/gonic/db"

func NewInternetRadioStation(irs *db.InternetRadioStation) *InternetRadioStation {
	return &InternetRadioStation{
		ID:          irs.SID(),
		Name:        irs.Name,
		StreamURL:   irs.StreamURL,
		HomepageURL: irs.HomepageURL,
	}
}
