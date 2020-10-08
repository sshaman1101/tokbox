package tokbox

// ArchiveStatusChanges notification body for the callback,
// see https://tokbox.com/developer/guides/archiving/ for more details.
type ArchiveStatusChanges struct {
	ID         string `json:"id"`
	Event      string `json:"event"`
	CreatedAt  int64  `json:"createdAt"`
	Duration   int64  `json:"duration"`
	Name       string `json:"name"`
	PartnerID  int64  `json:"partnerId"`
	Reason     string `json:"reason"`
	Resolution string `json:"resolution"`
	SessionID  string `json:"sessionId"`
	Size       int64  `json:"size"`
	Status     string `json:"status"`
	URL        string `json:"url"`
}
