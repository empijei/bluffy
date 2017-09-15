package bluffy

type Score int
type Metadata struct{}

type Account struct {
	name     string
	score    Score
	metadata *Metadata
	p        *player
}

//TODO handle disconnection here
