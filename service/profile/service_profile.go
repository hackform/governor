package profile

import (
	"github.com/hackform/governor"
	"github.com/hackform/governor/service/image"
	"io"
	"net/http"
)

type (
	resProfileUpdate struct {
		Userid string `json:"userid"`
	}

	resProfileModel struct {
		Email string `json:"contact_email"`
		Bio   string `json:"bio"`
		Image string `json:"image"`
	}
)

func (p *profileService) CreateProfile(userid, email, bio string) (*resProfileUpdate, error) {
	m, err := p.repo.New(userid, email, bio)
	if err != nil {
		return nil, err
	}

	if err := p.repo.Insert(m); err != nil {
		if governor.ErrorStatus(err) == http.StatusBadRequest {
			return nil, governor.NewErrorUser("", 0, err)
		}
		return nil, err
	}

	return &resProfileUpdate{
		Userid: userid,
	}, nil
}

func (p *profileService) UpdateProfile(userid, email, bio string) error {
	m, err := p.repo.GetByID(userid)
	if err != nil {
		if governor.ErrorStatus(err) == http.StatusNotFound {
			return governor.NewErrorUser("", 0, err)
		}
		return err
	}

	m.Email = email
	m.Bio = bio

	if err := p.repo.Update(m); err != nil {
		return err
	}
	return nil
}

func (p *profileService) UpdateImage(userid string, img io.Reader, imgSize int64, thumb64 string) error {
	m, err := p.repo.GetByID(userid)
	if err != nil {
		if governor.ErrorStatus(err) == http.StatusNotFound {
			return governor.NewErrorUser("", 0, err)
		}
		return err
	}

	if err := p.obj.Put(userid+"-profile", image.MediaTypeJpeg, imgSize, img); err != nil {
		return governor.NewError("Failed to store profile picture", http.StatusInternalServerError, err)
	}

	m.Image = thumb64
	if err := p.repo.Update(m); err != nil {
		return err
	}
	return nil
}

func (p *profileService) DeleteProfile(userid string) error {
	m, err := p.repo.GetByID(userid)
	if err != nil {
		if governor.ErrorStatus(err) == http.StatusNotFound {
			return governor.NewErrorUser("", 0, err)
		}
		return err
	}

	if err := p.repo.Delete(m); err != nil {
		return err
	}
	return nil
}

func (p *profileService) GetProfile(userid string) (*resProfileModel, error) {
	m, err := p.repo.GetByID(userid)
	if err != nil {
		if governor.ErrorStatus(err) == http.StatusNotFound {
			return nil, governor.NewErrorUser("", 0, err)
		}
		return nil, err
	}
	return &resProfileModel{
		Email: m.Email,
		Bio:   m.Bio,
		Image: m.Image,
	}, nil
}
