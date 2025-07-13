package config

const (
	LoginIndex = "/xtgl/login_slogin.html"        // 登录页
	INDEX_URL  = "/xtgl/index_initMenu.html"      // 首页
	CAPTCHA    = "/zfcaptchaLogin"                // 登录滑块验证码
	KCAPTCHA   = "/kaptcha"                       // 登录图形验证码
	PublicKey  = "/xtgl/login_getPublicKey.html"  // 登录公钥获取
	LOGOUT     = "/logout"                        // 登出
	LOGOUT2    = "/xtgl/login_logoutAccount.html" // 正方9.0登录会有

	Notifications = "/xtgl/index_cxDbsy.html"
	StudentPhoto  = "/xtgl/photo_cxXszp4.html" // 学生大头照

	INFO            = "/xsxxxggl/xsgrxxwh_cxXsgrxx.html"   // 个人信息获取
	SelectedCourses = "/xsxxxggl/xsxxwh_cxXsxkxx.html"     // 已选课程
	SCORE           = "/cjcx/cjcx_cxDgXscj.html"           // 查询成绩
	PersonalInfo    = "/cjcx/cjcx_cxXsgrcj.html"           // 可以查分，查个人信息
	ClassSchedule   = "/kbdy/bjkbdy_cxBjKb.html"           // 班级课表
	SchedulePolicy  = "/kbdy/bjkbdy_cxXnxqsfkz.html"       // 获取课表pdf
	ScheduleFile    = "/kbcx/xskbcx_cxXsShcPdf.html"       // 获取课表pdf
	SCHEDULE        = "/kbcx/xskbcx_cxXsKb.html"           // 课表信息
	Academia        = "/xsxy/xsxyqk_cxJxzxjhxfyqKcxx.html" // 学生生涯
	AcademiaIndex   = "/xsxy/xsxyqk_cxXsxyqkIndex.html"    // GPA

	// 选课
	CHOOSE_COURSE_INDEX            = "/xsxk/zzxkyzb_cxZzxkYzbIndex.html"          // 选课基础页面
	CHOOSE_COURSE_List_pre         = "/xsxk/zzxkyzb_cxZzxkYzbDisplay.html"        // 搜索课程前的参数准备，浏览器有缓存的话貌似会第一个请求这个
	CHOOSE_COURSE_courseList       = "/xsxk/zzxkyzb_cxZzxkYzbPartDisplay.html"    // 搜索课程接口
	CHOOSE_COURSE_courseDetail     = "/xsxk/zzxkyzbjk_cxJxbWithKchZzxkYzb.html"   // 查询课程号对应的详细信息
	CHOOSE_COURSE_chooseCourse_pre = "/xtsxk/zzxkyzb_cxXkTitleMsg.html"           // 选课前的一个请求会返回 flag=1
	CHOOSE_COURSE_chooseCourse     = "/xsxk/zzxkyzbjk_xkBcZyZzxkYzb.html"         // 选课
	CHOOSE_COURSE_SelectedList     = "/xsxk/zzxkyzb_cxZzxkYzbChoosedDisplay.html" // 查询已选课程接口
	CHOOSE_COURSE_quitCourse       = "/xsxk/zzxkyzb_tuikBcZzxkYzb.html"           // 退课
	ZY                             = "/xsxk/zzxkyzb_xkBcZypxZzxkYzb.html"         // 志愿 post
	CHOOSE_COURSE_new2             = "/xsxk/zzxkyzb_cxXsXktsxx.html"              // ?gnmkdm=N253512，不知道是啥会返回1
	CourseCategory                 = "/jxjhgl/common_cxKcJbxx.html"               // 根据课程号获取类别

	showDialog = "/xkgl/common_cxSbTitle.html"       // 上课时间冲突且可查看冲突 data:{"msg":msg/*,"xkxnm":$("#xkxnm").val(),"xkxqm":$("#xkxqm").val(),"jxb_ids":jxb_arr*/},
	notDefine  = "/xkgl/common_cxCtjxbListPage.html" // 冲突教学班 data:{"xnm":$("#xkxnm").val(),"xqm":$("#xkxqm").val(),"kch_id":kch_id,"jxb_ids":do_jxb_id}

	Evaluations  = "/zjjspxgl/xspxzjjs_cxXspxzjjsIndex.html" // 教学评价页面
	TeacherPhoto = "/photo/photo_cxJzgzp2.html"              // 教学评价教师大头照
)
