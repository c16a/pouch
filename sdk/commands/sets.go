package commands

type SAddCommand struct {
	Key    string
	Values []string
	line   string
}

func (c *SAddCommand) GetAction() CommandAction {
	return SAdd
}

func (c *SAddCommand) String() string {
	return c.line
}

type SCardCommand struct {
	Key  string
	line string
}

func (c *SCardCommand) GetAction() CommandAction {
	return SCard
}

func (c *SCardCommand) String() string {
	return c.line
}

type SDiffCommand struct {
	Key       string
	OtherKeys []string
	line      string
}

func (c *SDiffCommand) GetAction() CommandAction {
	return SDiff
}

func (c *SDiffCommand) String() string {
	return c.line
}

type SInterCommand struct {
	Key       string
	OtherKeys []string
	line      string
}

func (c *SInterCommand) GetAction() CommandAction {
	return SInter
}

func (c *SInterCommand) String() string {
	return c.line
}

type SUnionCommand struct {
	Key       string
	OtherKeys []string
	line      string
}

func (c *SUnionCommand) GetAction() CommandAction {
	return SUnion
}

func (c *SUnionCommand) String() string {
	return c.line
}

type SIsMemberCommand struct {
	Key   string
	Value string
	line  string
}

func (c *SIsMemberCommand) GetAction() CommandAction {
	return SIsMember
}

func (c *SIsMemberCommand) String() string {
	return c.line
}

type SMembersCommand struct {
	Key  string
	line string
}

func (c *SMembersCommand) GetAction() CommandAction {
	return SMembers
}

func (c *SMembersCommand) String() string {
	return c.line
}
