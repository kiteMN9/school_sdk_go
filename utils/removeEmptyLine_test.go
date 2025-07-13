package utils

import (
	"fmt"
	"testing"
)

func Test_RemoveEmptyLines(t *testing.T) {
	htmlContent := `
<!doctype html>

<html lang="zh-CN">







<head>

	<title>&nbsp;</title>

	





<meta http-equiv="X-UA-Compatible" content="IE=edge" />

<meta name="viewport" content="width=device-width, initial-scale=1">

<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />

<meta name="Copyright" content="zfsoft" />	

<link rel="icon" href="/logo/favicon.ico?t=1739944884352" type="image/x-icon" />

<link rel="shortcut icon" href="/logo/favicon.ico?t=1739944884352" type="image/x-icon" />

<style type="text/css">	

	.active{font-weight: bolder;}

	#navbar-tabs li{ margin-top: 2px;}

	#navbar-tabs li a{ border-top: 2px solid transparent;}

	#navbar-tabs li.active a{border-top: 2px solid #0770cd;}

</style>





	

<!--jQuery核心框架库 -->

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/js/other_jquery/jquery.min.js?ver=28985913"></script>

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/js/jquery-migrate.min.js?ver=28985913"></script>

<!--jQuery浏览器检测 -->

<script type="text/javascript" src="/js/browse/browse-judge.js?ver=28985913"></script>



<!--Bootstrap布局框架-->

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/plugins/bootstrap/js/bootstrap.min.js?ver=28985913" charset="utf-8"></script>

<!--jQuery常用工具扩展库：基础工具,资源加载工具,元素尺寸相关工具 -->

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/js/zftal/jquery.utils.contact-min.js?ver=28985913" charset="utf-8"></script>

<!--jQuery基础工具库：$.browser,$.cookie,$.actual等 -->

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/js/zftal/jquery.plugins.contact-min.js?ver=28985913" charset="utf-8"></script>

<!--jQuery自定义event事件库 -->

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/js/zftal/jquery.events.contact-min.js?ver=28985913" charset="utf-8"></script>

<!--JavaScript对象扩展库：Array,Date,Number,String -->

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/js/zftal/jquery.extends.contact-min.js?ver=28985913" charset="utf-8"></script>

<!--Bootbox弹窗插件-->

<link rel="stylesheet" type="text/css" href="/zftal-ui-v5-1.0.2/assets/plugins/bootbox/css/bootbox.css?ver=28985913" />

<script src="/zftal-ui-v5-1.0.2/assets/plugins/bootbox/bootbox.concat-min.js?ver=28985913" type="text/javascript" charset="utf-8"></script>

<script src="/zftal-ui-v5-1.0.2/assets/plugins/bootbox/lang/zh_CN.js?ver=28985913" type="text/javascript" charset="utf-8"></script>



<!--jQuery模拟滚动条库-->

<link rel="stylesheet" type="text/css" href="/zftal-ui-v5-1.0.2/assets/plugins/customscrollbar/css/jquery.mCustomScrollbar.min.css?ver=28985913" />

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/plugins/customscrollbar/js/jquery.mCustomScrollbar.min.js?ver=28985913" charset="utf-8"></script>

<!--jQuery.chosen美化插件-->

<link rel="stylesheet" type="text/css" href="/zftal-ui-v5-1.0.2/assets/plugins/chosen/css/chosen-min.css?ver=28985913" />

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/plugins/chosen/jquery.choosen.concat-min.js?ver=28985913" charset="utf-8"></script>

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/plugins/chosen/lang/zh_CN-min.js?ver=28985913" charset="utf-8"></script>

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/js/utils/jquery.utils.pinyin.min.js?ver=28985913" charset="utf-8"></script>

<!--[if lt IE 9]>

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/js/html5shiv.min.js?ver=28985913" charset="utf-8"></script>

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/js/respond.min.js?ver=28985913" charset="utf-8"></script>

<![endif]-->

<!--业务框架jQuery全局设置和通用函数库-->

<script type="text/javascript" src="/js/jquery.zftal.contact-min.js?ver=28985913"></script>

<!--国际化js库 -->

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/plugins/i18n/jquery.i18n-min.js?ver=28985913" charset="utf-8"></script>

<!--全局国际化js. -->

<script type="text/javascript" src="/js/globalweb/i18n-global_zh_CN.js?ver=28985913" charset="utf-8"></script>

<!--业务框架前端脚本国际化库-->

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/js/zftal/lang/jquery.zftal_zh_CN-min.js?ver=28985913"></script>

<!--密码强弱判断-->

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/js/utils/jquery.utils.strength.min.js?ver=28985913"></script>




<script type="text/javascript" src="/js\globalweb\comp\i18n\N253512_zh_CN.js?ver=28985913" charset="utf-8"></script>






</head>

<body>



<input type="hidden" id="dingdingbj" name="dingdingbj" value="" />

<div id="mobile-div" style="display:none">

			<div class="container">

			<div class="navbar-header">

				<button class="navbar-toggle" type="button" data-toggle="collapse" data-target=".bs-navbar-collapse">

					<span class="sr-only"> 自主选课</span> 

					<span class="icon-bar"></span> 

					<span class="icon-bar"></span> 

					<span class="icon-bar"></span>

				</button>

				<a href="#" id="topButton" class="navbar-brand" onclick="onClickMenu('/xsxk/zzxkyzb_cxZzxkYzbIndex.html','N253512')">

					自主选课

				</a>

				<script type="text/javascript">

					document.title="自主选课";

				</script>

			</div>

		</div>

<!-- navbar-end  -->

	<!-- 判断是否可切换角色 -->

	

</div>



	<!-- 头部 开始 -->

	<header class="navbar-inverse top2" id="show-head">

		<div class="container" id="navbar_container">

			<!-- 判断是否可切换角色 -->

			

					<div class="container">

			<div class="navbar-header">

				<button class="navbar-toggle" type="button" data-toggle="collapse" data-target=".bs-navbar-collapse">

					<span class="sr-only"> 自主选课</span> 

					<span class="icon-bar"></span> 

					<span class="icon-bar"></span> 

					<span class="icon-bar"></span>

				</button>

				<a href="#" id="topButton" class="navbar-brand" onclick="onClickMenu('/xsxk/zzxkyzb_cxZzxkYzbIndex.html','N253512')">

					自主选课

				</a>

				<script type="text/javascript">

					document.title="自主选课";

				</script>

			</div>

		</div>

<!-- navbar-end  -->

		</div>

	</header>

	<script>

		if(window.self !== window.top){

			 $('body').css({

				"background": "#fff"

			}) 

			$('body').find('.navbar-inverse').hide();			

		}

	</script>

	

	<!--头部 结束 -->

	<div style="width: 100%; padding: 0px; margin: 0px;" id="bodyContainer">

		<!-- requestMap中的参数为系统级别控制参数，请勿删除 -->

		<form id="requestMap">

			 <input type="hidden" id="sessionUserKey" value="[REDACTED_ID]" /> 

			 

			 	<input type="hidden" id="gnmkdmKey" value="N253512" />

			 

			 

		</form>

		<div class="container container-func sl_all_bg" id="yhgnPage">

			<div id="innerContainer">

				<!-- 放置页面显示内容 -->

				

	<div class="row sl_add_btn">

	    <div id="btn-groups" class="col-sm-12">

	    	<!-- 加载当前菜单栏目下操作   -->

			<div class="col-sm-12 col-lg-12 col-md-12" style="border-width: 0px"><div class="btn-toolbar" role="toolbar" style="float:right;"><div class="btn-group" id="but_ancd"> </div> </div></div>

			<!-- 加载当前菜单栏目下操作 -->

	    </div>

	</div>

	<div id="searchBox"></div> 

 		<div class="col-md-12 col-sm-12 border-b"  style="padding:8px 0px;">

			<div style="float:left;padding:10px 15px;">

				<h5>

				<font id="xkxn"></font> 学年 <font id="xkxq"></font> 学期<font id="txt_xklc"></font><span id="sysj"></span>

				&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<b>本学期选课要求</b>总学分最低&nbsp;<font color="red">1</font>

				

					&nbsp;&nbsp;最高&nbsp;<font color="red">100</font>

				

				

				&nbsp;&nbsp;&nbsp;本学期已选学分&nbsp;&nbsp;<font color="red" id="yxxfs">0</font>

				

				

				

				</h5>

			</div>

			<div style="margin-right:20px;float:right;padding:8px 15px;">

				<!-- <div style="float:left;margin-top:-2px;margin-right:20px;font-size:20px">【<a style="text-decoration:underline;" href="javascript:void(0)" onclick="$.showDialog(_path+'/xjyj/xsxyqk_ckXsXyxxHtmlView.html','我的修业情况',$.extend({},viewConfig,{width: ($('#yhgnPage').innerWidth()-200)+'px'}));">我的修业情况</a>】</div> -->

				<div id="quickXk" style="float:left;"></div>

				<div style="float:left;">

					<p style="margin-top:4px;margin-right:5px;float:left;border:1px solid #BCE8F1;background-color:#D9EDF7;height:15px;width:30px;"></p>未选

				</div>

				<div style="float:left;margin-left:20px">

					<p style="margin-top:4px;margin-right:5px;float:left;border:1px solid #BCE8F1;background-color:#fff7b2;height:15px;width:30px;"></p>重修未选

				</div>

				<div style="float:left;margin-left:20px">

					<p style="margin-top:4px;margin-right:5px;float:left;border:1px solid #BCE8F1;background-color:#C1FFC1;height:15px;width:30px;"></p>已选

				</div>

			</div>

		</div>



		<input type="hidden" name="iskxk" id="iskxk" value="1"/>

		<input type="hidden" name="jgh_id" id="jgh_id"/>

		<input type="hidden" name="to_kch" id="to_kch" value=""/>

		<input type="hidden" name="jzxkf" id="jzxkf" value="0"/>

		<input type="hidden" name="xkzgmc" id="xkzgmc" value="100"/>

		<input type="hidden" name="xkzgxf" id="xkzgxf" value="100"/>

		<input type="hidden" name="zkcs" id="zkcs" value="15"/>

		<input type="hidden" name="zxfs" id="zxfs" value="26.0"/>

		<input type="hidden" name="bdzcbj" id="bdzcbj" value="2"/>

		<input type="hidden" name="xkxnm" id="xkxnm" value="2024"/>

		<input type="hidden" name="xkxqm" id="xkxqm" value="12"/>

		<input type="hidden" name="xkxnmc" id="xkxnmc" value="2024-2025"/>

		<input type="hidden" name="xkxqmc" id="xkxqmc" value="2"/>

		<input type="hidden" name="xh_id" id="xh_id" value="[REDACTED_ID]"/>

		<input type="hidden" name="xqh_id" id="xqh_id" value="3"/>

		<input type="hidden" name="jg_id_1" id="jg_id_1" value="003"/>

		<input type="hidden" name="zyh_id" id="zyh_id" value="03201"/>

		<input type="hidden" name="zymc" id="zymc" value="化学工程与工艺"/>

		<input type="hidden" name="zyfx_id" id="zyfx_id" value="wfx"/>

		<input type="hidden" name="njdm_id" id="njdm_id" value="2023"/>

		<input type="hidden" name="njmc" id="njmc" value="2023"/>

		<input type="hidden" name="bh_id" id="bh_id" value="20230054"/>

		<input type="hidden" name="xbm" id="xbm" value="1"/>

		<input type="hidden" name="zh" id="zh" value=""/>

		<input type="hidden" name="xslbdm" id="xslbdm" value="wlb"/>

		<input type="hidden" name="mzm" id="mzm" value="01"/>

		<input type="hidden" name="xz" id="xz" value="4"/>

		<input type="hidden" name="ccdm" id="ccdm" value="3"/>

		<input type="hidden" name="xsbj" id="xsbj" value="4294967296"/>

		<input type="hidden" name="sjhm" id="sjhm" value="w"/>

		<input type="hidden" name="xszxzt" id="xszxzt" value="1"/>

		<input type="hidden" name="njdm_id_1" id="njdm_id_1" value="2023"/>

		<input type="hidden" name="zyh_id_1" id="zyh_id_1" value="03201"/>

		<input type="hidden" name="sfxsxkbz" id="sfxsxkbz" value="0"/>

		<input type="hidden" name="sfxskssj" id="sfxskssj" value="0"/>

		<input type="hidden" name="wrljxbbhkg" id="wrljxbbhkg" value="0"/>

		<input type="hidden" name="jxbzbkg" id="jxbzbkg" value="1"/>

		<input type="hidden" name="tykpzykg" id="tykpzykg" value="0"/>

		<input type="hidden" name="tkdxyzms" id="tkdxyzms" value="0"/>

		<input type="hidden" name="jxbzhkg" id="jxbzhkg" value="0"/>

		<input type="hidden" name="xxdm" id="xxdm" value="10305"/>

		<input type="hidden" name="xkgwckg" id="xkgwckg" value="0"/>

		<input type="hidden" name="cxkctskg" id="cxkctskg" value="0"/>

		<input type="hidden" name="kxqxktskg" id="kxqxktskg" value="0"/>

		<input type="hidden" name="tbtkxqxktskg" id=tbtkxqxktskg value="0"/>

		<input type="hidden" name="xkkczdsqkg" id="xkkczdsqkg" value="1"/>

		<input type="hidden" name="xkmcjzxskcs" id="xkmcjzxskcs" value="10"/>

		

		<input type="hidden" name="xkckctkckg" id="xkckctkckg" value=""/>

		<input type="hidden" name="xkbzsyljkg" id="xkbzsyljkg" value="0"/>



		<input type="hidden" name="zzxkgjcxkg_kkxy" id="zzxkgjcxkg_kkxy" value="1"/>

		<input type="hidden" name="zzxkgjcxkg_xqu" id="zzxkgjcxkg_xqu" value="0"/>

		<input type="hidden" name="zzxkgjcxkg_yqu" id="zzxkgjcxkg_yqu" value="0"/>

		<input type="hidden" name="zzxkgjcxkg_tjbj" id="zzxkgjcxkg_tjbj" value="0"/>

		<input type="hidden" name="zzxkgjcxkg_kclb" id="zzxkgjcxkg_kclb" value="1"/>

		<input type="hidden" name="zzxkgjcxkg_kcxz" id="zzxkgjcxkg_kcxz" value="1"/>

		<input type="hidden" name="zzxkgjcxkg_jxms" id="zzxkgjcxkg_jxms" value="1"/>

		<input type="hidden" name="zzxkgjcxkg_kcgs" id="zzxkgjcxkg_kcgs" value="1"/>

		<input type="hidden" name="zzxkgjcxkg_skxq" id="zzxkgjcxkg_skxq" value="1"/>

		<input type="hidden" name="zzxkgjcxkg_skjc" id="zzxkgjcxkg_skjc" value="1"/>

		<input type="hidden" name="zzxkgjcxkg_jxb" id="zzxkgjcxkg_jxb" value="1"/>

		<input type="hidden" name="zzxkgjcxkg_sfcx" id="zzxkgjcxkg_sfcx" value="1"/>

		<input type="hidden" name="zzxkgjcxkg_ywyl" id="zzxkgjcxkg_ywyl" value="1"/>

		<input type="hidden" name="zzxkgjcxkg_sksjct" id="zzxkgjcxkg_sksjct" value="0"/>

		


		<!-- 只有一个页签时，不显示页签 -->

			

				<input type="hidden" name="firstKklxdm" id="firstKklxdm" value="06"/>

				<input type="hidden" name="firstKklxmc" id="firstKklxmc" value="板块课(12)"/>

				<input type="hidden" name="firstXkkzId" id="firstXkkzId" value="2E5212DA3FD34793E065FCFCFE1D0407"/>

				<input type="hidden" name="firstNjdmId" id="firstNjdmId" value="2023"/>

				<input type="hidden" name="firstZyhId" id="firstZyhId" value="03201"/>

		 	

		 	<div class="panel panel-info"><ul class="nav"></ul></div>

		

		

	

	<div id="displayBox"></div>

	<div id="choosedBox"></div>

	<div id="endsign" style="display:none; text-align:center; height: 50px"><i class="red">......已到最后......</i></div><!-- （共 <font id="searchCount"></font> 条记录） -->

	<!-- <div id="waitsign" style="display:none; text-align:center; height: 50px"><i class="red bigger-300 icon-spinner icon-spin"></i></div> -->

	<div id="more" style="text-align:center; display:none"><font color="#2a6496" size="5">[<a href="javascript:void(0)" onclick="loadCoursesByPaged();">点此查看更多</a>]</font></div>

	<!--jQuery.jqGrid -->

<link rel="stylesheet" type="text/css" href="/zftal-ui-v5-1.0.2/assets/plugins/jqGrid/css/jquery.jqgrid.css?ver=28985913" />

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/plugins/jqGrid/jquery.jqgrid.src-min.js?ver=28985913" charset="utf-8"></script>

<script type="text/javascript" src="/zftal-ui-v5-1.0.2/assets/plugins/jqGrid/jquery.jqgrid.contact-min.js?ver=28985913" charset="utf-8"></script>


	<!--jQuery.validation -->
<link rel="stylesheet" type="text/css" href="/zftal-ui-v5-1.0.2/assets/plugins/validation/css/jquery.validate-min.css?ver=28985913" />


	<link rel="stylesheet" type="text/css" href="/js/plugins/searchbox/css/jquery.searchbox-min.css?ver=28985913" />


			</div>

		</div>

	</div>

	<!-- footer -->

	

<!-- footer --> 



<div id="footerID" class="footer"  style="background-color: " >

	

	<p>版权所有&#169; Copyright 1999-2023 正方软件股份有限公司　　中国·杭州西湖区紫霞街176号 互联网创新创业园2号301&nbsp;&nbsp;&nbsp;版本V-8.0.0</p>

</div>









<!-- footer-end -->

	<!-- footer-end -->

</body>






<!-- 软件评价 相应ini文件引用-->





<link rel="stylesheet"  type="text/css" href="/js/plugins/tagtree/tagTree.css?ver=28985913"/>

<script type='text/javascript' src="/js/plugins/tagtree/tagtree.js?ver=28985913"></script>

<script type='text/javascript' src="/js/plugins/tagtree/tagtreeBusiness.js?ver=28985913"></script>



</html>`
	result := RemoveEmptyLines(htmlContent)
	fmt.Println(result)
}
