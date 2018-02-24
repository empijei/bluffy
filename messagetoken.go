package bluffy

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

type Kind string

const (
	k_error  Kind = "error"
	k_status      = "status"
	k_update      = "update"
	k_token       = "token"
	k_action      = "action"
)

type message struct {
	Kind Kind
	Body string
}

func receiveMessage(r io.Reader) (m *message, err error) {
	dec := json.NewDecoder(r)
	m = &message{}
	err = dec.Decode(m)
	return
}

func (m *message) send(w io.Writer) error {
	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

var secret = "you should never hardcode secrets in sources"

type token struct {
	Nam string
	Sco Score
	CD  int64
	Sig string
}

func deserializeToken(s string) (*token, error) {
	buf, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var tok token
	err = json.Unmarshal(buf, &tok)
	return &tok, err
}

func (t *token) computeSign() string {
	type partialToken struct {
		Nam string
		Sco Score
		CD  int64
	}
	pt := partialToken{
		Nam: t.Nam,
		Sco: t.Sco,
		CD:  t.CD,
	}
	buf, err := json.Marshal(pt)
	if err != nil {
		//This should never happen
		panic(err)
	}
	h := sha256.New()
	_, _ = h.Write(buf)
	_, _ = h.Write([]byte(secret))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (t *token) sign() {
	t.Sig = t.computeSign()
}

func (t *token) isValid() bool {
	s := t.computeSign()
	if s != t.Sig {
		return false
	}
	return true
}

func (t *token) serialize() string {
	buf, err := json.Marshal(t)
	if err != nil {
		//This should never happen
		log.Println(err)
	}
	return base64.RawStdEncoding.EncodeToString(buf)
}

func (t *token) uid() uid {
	return uid(fmt.Sprintf("%s%d", t.Nam, t.CD))
}
