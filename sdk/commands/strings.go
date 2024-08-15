package commands

type GetCommand struct {
	Key  string
	line string
}

func (g *GetCommand) GetAction() CommandAction {
	return Get
}

func (g *GetCommand) String() string {
	return g.line
}

type SetCommand struct {
	Key   string
	Value string
	line  string
}

func (s *SetCommand) GetAction() CommandAction {
	return Set
}

func (s *SetCommand) String() string {
	return s.line
}
