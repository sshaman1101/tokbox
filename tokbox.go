package tokbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

const (
	Days30  = 2592000 // 30 * 24 * 60 * 60
	Weeks1  = 604800  // 7 * 24 * 60 * 60
	Hours24 = 86400   // 24 * 60 * 60
	Hours2  = 7200    // 60 * 60 * 2
	Hours1  = 3600    // 60 * 60
)

type MediaMode string

const (
	// The session will send streams using the OpenTok Media Router.
	MediaRouter MediaMode = "disabled"
	// The session will attempt send streams directly between clients. If clients cannot connect
	// due to firewall restrictions, the session uses the OpenTok TURN server to relay streams.
	P2P = "enabled"
)

type Role string

const (
	// Publisher can publish streams, subscribe to streams, and signal.
	Publisher Role = "publisher"
	// Subscriber can only subscribe to streams.
	Subscriber = "subscriber"
	// Moderator, in addition to the privileges granted to a publisher, in clients using the OpenTok.js 2.2
	// library, can call the `forceUnpublish()` and
	// `forceDisconnect()` method of the Session object.
	Moderator = "moderator"
)

type Tokbox struct {
	key    string
	secret string
	client *http.Client
}

func New(key, secret string) *Tokbox {
	return &Tokbox{
		key:    key,
		secret: secret,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (t *Tokbox) jwtToken() (string, error) {
	type TokboxClaims struct {
		jwt.StandardClaims
		Ist string `json:"ist,omitempty"`
	}

	claims := TokboxClaims{
		Ist: "project",
		StandardClaims: jwt.StandardClaims{
			Issuer:    t.key,
			IssuedAt:  time.Now().UTC().Unix(),
			ExpiresAt: time.Now().UTC().Unix() + (2 * 24 * 60 * 60), // 2 hours; //NB: The maximum allowed expiration time range is 5 minutes.
			Id:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.secret))
}

func (t *Tokbox) SessionFromID(id string) *Session {
	return &Session{
		api:       t,
		SessionID: id,
	}
}

// Creates a new tokbox session or returns an error.
// See README file for full documentation: https://github.com/pjebs/tokbox
func (t *Tokbox) NewSession(location string, mm MediaMode) (*Session, error) {
	params := url.Values{}
	params.Add("p2p.preference", string(mm))
	if len(location) > 0 {
		params.Add("location", location)
	}

	target := "https://api.opentok.com/session/create"
	req, err := t.newRequest(http.MethodPost, target, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	res, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tokbox returns non-200 error: %v", res.StatusCode)
	}

	var sessions []Session
	if err = json.NewDecoder(res.Body).Decode(&sessions); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response: %w", err)
	}

	if len(sessions) < 1 {
		return nil, fmt.Errorf("tokbox did not return a session")
	}

	o := sessions[0]
	o.api = t
	return &o, nil
}

// newRequest returns *http.Request with auth headers
func (t *Tokbox) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Create jwt token
	token, err := t.jwtToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-OPENTOK-AUTH", token)

	return req, nil
}
