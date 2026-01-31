package GPA

import "testing"

func TestGPA(t *testing.T) {
	htmlText := `
<!doctype html>
<html>
<body>
<div class="row sl_add_btn">
    <div class="col-sm-12" style="margin-top: 10px;">
        <div class="col-sm-12 col-lg-12 col-md-12" style="border-width: 0px"><div class="btn-toolbar" role="toolbar" style="float:right;"><div class="btn-group" id="but_ancd"> </div> </div></div>
        <!-- 加载当前菜单栏目下操作 -->
    </div>
</div>

<!--查询条件  开始 -->
    <form id="form" name="form" action="/xsxy/xsxyqk_cxXsxyqkIndex.html" method="post" role="form">
        <input type="hidden" name="xh_id" value="220101011011" id="xh_id"/>
        <input type="hidden" name="xxdm" value="10305" id="xxdm"/>
        <input type="hidden" id="jsdm" name="jsdm" value='xs'/>
        <input type="hidden" name="jhwkcgsjdmc" id="jhwkcgsjdmc" value=""/>
        <input type="hidden" name="cjlrxn" value="2024" id="cjlrxn"/>
        <input type="hidden" name="cjlrxq" value="12" id="cjlrxq"/>
        <input type="hidden" name="bkcjlrxn" value="2020" id="bkcjlrxn"/>
        <input type="hidden" name="bkcjlrxq" value="12" id="bkcjlrxq"/>
        <input type="hidden" name="xscjcxkz" value="0" id="xscjcxkz"/>
        <input type="hidden" name="cjcxkzzt" value="0" id="cjcxkzzt"/>
        <input type="hidden" name="cjztkz" value="0" id="cjztkz"/>
        <input type="hidden" name="cjzt" value="" id="cjzt"/>
        <input type="hidden" name="xyxdqktsxx" value="" id="xyxdqktsxx"/>
        <div class="clearfix"></div>
	        <div style="margin-bottom:0px;margin-top: 20px" role="alert" id="alertBox">
	            <font size="2px" style="font-weight: bold">张三&nbsp;同学，您的课程修读情况（供参考）：(
	                统计时间<!-- 统计时间 -->2024-01-01 12:30:00之前有效<!-- 之前有效 -->)</font>
	            
	                
	                    <font size="2px">当前所有课程<!-- 当前所有课程 -->
			            	
	                       <a class="clj" name="showGpa"> 平均学分绩点<!-- 平均学分绩点 --></a>（GPA）：
	                        <font size="2px" style="color: red;">
	                            
	                            
	                            
	                            2.55
	                        </font>
	                    </font>&nbsp;
	            <font size="2px"> 计划总课程<!-- 计划总课程 -->&nbsp;118&nbsp;
	                门<!-- 门 -->      通过<!-- 通过 -->&nbsp;51&nbsp;
	                门<!-- 门 -->，<font size="2px" style="color: red;">未通过<!-- 未通过 -->
	                    &nbsp;12&nbsp;门&nbsp;<!-- 门 --></font>；未修<!-- 未修 -->&nbsp;54&nbsp;
	                门<!-- 门 -->；</font>
	            <font size="2px" style="color: blue;">在读&nbsp;1&nbsp;
	                门<!-- 门 -->！</font>
	            <font size="2px"> 计划外<!-- 计划外 -->：     通过<!-- 通过 -->&nbsp;3&nbsp;
	                门<!-- 门 -->，</font><font size="2px" style="color: red;">
	            未通过<!-- 未通过 --> &nbsp;1&nbsp;门<!-- 门 --></font>
	        </div>
        <div class="position_r">
            <ul class="treeview" id="ul220101011011start">
        </div>
    </form>
<br/>
<div class="row col-sm-12">
    <div class="alert alert-dismissible red" role="alert">
        <span class=""></span>
        <span class=""></span>
        <strong>提示<!-- 提示 -->:</strong>&nbsp;此页面信息仅做学业修读情况参考<!-- 此页面信息仅做学业修读情况参考 -->。
    </div>
</div>
<div class="row col-sm-12 row1">
    <div class="col-md-1 col-sm-6">
        <div class="form-group">
            <label for="" class="col-sm-2 control-label " style="margin-top:0px">
                <div class="png_ico_tjxk tjxk4" title="已修"></div>
            </label>
            <div class="col-sm-10">
                <div style="font-size:12px;">已修<!-- 已修 --></div>
            </div>
        </div>
    </div>
    <div class="col-md-1 col-sm-6">
        <div class="form-group">
            <label for="" class="col-sm-2 control-label " style="margin-top:0px">
                <div class="png_ico_tjxk tjxk1" title="在修"></div>
            </label>
            <div class="col-sm-10">
                <div style="font-size:12px;">在修<!-- 在修 --></div>
            </div>
        </div>
    </div>
    <div class="col-md-1 col-sm-6">
        <div class="form-group">
            <label for="" class="col-sm-2 control-label " style="margin-top:0px">
                <div class="png_ico_tjxk tjxk3" title="未修"></div>
            </label>
            <div class="col-sm-10">
                <div style="font-size:12px;">未修<!-- 未修 --></div>
            </div>
        </div>
    </div>
    <div class="col-md-1 col-sm-6">
        <div class="form-group">
            <label for="" class="col-sm-2 control-label " style="margin-top:0px">
                <div class="png_ico_tjxk tjxk2" title="未过"></div>
            </label>
            <div class="col-sm-10">
                <div style="font-size:12px;">未过<!-- 未过 --></div>
            </div>
        </div>
    </div>
    <div class="col-md-1 col-sm-6">
        <div class="form-group">
            <label for="" class="col-sm-2 control-label " style="margin-top:0px">
                <div class="zt3 zt33" title="学分已满"></div>
            </label>
            <div class="col-sm-10">
                <div style="font-size:12px;">学分已满<!-- 学分已满 --></div>
            </div>
        </div>
    </div>
    <div class="col-md-1 col-sm-6">
        <div class="form-group">
            <label for="" class="col-sm-2 control-label " style="margin-top:0px">
                <div class="zt3 zt32" title="学分超出"></div>
            </label>
            <div class="col-sm-10">
                <div style="font-size:12px;">学分超出<!-- 学分超出 --></div>
            </div>
        </div>
    </div>
    <div class="col-md-1 col-sm-6">
        <div class="form-group">
            <label for="" class="col-sm-2 control-label " style="margin-top:0px">
                <div class="zt3 zt31" title="学分未满"></div>
            </label>
            <div class="col-sm-10">
                <div style="font-size:12px;">学分未满<!-- 学分未满 --></div>
            </div>
        </div>
    </div>
    <div class="col-md-1 col-sm-6">
        <div class="form-group">
            <label for="" class="col-sm-2 control-label " style="margin-top:0px">
                <div class="png_ico_tjxk tjxk5" title="课程替代"></div>
            </label>
            <div class="col-sm-10">
                <div style="font-size:12px;">课程替代<!-- 课程替代 --></div>
            </div>
        </div>
    </div>
    <div class="col-md-1 col-sm-6">
        <div class="form-group">
            <label for="" class="col-sm-2 control-label " style="margin-top:0px">
                <div class="jdwg" title="节点未过"></div>
            </label>
            <div class="col-sm-10">
                <div style="font-size:12px;">节点未过<!-- 节点未过--></div>
            </div>
        </div>
    </div>
</div>
</body>


</html>
`
	GPA(htmlText)
}
