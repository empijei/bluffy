package bluffy

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
)

//action is a possible event fired by a player
type action uint

const (
	a_bluff action = iota
	a_pick
	a_win
	a_disconnect
)

func (a action) String() string {
	if a > a_disconnect {
		return "Unknown action"
	}
	var actions = [...]string{
		"Declared",
		"Decided",
		"Won",
		"Disconnected",
	}
	return actions[a]
}

//playeraction connects a player to an action on a suit
type playeraction struct {
	*player
	action
	suit
}

func (pa playeraction) String() string {
	return strings.TrimSpace(fmt.Sprintf("Player %d %s %s", pa.player.id, pa.action, pa.suit))
}

type points int

const winpoints points = 5

//match is an active match being played
type match struct {
	//state is the current state for the FSA representing the match state
	state state
	//players is the list of players
	players []*player //Maybe [2]*Player?
	//fightsLeft counts how many fights are left to fight
	fightsLeft int
	//wins is a list of indexes in the players list that keeps track of which

	sync.Mutex
	//keeps count of moves in a state
	movecounter int
	//err is used to keep track of errors
	err error
}

func newMatch(a1 *account, a2 *account) *match {
	shift := rand.Int() % 2
	i, j := shift%2, (shift+1)%2
	m := &match{}
	p1 := &player{
		id:      uint(i),
		account: a1,
		match:   m,
		role:    r_attacker,
	}
	a1.p = p1
	p2 := &player{
		id:      uint(j),
		account: a2,
		match:   m,
		role:    r_defender,
	}
	a2.p = p2
	m.players = make([]*player, 2)
	m.players[i] = p1
	m.players[j] = p2
	m.state = bluffState
	return m
}

func (m *match) advance(pa playeraction) error {
	m.Lock()
	defer m.Unlock()
	if m.state == nil {
		return matchHasEnded
	}
	if pa.action == a_disconnect {
		m.state = endMatchState
	}
	m.state = m.state(m, pa)
	if m.err != nil {
		err := m.err
		m.err = nil
		return err
	}
	return nil
}

//type representing the state of a match
type state func(*match, playeraction) state

var invalidAction = errors.New("Invalid action")
var matchHasEnded = errors.New("matchHasEnded")

func bluffState(m *match, pa playeraction) state {
	if pa.action != a_bluff {
		m.err = invalidAction
		return bluffState
	}
	switch {
	case m.movecounter < len(m.players)-1:
		pa.player.bluff = pa.suit
		m.movecounter++
		return bluffState
	case m.movecounter == len(m.players)-1:
		pa.player.bluff = pa.suit
		m.movecounter = 0
		//TODO notify players
		return betState
	default:
		m.err = invalidAction
		return bluffState
	}
}

func betState(m *match, pa playeraction) state {
	if pa.action != a_pick {
		m.err = invalidAction
		return betState
	}
	switch {
	case m.movecounter < len(m.players)-1:
		pa.player.pick = pa.suit
		m.movecounter++
		return betState
	case m.movecounter == len(m.players)-1:
		pa.player.pick = pa.suit
		fight(m.players)
		for _, p := range m.players {
			if p.points >= winpoints {
				//TODO inform about victory
				//update accounts

				//empty move on endMatchState
				return endMatchState(m, playeraction{player: p, action: a_win})
			}
		}
		m.movecounter = 0
		return bluffState
	default:
		m.err = invalidAction
		return betState
	}
}

func endMatchState(m *match, pa playeraction) state {
	//TODO inform players,
	//delete match,
	//destroy players,
	//disconnect accounts
	return nil
}
