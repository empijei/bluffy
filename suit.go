package bluffy

//suit is a playable in-game suite
type suit int

const (
	s_none suit = iota
	s_hearts
	s_tiles
	s_clovers
	s_pikes
)

func (s suit) String() string {
	if s > s_pikes {
		return "Unknown suit"
	}
	var suits = [...]string{
		"",        //none
		"Hearts",  //Cuori
		"Tiles",   //Quadri
		"Clovers", //Fiori
		"Pikes",   //Picche
	}
	return suits[s]
}

func attackSuit(a, d suit) (ap, sf points) {
	type pts int
	const (
		tie pts = iota
		atw
		dew
	)
	//rows are the atk choice, columns are the def choice
	var clash = [4][4]pts{
		[4]pts{tie, tie, dew, atw},
		[4]pts{atw, tie, tie, dew},
		[4]pts{dew, atw, tie, tie},
		[4]pts{tie, dew, atw, tie},
	}

	//-1 is to shift suit to start from 0 instead of 1
	result := clash[a-1][d-1]
	switch result {
	case atw:
		return 1, 0
	case dew:
		return 0, 1
	default: //tie
		return 0, 0
	}
}
