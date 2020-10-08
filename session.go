package tokbox

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type Session struct {
	SessionID      string `json:"session_id"`
	ProjectID      string `json:"project_id"`
	PartnerID      string `json:"partner_id"`
	CreateDt       string `json:"create_dt"`
	Status         string `json:"session_status"`
	MediaServerURL string `json:"media_server_url"`

	api *Tokbox
}

func (s *Session) Token(role Role, connectionData string, expiration int64) (string, error) {
	now := time.Now().UTC().Unix()

	toHMAC := url.Values{}
	toHMAC.Add("session_id", s.SessionID)
	toHMAC.Add("create_time", fmt.Sprintf("%d", now))
	if expiration > 0 {
		toHMAC.Add("expire_time", fmt.Sprintf("%d", now+expiration))
	}
	if len(role) > 0 {
		toHMAC.Add("role", string(role))
	}
	if len(connectionData) > 0 {
		toHMAC.Add("connection_data", connectionData)
	}
	toHMAC.Add("nonce", fmt.Sprintf("%d", rand.Intn(999999)))

	encoded := toHMAC.Encode()
	h := hmac.New(sha1.New, []byte(s.api.secret))
	n, err := h.Write([]byte(encoded))
	if err != nil {
		return "", err
	}

	if n != len(encoded) {
		return "", fmt.Errorf("hmac: not enough bytes written, expected %d, got %d", n, len(encoded))
	}

	values := url.Values{}
	values.Add("partner_id", s.api.key)
	values.Add("sig", fmt.Sprintf("%x:%s", h.Sum(nil), encoded))

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write([]byte(values.Encode()))
	encoder.Close()
	return fmt.Sprintf("T1==%s", buf.String()), nil
}

type startArchiveRequest struct {
	SessionID  string            `json:"sessionId"`
	HasAudio   bool              `json:"hasAudio"`
	HasVideo   bool              `json:"hasVideo"`
	Layout     map[string]string `json:"layout"`
	Name       string            `json:"name"`
	OutputMode string            `json:"outputMode"` // composed
	Resolution string            `json:"resolution"` // "1280x720" | "640x480"
}

type StartArchiveResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CreatedAt  int64  `json:"createdAt"`
	Duration   int    `json:"duration"`
	Event      string `json:"event"`
	HasAudio   bool   `json:"hasAudio"`
	HasVideo   bool   `json:"hasVideo"`
	OutputMode string `json:"outputMode"`
	PartnerID  int    `json:"partnerId"`
	Password   string `json:"password"`
	ProjectID  int    `json:"projectId"`
	Reason     string `json:"reason"`
	Resolution string `json:"resolution"`
	SessionID  string `json:"sessionId"`
	Sha256Sum  string `json:"sha256sum"`
	Size       int    `json:"size"`
	Status     string `json:"status"`
	UpdatedAt  int64  `json:"updatedAt"`
}

func newStartArchiveRequest(sid, name string) startArchiveRequest {
	return startArchiveRequest{
		SessionID: sid,
		Name:      name,
		HasAudio:  true,
		HasVideo:  true,
		Layout: map[string]string{
			"type": "bestFit",
		},
		OutputMode: "composed",
		Resolution: "1280x720",
	}
}

func (s *Session) StartArchive(name string) (*StartArchiveResponse, error) {
	reqBody := newStartArchiveRequest(s.SessionID, name)
	bs, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	target := fmt.Sprintf("https://api.opentok.com/v2/project/%s/archive", s.api.key)
	req, err := s.api.newRequest(http.MethodPost, target, bytes.NewReader(bs))
	if err != nil {
		return nil, err
	}

	resp, err := s.api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform http request: %w", err)
	}

	bs, _ = ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with non-200 code: %d: %s", resp.StatusCode, string(bs))
	}

	sar := StartArchiveResponse{}
	if err := json.Unmarshal(bs, &sar); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &sar, nil
}

func (s *Session) StopArchive(id string) error {
	target := fmt.Sprintf("https://api.opentok.com/v2/project/%s/archive/%s/stop", s.api.key, id)
	req, err := s.api.newRequest(http.MethodPost, target, nil)
	if err != nil {
		return err
	}

	resp, err := s.api.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform http request: %w", err)
	}

	bs, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with non-200 code: %d: %s", resp.StatusCode, string(bs))
	}

	return nil
}
