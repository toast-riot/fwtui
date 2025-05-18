package create_rule

type Action string

const (
	ActionAllow  Action = "allow"
	ActionDeny   Action = "deny"
	ActionReject Action = "reject"
)

var actions = []Action{ActionAllow, ActionDeny, ActionReject}
