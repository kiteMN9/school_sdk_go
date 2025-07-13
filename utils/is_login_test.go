package utils

import (
	"fmt"
	"testing"
)

// func Test_identy(t *testing.T) {
func Test_User_is_login(t *testing.T) {
	account := "1111111"
	text := `
<!doctype html>

<html>

<head>
<script type="text/javascript">
	var _path 		= "";
	var _systemPath = "";
	var _stylePath  = "/zftal-ui-v5-1.0.2";
	var _reportPath = "http://10.0.4.7:80/WebReport";
	var _localeKey 			= "zh_CN";
	jQuery(function($){
		$('[data-toggle*="validation"]').trigger("validation");
		$('[data-toggle*="fixed"]').trigger("fixed");
		$('[data-toggle*="tagTree"]').trigger("tagTree");
		if($.fn.tooltip){
			$('[data-toggle*="tooltip"]').tooltip({container:'body'});
		}
	});
</script>

<style type="text/css">
	.captcha_modal{
		width: 380px;
		height: 330px;
		z-index: 9999;
		top: 150px;
		margin: auto;
		position: absolute;
		box-sizing: border-box;
		border-radius: 2px;
		background-color: #fff;
		box-shadow: 0 0 10px rgba(0,0,0,.3);
		left: 40%;
	}
</style>
<!-- 文件操作相关js -->
<script type="text/javascript" src="/js/globalweb/comp/file/file.js?ver=28985913"></script>
<!--教务系统通用业务js引用:比如学年，学期等公共的信息会放在这里-->
<script type="text/javascript" src="/js/globalweb/comp/i18n/jwglxt-common_zh_CN.js?ver=28985913"></script>
<!--业务模块的properties初始化-->
<!--国际化js库 -->
<script type="text/javascript" src="/js\globalweb\comp\i18n\N253512_zh_CN.js?ver=28985913" charset="utf-8"></script>
<!--移动端审核-->
	<style type="text/css">
		.kcgs,.kcxz{
			text-align:center
		}
	</style>
</head>


<body>
	<input type="hidden" name="rwlx" id="rwlx" value="2"/>
	<input type="hidden" name="kklxpx" id="kklxpx" value="4510"/>
	<input type="hidden" name="xklc" id="xklc" value="1"/>
	<input type="hidden" name="xklcmc" id="xklcmc" value="第1轮"/>
	<input type="hidden" name="bklx_id" id="bklx_id" value="0"/>
	<input type="hidden" name="zckz" id="zckz" value="0"/>
	<input type="hidden" name="sfzyxk" id="sfzyxk" value="0"/>
	<input type="hidden" name="sfyjxk" id="sfyjxk" value="0"/>
	<input type="hidden" name="zdzys" id="zdzys" value="1"/>
	<input type="hidden" name="sfqzxk" id="sfqzxk" value="0"/>
	<input type="hidden" name="sfkknj" id="sfkknj" value="0"/>
	<input type="hidden" name="gnjkxdnj" id="gnjkxdnj" value="0"/>
	<input type="hidden" name="sfkkzy" id="sfkkzy" value="0"/>
	<input type="hidden" name="kzybkxy" id="kzybkxy" value="0"/>
	<input type="hidden" name="sfznkx" id="sfznkx" value="0"/>
	<input type="hidden" name="zdkxms" id="zdkxms" value="0"/>
	<input type="hidden" name="sfkxq" id="sfkxq" value="0"/>
	<input type="hidden" name="sfkcfx" id="sfkcfx" value="0"/>
	<input type="hidden" name="kkbk" id="kkbk" value="0"/>
	<input type="hidden" name="kkbkdj" id="kkbkdj" value="0"/>
	<input type="hidden" name="sfkxk" id="sfkxk" value="1"/>
	<input type="hidden" name="sfktk" id="sfktk" value="1"/>
	<input type="hidden" name="txbsfrl" id="txbsfrl" value="0"/>
	<input type="hidden" name="tktjrs" id="tktjrs" value="0"/>
	<input type="hidden" name="rlzlkz" id="rlzlkz" value="1"/>
	<input type="hidden" name="rlkz" id="rlkz" value="0"/>
	<input type="hidden" name="xkly" id="xkly" value="0"/>
	<input type="hidden" name="sfyxsksjct" id="sfyxsksjct" value="0"/>
	<input type="hidden" name="ddkzbj" id="ddkzbj" value="0"/>
	<input type="hidden" name="cxddkzbj" id="cxddkzbj" value="0"/>
	<input type="hidden" name="sfkkjyxdxnxq" id="sfkkjyxdxnxq" value="0"/>
	<input type="hidden" name="sfkgbcx" id="sfkgbcx" value="0"/>
	<input type="hidden" name="xkxskcgskg" id="xkxskcgskg" value="1"/>
	<input type="hidden" name="jxbzcxskg" id="jxbzcxskg" value="0"/>
	<input type="hidden" name="xkzgbj" id="xkzgbj" value="0"/>
	<input type="hidden" name="sfrxtgkcxd" id="sfrxtgkcxd" value="0"/>
	<input type="hidden" name="tykczgxdcs" id="tykczgxdcs" value="0"/>
	<input type="hidden" name="sxrlkzlx" id="sxrlkzlx" value=""/>
	<input type="hidden" name="isinxksj" id="isinxksj" value="1"/>
	<input type="hidden" name="xksjxskz" id="xksjxskz" value="0"/>
	<input type="hidden" name="lnzkcs" id="lnzkcs" value=""/>
	<input type="hidden" name="lnzxfs" id="lnzxfs" value=""/>
	<input type="hidden" name="bxqzdxkxf" id="bxqzdxkxf" value="0"/>
	<input type="hidden" name="bxqzgxkxf" id="bxqzgxkxf" value="1.5"/>
	<input type="hidden" name="bxqzdxkmc" id="bxqzdxkmc" value="0"/>
	<input type="hidden" name="bxqzgxkmc" id="bxqzgxkmc" value="1"/>
	<input type="hidden" name="lnzdxkxf" id="lnzdxkxf" value="0"/>
	<input type="hidden" name="lnzgxkxf" id="lnzgxkxf" value="0"/>
	<input type="hidden" name="lnzdyhxf" id="lnzdyhxf" value="-1"/>
	<input type="hidden" name="lnzgyhxf" id="lnzgyhxf" value="-1"/>
	<input type="hidden" name="lnzdxkmc" id="lnzdxkmc" value="0"/>
	<input type="hidden" name="lnzgxkmc" id="lnzgxkmc" value="0"/>
	<input type="hidden" name="bbhzxjxb" id="bbhzxjxb" value="0"/>
	<input type="hidden" name="syts" id="syts" value="1"/>
	<input type="hidden" name="syxs" id="syxs" value="45"/>
	<input type="hidden" name="jdlx" id="jdlx" value="0"/>
	<input type="hidden" name="bjgkczxbbjwcx" id="bjgkczxbbjwcx" value="0"/>
	<input type="hidden" name="jspage" id="jspage" value="0"/>
	<input type="hidden" name="globJsPage" id="globJsPage" value="0"/>
	<input type="hidden" name="isEnd" id="isEnd" value="false"/>
	<div id="contentBox">
			<div class="clearfix"></div>
			<div class="panel panel-info">
				<div class="panel-heading">&nbsp;</div>
				<div class="panel-body">
					<!-- 请使用上方的查询工具条查询所需要选的教学班！ -->
					<div class="nodata"><span>请使用上方的查询工具条查询所需要选的教学班！</span></div>
				</div>
			</div>
	</div>
<script type="text/javascript" src="/js/comp/jwglxt/xkgl/xsxk/zzxkYzbZy.js?ver=28985913"></script>
</body>
</html>`
	s := UserIsLogin(account, text)
	fmt.Println(s)
}
