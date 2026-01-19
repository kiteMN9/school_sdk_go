package cas2

import (
	"school_sdk/config"
	"testing"
)

func Test_Login(*testing.T) {
	cas := NewCas("", "", config.ChromeUA, false)
	cas.Login()
	//check_code.SaveImgStream(cas.getCaptchaImage(), "./kap_img/", "")
	cas.GetJwCookie()
}

func Test_wxLogin(*testing.T) {
	cas := NewCasWX("", "")
	//check_code.SaveImgStream(cas.getCaptchaImage(), "./kap_img/", "")
	cas.WXLogin()
}
