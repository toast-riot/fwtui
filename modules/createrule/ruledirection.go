package createrule

type Direction string

const (
	DirectionIn  Direction = "in"
	DirectionOut Direction = "out"
)

var directions = []Direction{DirectionIn, DirectionOut}
