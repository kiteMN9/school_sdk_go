package config

const (
	LoginIndex = "xtgl/login_slogin.html"        // 登录页
	INDEX_URL  = "xtgl/index_initMenu.html"      // 首页
	CAPTCHA    = "zfcaptchaLogin"                // 登录滑块验证码
	KAPTCHA    = "kaptcha"                       // 登录图形验证码
	PublicKey  = "xtgl/login_getPublicKey.html"  // 登录公钥获取
	LOGOUT     = "logout"                        // 登出
	LOGOUT2    = "xtgl/login_logoutAccount.html" // 正方9.0登录会有
	Language   = "xtgl/init_changeLocal.html"
)

const (
	Notifications = "/xtgl/index_cxDbsy.html" // 一些通知
	StudentPhoto  = "/xtgl/photo_cxXszp.html" // null,4学生大头照; 2 null, 3无上传照片权限
	StudentName   = "/xtgl/index_cxYhxxIndex.html"

	INFO            = "/xsxxxggl/xsgrxxwh_cxXsgrxx.html" // 个人信息获取 HTML
	InfoJson        = "/xsxxxggl/xsxxwh_cxCkDgxsxx.html"
	ExtraInfoURL    = "/xszbbgl/xszbbgl_cxXszbbsqIndex.html?doType=details&gnmkdm=N106005" // GET学生证补办，post
	SelectedCourses = "/xsxxxggl/xsxxwh_cxXsxkxx.html"                                     // 已选课程 json
	SCORE           = "/cjcx/cjcx_cxDgXscj.html"                                           // 查询成绩 or 不带参数是挂科json
	PersonalInfo    = "/cjcx/cjcx_cxXsgrcj.html"                                           // 可以查分，查个人信息，不带参数是挂科json
	ClassSchedule   = "/kbdy/bjkbdy_cxBjKb.html"                                           // 班级课表
	SchedulePolicy  = "/kbdy/bjkbdy_cxXnxqsfkz.html"                                       // 获取课表pdf
	ScheduleFile    = "/kbcx/xskbcx_cxXsShcPdf.html"                                       // 获取课表pdf
	Schedule        = "/kbcx/xskbcx_cxXsKb.html"                                           // 课表信息
	LittleSchedule  = "kbcx/xskbcx_cxXskbSimpleIndex.html"                                 // 选课里的小课表
	Academia        = "/xsxy/xsxyqk_cxJxzxjhxfyqKcxx.html"                                 // 学生生涯 post
	AcademiaIndex   = "/xsxy/xsxyqk_cxXsxyqkIndex.html"                                    // GPA

	showDialog = "/xkgl/common_cxSbTitle.html"       // 上课时间冲突且可查看冲突 data:{"msg":msg/*,"xkxnm":$("#xkxnm").val(),"xkxqm":$("#xkxqm").val(),"jxb_ids":jxb_arr*/},
	Conflict   = "/xkgl/common_cxCtjxbListPage.html" // 冲突教学班 data:{"xnm":$("#xkxnm").val(),"xqm":$("#xkxqm").val(),"kch_id":kch_id,"jxb_ids":do_jxb_id}

	Evaluations  = "/zjjspxgl/xspxzjjs_cxXspxzjjsIndex.html?gnmkdm=N408125" // 教学评价页面
	TeacherPhoto = "/photo/photo_cxJzgzp2.html"                             // 教学评价教师大头照
)

// 选课
const (
	ChooseCourseIndex        = "/xsxk/zzxkyzb_cxZzxkYzbIndex.html"        // 选课基础页面
	ChooseCourseListPre      = "/xsxk/zzxkyzb_cxZzxkYzbDisplay.html"      // 搜索课程前的参数准备，浏览器有缓存的话貌似会第一个请求这个
	ChooseCourseCourseList   = "/xsxk/zzxkyzb_cxZzxkYzbPartDisplay.html"  // 搜索课程接口
	ChooseCourseCourseDetail = "/xsxk/zzxkyzbjk_cxJxbWithKchZzxkYzb.html" // 查询课程号对应的详细信息
	ChooseCourse             = "/xsxk/zzxkyzbjk_xkBcZyZzxkYzb.html"       // 发送选课
	// chooseCoursesQuickly
	//xsxk/zzxkyzb_xkZzxkyzbQuickly.html",{"xkkz_id":$("#xkkz_id").val(

	CourseSelectedList = "/xsxk/zzxkyzb_cxZzxkYzbChoosedDisplay.html" // 查询已选课程接口
	QuitCourse         = "/xsxk/zzxkyzb_tuikBcZzxkYzb.html"           // 退课
	// xsxk/zzxkyzb_tuikBcXkyix.html
	//ZY                             = "/xsxk/zzxkyzb_xkBcZypxZzxkYzb.html" // 志愿 post "no-permission"
	//CHOOSE_COURSE_new2             = "/xsxk/zzxkyzb_cxXsXktsxx.html"      // ?gnmkdm=N253512，不知道是啥会返回1
	//CourseCategory                 = "/jxjhgl/common_cxKcJbxx.html"       // 根据课程号获取类别
	//CHOOSE_COURSE_chooseCourse_pre = "/xtsxk/zzxkyzb_cxXkTitleMsg.html"   // 发送选课前的一个请求会返回 flag=1 没什么用
	// xsxk/zzxkyzb_xkZyZzxkYzbZjxb.html
)

const (
// zfn_api
// url_view     = "bysxxcx/xscjzbdy_dyXscjzbView.html"
// url_window   = "bysxxcx/xscjzbdy_dyCjdyszxView.html"
// url_policy   = "xtgl/bysxxcx/xscjzbdy_cxXsCount.html"
// url_filetype = "bysxxcx/xscjzbdy_cxGswjlx.html"
// url_common   = "common/common_cxJwxtxx.html"
// url_file     = "bysxxcx/xscjzbdy_dyList.html"
// url_progress = "xtgl/progress_cxProgressStatus.html"
)
