package bluffy

import (
	"fmt"
	"sync"
)

var connectedAccounts = struct {
	sync.RWMutex
	acc map[uid]*account
}{acc: make(map[uid]*account)}

type uid string

type Score int

/*
type command uint

const (
	c_bluff command = iota
	c_pick
	c_disconnect
)

func (c command) String() string {
	if c > c_disconnect {
		return "Unknown command"
	}
	var commands = [...]string{
		"bluff",
		"pick",
		"disconnect",
	}
	return commands[c]
}
*/

type account struct {
	name         string
	score        Score
	creationDate int64
	p            *player

	outMsg   chan *message
	listened chan struct{}
}

func accountFromToken(t *token) (a *account) {
	a = &account{
		name:         t.Nam,
		score:        t.Sco,
		creationDate: t.CD,
		outMsg:       make(chan *message, 10),
		listened:     make(chan struct{}, 2),
	}
	a.listened <- struct{}{}
	return
}

func (a *account) uid() uid {
	return uid(fmt.Sprintf("%s%d", a.name, a.creationDate))
}

func (a *account) bluff(s suit) error {
	return a.p.match.advance(playeraction{
		player: a.p,
		suit:   s,
		action: a_bluff,
	})
}

func (a *account) pick(s suit) error {
	return a.p.match.advance(playeraction{
		player: a.p,
		suit:   s,
		action: a_pick,
	})
}

func (a *account) disconnect() {
	_ = a.p.match.advance(playeraction{
		player: a.p,
		action: a_disconnect,
	})
}

func (a *account) send(m *message) {
	a.outMsg <- m
}
