package commands

type DelCommand struct {
	Key  string
	line string
}

func (d *DelCommand) String() string {
	return d.line
}

func (d *DelCommand) GetAction() CommandAction {
	return Del
}
