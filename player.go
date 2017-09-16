package bluffy

type role uint

const (
	r_attacker role = iota
	r_defender
)

func (r role) String() string {
	if r > r_defender {
		return "Unknown role"
	}
	var roles = [...]string{
		"Attacker",
		"Defender",
	}
	return roles[r]
}

type player struct {
	id uint
	*account
	*match
	role
	points

	bluff suit
	pick  suit
}

func fight(ps []*player) {
	var atk, def *player
	for _, p := range ps {
		if p.role == r_attacker {
			atk = p
			continue
		}
		def = p
	}
	if atk == nil || def == nil {
		//this should never happen
		panic("Wrong roles in players of a match")
	}
	ap, dp := attackSuit(atk.pick, def.pick)
	atk.points += ap
	def.points += dp
	//TODO inform players about results
	return
}
