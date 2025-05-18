package create_rule

type Action string

const (
	ActionAllow Action = "allow"
	ActionDeny  Action = "deny"
)

var actions = []Action{ActionAllow, ActionDeny}
