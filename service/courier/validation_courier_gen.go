// Code generated by go generate forge validation v0.3; DO NOT EDIT.

package courier

func (r reqLinkGet) valid() error {
	if err := validhasLinkID(r.LinkID); err != nil {
		return err
	}
	return nil
}

func (r reqGetGroup) valid() error {
	if err := validAmount(r.Amount); err != nil {
		return err
	}
	if err := validOffset(r.Offset); err != nil {
		return err
	}
	return nil
}

func (r reqLinkPost) valid() error {
	if err := validLinkID(r.LinkID); err != nil {
		return err
	}
	if err := validURL(r.URL); err != nil {
		return err
	}
	if err := validhasBrandID(r.BrandID); err != nil {
		return err
	}
	if err := validhasCreatorID(r.CreatorID); err != nil {
		return err
	}
	return nil
}

func (r reqBrandGet) valid() error {
	if err := validhasBrandID(r.BrandID); err != nil {
		return err
	}
	return nil
}

func (r reqBrandPost) valid() error {
	if err := validBrandID(r.BrandID); err != nil {
		return err
	}
	if err := validhasCreatorID(r.CreatorID); err != nil {
		return err
	}
	return nil
}
