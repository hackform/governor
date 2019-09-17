// Code generated by go generate forge validation v0.1.0. DO NOT EDIT.
package profile

func (r reqProfileGetID) valid() error {
	if err := validhasUserid(r.Userid); err != nil {
		return err
	}
	return nil
}

func (r reqProfileModel) valid() error {
	if err := validhasUserid(r.Userid); err != nil {
		return err
	}
	if err := validEmail(r.Email); err != nil {
		return err
	}
	if err := validBio(r.Bio); err != nil {
		return err
	}
	return nil
}
