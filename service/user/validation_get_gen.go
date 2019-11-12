// Code generated by go generate forge validation v0.2; DO NOT EDIT.

package user

func (r reqUserGetID) valid() error {
	if err := validhasUserid(r.Userid); err != nil {
		return err
	}
	return nil
}

func (r reqUserGetUsername) valid() error {
	if err := validhasUsername(r.Username); err != nil {
		return err
	}
	return nil
}

func (r reqGetRoleUser) valid() error {
	if err := validhasRole(r.Role); err != nil {
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

func (r reqGetUserBulk) valid() error {
	if err := validAmount(r.Amount); err != nil {
		return err
	}
	if err := validOffset(r.Offset); err != nil {
		return err
	}
	return nil
}

func (r reqGetUsers) valid() error {
	if err := validhasUserids(r.Userids); err != nil {
		return err
	}
	return nil
}
