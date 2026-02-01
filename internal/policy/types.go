package policy

type Decision struct {
	Action string `json:"action"`
	RuleID string `json:"rule_id"`
	Reason string `json:"reason"`
}

const ActionAllow = "allow"

const ActionDeny = "deny"

func NewAllowDecision(ruleID, reason string) Decision {
	return Decision{
		Action: ActionAllow,
		RuleID: ruleID,
		Reason: reason,
	}
}

func NewDenyDecision(ruleID, reason string) Decision {
	return Decision{
		Action: ActionDeny,
		RuleID: ruleID,
		Reason: reason,
	}
}

func (d Decision) IsAllowed() bool {
	return d.Action == ActionAllow
}
