package tokbox

// Adapted from https://github.com/cioc/tokbox

import (
	"log"
	"testing"
)

const key = "46748412"
const secret = "33b63be7b6b356ad6283be7b5184256dddd8dc26"

func TestToken(t *testing.T) {
	tokbox := New(key, secret)
	session, err := tokbox.NewSession("", P2P)
	if err != nil {
		log.Fatal(err)
		t.FailNow()
	}
	log.Println(session)
	token, err := session.Token(Publisher, "", Hours24)
	if err != nil {
		log.Fatal(err)
		t.FailNow()
	}
	log.Println(token)
}
