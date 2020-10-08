package tokbox

// Adapted from https://github.com/cioc/tokbox

import (
	"encoding/json"
	"fmt"
	"testing"
)

const key = "<key>"
const secret = "<secret>"

func TestToken(t *testing.T) {
	tokbox := New(key, secret)
	session, err := tokbox.NewSession("", P2P)
	if err != nil {
		t.Logf("failed to issue new session: %v", err)
		t.FailNow()
	}

	fmt.Println("sessid: ", session.SessionID)

	token, err := session.Token(Publisher, "", Hours24)
	if err != nil {
		t.Logf("failed to obtain token: %v", err)
		t.FailNow()
	}

	fmt.Println("token :: ", token)

	list, err := session.ArchiveList()
	if err != nil {
		t.Logf("failed to get list: %v", err)
		t.FailNow()
	}

	bs, _ := json.MarshalIndent(list, "", "  ")
	fmt.Printf(string(bs))
}
