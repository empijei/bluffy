package bluffy

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"
)

var e_wrongAPI = errors.New("Wrong kind of message for this api")
var e_unauthorized = errors.New("Unauthorized")

type authCtxKey string

const c_token authCtxKey = "token"

var q = newQueue(1000)

//TODO implement inactivity kick/disconnect

func Serve(port, authSecret string) {
	//port := os.Getenv("PORT")
	//if port == "" {
	//log.Fatal("$PORT must be set")
	//}

	secret = authSecret

	http.HandleFunc("/enqueue", auth(enqueue))
	http.HandleFunc("/bluff", auth(bluff))
	http.HandleFunc("/pick", auth(pick))
	http.HandleFunc("/pushback", auth(pushback))
	http.HandleFunc("/register", register)

	err := http.ListenAndServe(":"+port, nil)
	log.Fatal(err)
}

func auth(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serTok := r.Header.Get(k_token)
		if serTok == "" {
			writeError(w, e_unauthorized, http.StatusUnauthorized)
			return
		}
		t, err := deserializeToken(serTok)
		if err != nil {
			log.Println(err)
			writeError(w, e_unauthorized, http.StatusUnauthorized)
			return
		}
		if !t.isValid() {
			writeError(w, e_unauthorized, http.StatusUnauthorized)
			return
		}
		c := context.WithValue(r.Context(), c_token, t)
		f(w, r.WithContext(c))
	}
}

func enqueue(w http.ResponseWriter, r *http.Request) {
	t := r.Context().Value(c_token).(*token)
	a := accountFromToken(t)
	connectedAccounts.Lock()
	connectedAccounts.acc[a.uid()] = a
	connectedAccounts.Unlock()
	err := q.enqueue(a)
	if err != nil {
		//TODO disconnect account
		writeInternalError(w, err)
		return
	}
	writeOK(w)
}

func bluff(w http.ResponseWriter, r *http.Request) {
	a, s, err := processAction(r)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	err = a.bluff(s)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	writeOK(w)
}

func pick(w http.ResponseWriter, r *http.Request) {
	a, s, err := processAction(r)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	err = a.pick(s)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	writeOK(w)
}

func register(w http.ResponseWriter, r *http.Request) {
	m, err := receiveMessage(r.Body)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	if m.Kind != k_token {
		writeInternalError(w, e_wrongAPI)
		return
	}
	if len(m.Body) > 23 {
		m.Body = m.Body[:23]
	}
	t := &token{
		Nam: m.Body,
		CD:  time.Now().UnixNano(),
	}
	t.sign()
	m.Body = t.serialize()
	writeMessage(w, m)
}

//this is a BOSH server→client endpoint
func pushback(w http.ResponseWriter, r *http.Request) {
	t := r.Context().Value(c_token).(*token)
	a, err := findAccount(t.uid())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	select {
	case <-a.listened:
		go func() {
			defer func() { a.listened <- struct{}{} }()
			m := <-a.outMsg
			if m == nil {
				m = &message{
					Kind: k_status,
					Body: "Disconnected",
				}
			}
			writeMessage(w, m)
		}()
	default:
		writeInternalError(w, errors.New("Already connected"))
		return
	}
}

func processAction(r *http.Request) (a *account, s suit, err error) {
	t := r.Context().Value(c_token).(*token)
	m, err := receiveMessage(r.Body)
	if err != nil {
		return
	}
	if m.Kind != k_action {
		err = e_wrongAPI
		return
	}
	ss := m.Body
	s, err = parseSuit(ss)
	if err != nil {
		return
	}
	a, err = findAccount(t.uid())
	return
}

func findAccount(u uid) (a *account, err error) {
	connectedAccounts.RLock()
	defer connectedAccounts.RUnlock()
	a, ok := connectedAccounts.acc[u]
	if !ok {
		return nil, errors.New("Account not connected")
	}
	return
}

func writeHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func writeError(w http.ResponseWriter, _err error, status int) {
	w.WriteHeader(status)
	writeHeaders(w)
	m := &message{Kind: k_error, Body: _err.Error()}
	err := m.send(w)
	//TODO
	_ = err
}

func writeInternalError(w http.ResponseWriter, err error) {
	writeError(w, err, http.StatusInternalServerError)
}

func writeOK(w http.ResponseWriter) {
	writeHeaders(w)
	m := &message{Kind: k_status, Body: "OK"}
	err := m.send(w)
	//TODO
	_ = err
}

func writeMessage(w http.ResponseWriter, m *message) {
	writeHeaders(w)
	err := m.send(w)
	//TODO
	_ = err
}
