package commands

import "fmt"

type AuthChallengeResponseCommand struct {
	ClientId           string
	ChallengeSignature string
	line               string
}

func (c *AuthChallengeResponseCommand) GetAction() CommandAction {
	return AuthChallengeResponse
}

func (c *AuthChallengeResponseCommand) String() string {
	return fmt.Sprintf("%s %s %s", AuthChallengeResponse, c.ClientId, c.ChallengeSignature)
}

type AuthChallengeRequestCommand struct {
	Challenge string
	line      string
}

func (a *AuthChallengeRequestCommand) GetAction() CommandAction {
	return AuthChallengeRequest
}

func (a *AuthChallengeRequestCommand) String() string {
	return fmt.Sprintf("%s %s", AuthChallengeRequest, a.Challenge)
}
