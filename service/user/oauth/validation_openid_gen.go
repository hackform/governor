// Code generated by go generate forge validation v0.3; DO NOT EDIT.

package oauth

func (r reqOAuthAuthorize) valid() error {
	if err := validhasUserid(r.Userid); err != nil {
		return err
	}
	if err := validhasClientID(r.ClientID); err != nil {
		return err
	}
	if err := validOidScope(r.Scope); err != nil {
		return err
	}
	if err := validOidNonce(r.Nonce); err != nil {
		return err
	}
	if err := validOidCodeChallenge(r.CodeChallenge); err != nil {
		return err
	}
	if err := validOidCodeChallengeMethod(r.CodeChallengeMethod); err != nil {
		return err
	}
	return nil
}

func (r reqGetConnectionGroup) valid() error {
	if err := validhasUserid(r.Userid); err != nil {
		return err
	}
	if err := validAmount(r.Amount); err != nil {
		return err
	}
	if err := validOffset(r.Offset); err != nil {
		return err
	}
	return nil
}

func (r reqGetConnection) valid() error {
	if err := validhasUserid(r.Userid); err != nil {
		return err
	}
	if err := validhasClientID(r.ClientID); err != nil {
		return err
	}
	return nil
}
