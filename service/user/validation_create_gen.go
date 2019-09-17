// Code generated by go generate forge validation v0.1.0. DO NOT EDIT.
package user

func (r reqUserPost) valid() error {
	if err := validUsername(r.Username); err != nil {
		return err
	}
	if err := validPassword(r.Password); err != nil {
		return err
	}
	if err := validEmail(r.Email); err != nil {
		return err
	}
	if err := validFirstName(r.FirstName); err != nil {
		return err
	}
	if err := validLastName(r.LastName); err != nil {
		return err
	}
	return nil
}

func (r reqUserPostConfirm) valid() error {
	if err := validhasToken(r.Key); err != nil {
		return err
	}
	return nil
}

func (r reqUserDelete) valid() error {
	if err := validhasUserid(r.Userid); err != nil {
		return err
	}
	if err := validhasUsername(r.Username); err != nil {
		return err
	}
	if err := validhasPassword(r.Password); err != nil {
		return err
	}
	return nil
}
