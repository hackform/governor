package profile

import (
	"github.com/hackform/governor"
	"github.com/hackform/governor/service/image"
	"github.com/hackform/governor/service/user/gate"
	"github.com/labstack/echo"
	"io"
	"net/http"
)

type (
	reqProfileGetID struct {
		Userid string `json:"userid"`
	}

	reqProfileModel struct {
		Userid string `json:"-"`
		Email  string `json:"contact_email"`
		Bio    string `json:"bio"`
	}
)

func (r *reqProfileGetID) valid() error {
	if err := hasUserid(r.Userid); err != nil {
		return err
	}
	return nil
}

func (r *reqProfileModel) valid() error {
	if err := hasUserid(r.Userid); err != nil {
		return err
	}
	if err := validEmail(r.Email); err != nil {
		return err
	}
	if err := validBio(r.Email); err != nil {
		return err
	}
	return nil
}

func (p *profileRouter) createProfile(c echo.Context) error {
	rprofile := reqProfileModel{}
	if err := c.Bind(&rprofile); err != nil {
		return err
	}
	rprofile.Userid = c.Get("userid").(string)
	if err := rprofile.valid(); err != nil {
		return err
	}

	res, err := p.service.CreateProfile(rprofile.Userid, rprofile.Email, rprofile.Bio)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, res)
}

func (p *profileRouter) updateProfile(c echo.Context) error {
	rprofile := reqProfileModel{}
	if err := c.Bind(&rprofile); err != nil {
		return err
	}
	rprofile.Userid = c.Get("userid").(string)
	if err := rprofile.valid(); err != nil {
		return err
	}

	if err := p.service.UpdateProfile(rprofile.Userid, rprofile.Email, rprofile.Bio); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (p *profileRouter) updateImage(c echo.Context) error {
	img := c.Get("image").(io.Reader)
	imgSize := c.Get("imagesize").(int64)
	thumb64 := c.Get("thumbnail").(string)
	userid := c.Get("userid").(string)

	if err := p.service.UpdateImage(userid, img, imgSize, thumb64); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (p *profileRouter) deleteProfile(c echo.Context) error {
	rprofile := reqProfileGetID{
		Userid: c.Param("id"),
	}
	if err := rprofile.valid(); err != nil {
		return err
	}

	if err := p.service.DeleteProfile(rprofile.Userid); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (p *profileRouter) getOwnProfile(c echo.Context) error {
	userid := c.Get("userid").(string)

	res, err := p.service.GetProfile(userid)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (p *profileRouter) getProfile(c echo.Context) error {
	rprofile := reqProfileGetID{
		Userid: c.Param("id"),
	}
	if err := rprofile.valid(); err != nil {
		return err
	}

	res, err := p.service.GetProfile(rprofile.Userid)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (p *profileRouter) getProfileImage(c echo.Context) error {
	rprofile := reqProfileGetID{
		Userid: c.Param("id"),
	}
	if err := rprofile.valid(); err != nil {
		return err
	}

	image, contentType, err := p.service.GetProfileImage(rprofile.Userid)
	if err != nil {
		return err
	}
	return c.Stream(http.StatusOK, contentType, image)
}

func (p *profileRouter) getProfileImageCC(c echo.Context) (string, error) {
	rprofile := reqProfileGetID{
		Userid: c.Param("id"),
	}
	if err := rprofile.valid(); err != nil {
		return "", err
	}

	objinfo, err := p.service.StatProfileImage(rprofile.Userid)
	if err != nil {
		return "", err
	}

	return objinfo.ETag, nil
}

func (p *profileRouter) mountProfileRoutes(conf governor.Config, r *echo.Group) error {
	r.POST("", p.createProfile, gate.User(p.service.gate))
	r.PUT("", p.updateProfile, gate.User(p.service.gate))
	r.PUT("/image", p.updateImage, gate.User(p.service.gate), p.service.img.LoadJpeg("image", image.Options{
		Width:          384,
		Height:         384,
		ThumbWidth:     32,
		ThumbHeight:    32,
		Quality:        85,
		ThumbQuality:   85,
		Crop:           true,
		ContextField:   "image",
		SizeField:      "imagesize",
		ThumbnailField: "thumbnail",
	}))
	r.DELETE("/:id", p.deleteProfile, gate.OwnerOrAdmin(p.service.gate, "id"))
	r.GET("", p.getOwnProfile, gate.User(p.service.gate))
	r.GET("/:id", p.getProfile, p.service.cc.Control(true, false, min15, nil))
	r.GET("/:id/image", p.getProfileImage, p.service.cc.Control(true, false, min15, p.getProfileImageCC))
	return nil
}
