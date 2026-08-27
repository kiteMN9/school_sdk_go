package cas2

import (
	"school_sdk/client/config"
	bconfig "school_sdk/config"
	"testing"
)

func Test_Login(*testing.T) {
	cas := NewCas("", "", bconfig.ChromeUA, false, &config.Data{})
	cas.Login()
	//check_code.SaveImgStream(cas.getCaptchaImage(), "./kap_img/", "")
	cas.GetJwCookie()
}

func Test_wxLogin(*testing.T) {
	cas := NewCasWX("", "")
	//check_code.SaveImgStream(cas.getCaptchaImage(), "./kap_img/", "")
	cas.WXLogin()
}
