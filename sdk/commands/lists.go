package commands

type LPushCommand struct {
	Key    string
	Values []string
	line   string
}

func (lp *LPushCommand) String() string {
	return lp.line
}

func (lp *LPushCommand) GetAction() CommandAction {
	return LPush
}

type RPushCommand struct {
	Key    string
	Values []string
	line   string
}

func (rp *RPushCommand) String() string {
	return rp.line
}

func (rp *RPushCommand) GetAction() CommandAction {
	return RPush
}

type LLenCommand struct {
	Key  string
	line string
}

func (ll *LLenCommand) String() string {
	return ll.line
}

func (ll *LLenCommand) GetAction() CommandAction {
	return LLen
}

type LPopCommand struct {
	Key   string
	Count int
	line  string
}

func (lp *LPopCommand) String() string {
	return lp.line
}

func (lp *LPopCommand) GetAction() CommandAction {
	return LPop
}

type RPopCommand struct {
	Key   string
	Count int
	line  string
}

func (rp *RPopCommand) String() string {
	return rp.line
}

func (rp *RPopCommand) GetAction() CommandAction {
	return RPop
}

type LRangeCommand struct {
	Key   string
	Start int
	End   int
	line  string
}

func (lr *LRangeCommand) String() string {
	return lr.line
}

func (lr *LRangeCommand) GetAction() CommandAction {
	return LRange
}
