package config

const (
	EdgeUA    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36 Edg/151.0.0.0"
	ChromeUA  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36"
	FireFoxUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:152.0) Gecko/20100101 Firefox/152.0"
	//ApppleUA = "Mozilla/5.0 (Macintosh; Apple macOS 26_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Safari/537.36 Chrome/142.0.0.0"
	//USERAGENT = "Mozilla/5.0 (iPad; CPU OS 8_0_2 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Version/8.0 Mobile/12A405 Safari/600.1.4"
)

//func CheckUALegal(UA string) bool {
//	// 合法是true 非法是 false
//	if UA == "" {
//		return false
//	}
//	if strings.Contains(UA, "Mozilla/") || strings.Contains(UA, "/") {
//		return true
//	}
//	return false
//}

//if(intBro.msie==true||intBro.safari==true||intBro.mozilla==true||intBro.chrome==true){
//if(intBro.msie==true&&intBro.version<9){
//window.location.href = _path+"/xtgl/init_cxBrowser.html";
//}
