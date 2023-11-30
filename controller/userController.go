package controller

import (
	"github.com/goccy/go-json"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"dning.com/pro02/common"
	"dning.com/pro02/dto"
	"dning.com/pro02/model"
	"dning.com/pro02/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Register 注册
func Register(c *gin.Context) {
	//db := common.GetDB()
	rds := common.GetRDS()

	requestStudent := model.Student{}
	c.Bind(&requestStudent)
	// 获取参数
	name := requestStudent.Name
	sid := requestStudent.SID
	password := requestStudent.Password

	// 数据验证
	if len(sid) != 11 {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "学号必须为11位")
		return
	}

	if len(password) < 6 {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "密码不能少于6位")
		return
	}
	//// 如果名称没传，随机分配10位字符串
	//if len(name) == 0 {
	//	name = util.RandomName(10)
	//}

	log.Println(name, sid, password)

	// 判断学号是否存在
	if isSIDExist(rds, sid) {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "用户已经存在")
		return
	}
	// 创建用户
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "密码加密错误")
		return
	}

	newStudent := model.Student{
		Name:         name,
		SID:          sid,
		Password:     string(hashPassword),
		NormalCount:  "0",
		LeaveCount:   "0",
		LateCount:    "0",
		EarlyCount:   "0",
		AbsenceCount: "0",
	}
	//db.Create(&newUser)
	reply, err := rds.Do("HMSET", redis.Args{}.Add("student:"+newStudent.SID).AddFlat(&newStudent)...)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		log.Printf("创建新用户出错, error: %v", err)
		return
	}
	log.Printf("创建新用户状态：%v", reply)

	// 发放token
	token, err := common.ReleaseToken(newStudent)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		log.Printf("token generate error: %v", err)
		return
	}

	response.Success(c, gin.H{"token": token}, "注册成功")

}

// Login 登录
func Login(c *gin.Context) {
	//db := common.GetDB()
	rds := common.GetRDS()

	// 获取参数
	loginStudent := model.Student{}
	c.Bind(&loginStudent)
	// 获取参数
	sid := loginStudent.SID
	password := loginStudent.Password

	// 验证数据
	if len(sid) != 11 {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "学号必须为11位")
		return
	}

	if len(password) < 6 {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "密码不能少于6位")
		return
	}

	// 判断学号是否存在
	if !isSIDExist(rds, sid) {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "用户不存在")
		return
	}

	// 获取用户密码并判断密码是否正确
	var student model.Student
	stringMap, err2 := redis.StringMap(rds.Do("HGETALL", redis.Args{}.Add("student:"+sid)...))
	if err2 != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		log.Printf("获取用户密码时连接redis出错，err: " + err2.Error())
	}
	student.Name = stringMap["name"]
	student.SID = stringMap["sid"]
	student.Password = stringMap["password"]

	if err := bcrypt.CompareHashAndPassword([]byte(student.Password), []byte(password)); err != nil {
		response.Response(c, http.StatusBadRequest, 400, nil, "密码错误")
		return
	}

	// 发放token
	token, err := common.ReleaseToken(student)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		log.Printf("token generate error: %v", err)
		return
	}

	response.Success(c, gin.H{"token": token}, "登录成功")
}

// Info 获取user信息（有中间件）
func Info(c *gin.Context) {
	student, _ := c.Get("student")
	//c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"user": dto.ToUserDto(user.(model.User))}})
	response.Success(c, gin.H{"code": 200, "data": gin.H{"student": dto.ToStudentDto(student.(model.Student))}}, "")
}

// Leave 处理请假情况
func Leave(c *gin.Context) {
	rds := common.GetRDS()

	// 获取参数
	studentInfo := model.StudentLeave{}
	c.Bind(&studentInfo)

	// 验证数据
	// 判断学号是否存在
	if !isSIDExist(rds, studentInfo.SID) {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "学生不存在")
		log.Printf("处理请假情况时没找到学号，SID:%v", studentInfo.SID)
		return
	}

	// 查找当天是否已经请过假
	length, _ := redis.Int(rds.Do("LLEN", redis.Args{}.Add("studentLeave:"+studentInfo.SID+":Date")...))
	for i := 0; i < length; i++ {
		date, _ := redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+studentInfo.SID+":Date").Add(i)...))
		if date == studentInfo.Date {
			response.Response(c, http.StatusUnprocessableEntity, 422, nil, "当天已经请过假")
			log.Printf("请假的那天已经请过假了！")
			return
		}
	}

	// 遍历studentInfo
	typeInfo := reflect.TypeOf(studentInfo)
	valueInfo := reflect.ValueOf(studentInfo)
	for i := 0; i < valueInfo.NumField(); i++ {
		key := typeInfo.Field(i).Name
		val := valueInfo.Field(i).Interface()
		reply, err := rds.Do("RPUSH", redis.Args{}.Add("studentLeave:"+studentInfo.SID+":"+key).Add(val)...)
		if err != nil {
			log.Printf("向redis写入请假信息时出错，err: " + err.Error())
			response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		}
		log.Printf("向redis写入请假信息状态：%v", reply)
	}

	response.Success(c, gin.H{"SID": studentInfo.SID}, "请假发送成功")
}

// DeleteLeave 撤回请假
func DeleteLeave(c *gin.Context) {
	rds := common.GetRDS()
	index := c.Query("index")
	sid := c.Query("SID")
	if !isSIDExist(rds, sid) {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "学生不存在")
		log.Printf("撤回请假时没找到学号，SID:%v", sid)
		return
	}

	// 删除名字
	rds.Do("LSET", redis.Args{}.Add("studentLeave:"+sid+":Name").Add(index).Add("del")...)
	rds.Do("LREM", redis.Args{}.Add("studentLeave:"+sid+":Name").Add(1).Add("del")...)
	// 删除班级
	rds.Do("LSET", redis.Args{}.Add("studentLeave:"+sid+":SClass").Add(index).Add("del")...)
	rds.Do("LREM", redis.Args{}.Add("studentLeave:"+sid+":SClass").Add(1).Add("del")...)
	// 删除日期
	rds.Do("LSET", redis.Args{}.Add("studentLeave:"+sid+":Date").Add(index).Add("del")...)
	rds.Do("LREM", redis.Args{}.Add("studentLeave:"+sid+":Date").Add(1).Add("del")...)
	// 删除理由
	rds.Do("LSET", redis.Args{}.Add("studentLeave:"+sid+":Reason").Add(index).Add("del")...)
	rds.Do("LREM", redis.Args{}.Add("studentLeave:"+sid+":Reason").Add(1).Add("del")...)
	// 删除是否批假
	rds.Do("LSET", redis.Args{}.Add("studentLeave:"+sid+":IsApproved").Add(index).Add("del")...)
	rds.Do("LREM", redis.Args{}.Add("studentLeave:"+sid+":IsApproved").Add(1).Add("del")...)

	response.Success(c, gin.H{"SID": sid}, "撤回请假成功！")
}

// LeaveInfo 获取请假情况
func LeaveInfo(c *gin.Context) {
	rds := common.GetRDS()
	sid := c.Query("SID")

	if !isSIDExist(rds, sid) {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "学生不存在")
		log.Printf("获取请假情况时没找到学号，SID:%v", sid)
		return
	}
	studentLeaveInfo := model.StudentLeave{}
	studentLeaveInfo.SID = sid
	var studentLeaveInfos []model.StudentLeave

	length, _ := redis.Int(rds.Do("LLEN", redis.Args{}.Add("studentLeave:"+studentLeaveInfo.SID+":Name")...))
	for i := 0; i < length; i++ {
		studentLeaveInfo.Name, _ = redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+studentLeaveInfo.SID+":Name").Add(i)...))
		studentLeaveInfo.SClass, _ = redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+studentLeaveInfo.SID+":SClass").Add(i)...))
		studentLeaveInfo.Date, _ = redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+studentLeaveInfo.SID+":Date").Add(i)...))
		studentLeaveInfo.Reason, _ = redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+studentLeaveInfo.SID+":Reason").Add(i)...))
		studentLeaveInfo.IsApproved, _ = redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+studentLeaveInfo.SID+":IsApproved").Add(i)...))
		studentLeaveInfos = append(studentLeaveInfos, studentLeaveInfo)
	}
	jsonData, _ := json.Marshal(studentLeaveInfos)
	log.Printf("studentLeaveInfo: %v", string(jsonData))
	response.Success(c, gin.H{"studentLeaveInfos": string(jsonData)}, "获取请假信息成功！")
}

// TeacherRegister 教师注册
func TeacherRegister(c *gin.Context) {
	rds := common.GetRDS()

	requestTeacher := model.Teacher{}
	c.Bind(&requestTeacher)
	// 获取参数
	name := requestTeacher.Name
	tid := requestTeacher.TID
	password := requestTeacher.Password

	// 数据验证
	if len(tid) != 6 {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "教工号必须为6位")
		return
	}

	if len(password) < 6 {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "密码不能少于6位")
		return
	}

	log.Println(name, tid, password)

	// 判断学号是否存在
	if isSIDExist2(rds, tid) {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "用户已经存在")
		return
	}
	// 创建用户
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "密码加密错误")
		return
	}

	newTeacher := model.Teacher{
		Name:     name,
		TID:      tid,
		Password: string(hashPassword),
	}
	//db.Create(&newUser)
	reply, err := rds.Do("HMSET", redis.Args{}.Add("teacher:"+newTeacher.TID).AddFlat(&newTeacher)...)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		log.Printf("创建新用户出错, error: %v", err)
		return
	}
	log.Printf("创建新用户状态：%v", reply)

	// 发放token
	token, err := common.ReleaseToken2(newTeacher)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		log.Printf("token generate error: %v", err)
		return
	}

	response.Success(c, gin.H{"token": token}, "注册成功")
}

// TeacherLogin 教师登录
func TeacherLogin(c *gin.Context) {
	rds := common.GetRDS()

	// 获取参数
	loginTeacher := model.Teacher{}
	c.Bind(&loginTeacher)
	// 获取参数
	tid := loginTeacher.TID
	password := loginTeacher.Password

	// 验证数据
	if len(tid) != 6 {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "教工号必须为6位")
		return
	}

	if len(password) < 6 {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "密码不能少于6位")
		return
	}

	// 判断学号是否存在
	if !isSIDExist2(rds, tid) {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "用户不存在")
		return
	}

	// 获取用户密码并判断密码是否正确
	var teacher model.Teacher
	stringMap, err2 := redis.StringMap(rds.Do("HGETALL", redis.Args{}.Add("teacher:"+tid)...))
	if err2 != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		log.Printf("获取用户密码时连接redis出错，err: " + err2.Error())
	}
	teacher.Name = stringMap["name"]
	teacher.TID = stringMap["tid"]
	teacher.Password = stringMap["password"]

	if err := bcrypt.CompareHashAndPassword([]byte(teacher.Password), []byte(password)); err != nil {
		response.Response(c, http.StatusBadRequest, 400, nil, "密码错误")
		return
	}

	// 发放token
	token, err := common.ReleaseToken2(teacher)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		log.Printf("token generate error: %v", err)
		return
	}

	response.Success(c, gin.H{"token": token}, "登录成功")
}

// TeacherInfo 教师信息
func TeacherInfo(c *gin.Context) {
	teacher, _ := c.Get("teacher")
	//c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"user": dto.ToUserDto(user.(model.User))}})
	response.Success(c, gin.H{"code": 200, "data": gin.H{"teacher": dto.ToTeacherDto(teacher.(model.Teacher))}}, "")
}

// StudentInfo 获取所有学生信息
func StudentInfo(c *gin.Context) {
	rds := common.GetRDS()

	keys, err := redis.Strings(rds.Do("KEYS", "student:*"))
	if err != nil {
		log.Printf("获取全部学生信息时查询redis发生错误，err: %v", err)
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常！")
		return
	}
	var studentInfos []dto.StudentDto
	for _, key := range keys {
		sid := key[8:]

		student := dto.StudentDto{}
		stringMap, _ := redis.StringMap(rds.Do("HGETALL", redis.Args{}.Add("student:"+sid)...))
		student.Name = stringMap["name"]
		student.SID = stringMap["sid"]
		student.NormalCount = stringMap["normalCount"]
		student.LeaveCount = stringMap["leaveCount"]
		student.LateCount = stringMap["lateCount"]
		student.EarlyCount = stringMap["earlyCount"]
		student.AbsenceCount = stringMap["absenceCount"]
		studentInfos = append(studentInfos, student)
	}
	jsonData, _ := json.Marshal(studentInfos)
	response.Success(c, gin.H{"studentInfos": string(jsonData)}, "获取学生信息成功！")
	log.Printf("%v", string(jsonData))
}

// StudentLeaveInfo 获取所有学生请假信息
func StudentLeaveInfo(c *gin.Context) {
	rds := common.GetRDS()
	keys, err := redis.Strings(rds.Do("KEYS", "student:*"))
	if err != nil {
		log.Printf("获取全部学生请假信息时查询redis发生错误，err: %v", err)
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常！")
		return
	}
	var studentLeaveInfos []model.StudentLeave
	for _, key := range keys {
		sid := key[8:]

		length, _ := redis.Int(rds.Do("LLEN", redis.Args{}.Add("studentLeave:"+sid+":Name")...))
		if length == 0 {
			continue
		}
		for i := 0; i < length; i++ {
			isApproved, _ := redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+sid+":IsApproved").Add(i)...))
			if isApproved != "0" {
				continue
			}
			studentLeaveInfo := model.StudentLeave{}
			studentLeaveInfo.SID = sid
			studentLeaveInfo.Name, _ = redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+sid+":Name").Add(i)...))
			studentLeaveInfo.SClass, _ = redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+sid+":SClass").Add(i)...))
			studentLeaveInfo.Date, _ = redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+sid+":Date").Add(i)...))
			studentLeaveInfo.Reason, _ = redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+sid+":Reason").Add(i)...))
			studentLeaveInfos = append(studentLeaveInfos, studentLeaveInfo)
		}
	}
	jsonData, _ := json.Marshal(studentLeaveInfos)
	response.Success(c, gin.H{"studentLeaveInfos": string(jsonData)}, "获取请假信息成功！")
	log.Printf("%v", string(jsonData))
	//fmt.Println(string(jsonData))
}

// ApproveLeave 批假
func ApproveLeave(c *gin.Context) {
	rds := common.GetRDS()
	sid := c.Query("SID")
	date := c.Query("date")
	approveValue := c.Query("approveValue")

	// 查询数据
	length, _ := redis.Int(rds.Do("LLEN", redis.Args{}.Add("studentLeave:"+sid+":Date")...))
	index := 0
	for i := 0; i < length; i++ {
		dateInfo, _ := redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+sid+":Date").Add(i)...))
		if dateInfo == date {
			index = i
			break
		}
	}

	// 将当天的isApproved置为approveValue
	_, err := rds.Do("LSET", redis.Args{}.Add("studentLeave:"+sid+":IsApproved").Add(index).Add(approveValue)...)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		log.Printf("批假时设置isApproved发生错误，err:%v", err)
		return
	}

	// 将请假次数+1
	if approveValue == "1" {
		countStr, _ := redis.String(rds.Do("HGET", redis.Args{}.Add("student:"+sid).Add("leaveCount")...))
		count, _ := strconv.Atoi(countStr)
		count++
		countStr = strconv.Itoa(count)
		rds.Do("HSET", redis.Args{}.Add("student:"+sid).Add("leaveCount").Add(countStr)...)
	}

	response.Success(c, nil, "批假成功！")

}

// SetClassTime 教师设置上课时间
func SetClassTime(c *gin.Context) {
	rds := common.GetRDS()
	classDate := c.Query("classDate")
	classBeginTime := c.Query("classBeginTime")
	classOverTime := c.Query("classOverTime")

	if classDate == "" || classBeginTime == "" || classOverTime == "" {
		response.Response(c, http.StatusBadRequest, 400, nil, "get请求错误")
		log.Println("没收到上课时间")
		return
	}

	// 将上课时间存储到redis中
	classBeginTime = classBeginTime + ":00"
	classOverTime = classOverTime + ":00"
	classTimeStr := classBeginTime + "-" + classOverTime
	_, err := rds.Do("HSET", redis.Args{}.Add("classTime").Add(classDate).Add(classTimeStr)...)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常！")
		log.Printf("将上课时间存储到redis中时出错，err：%v", err)
		return
	}

	keys, err := redis.Strings(rds.Do("KEYS", "student:*"))
	if err != nil {
		log.Printf("获取全部学生签到信息时查询redis发生错误，err: %v", err)
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常！")
		return
	}

	for _, key := range keys {
		sid := key[8:]
		rds.Do("HSET", redis.Args{}.Add("studentSign:"+sid+":IsSignIn").Add(classDate).Add("0")...)
		rds.Do("HSET", redis.Args{}.Add("studentSign:"+sid+":IsSignOut").Add(classDate).Add("0")...)
	}

	response.Success(c, nil, "设置上课时间成功！")
}

// SignIn 签到
func SignIn(c *gin.Context) {
	rds := common.GetRDS()
	sid := c.Query("SID")
	signInDate := c.Query("signInDate")
	signInTimeStr := c.Query("signInTime")

	if signInDate == "" || signInTimeStr == "" || sid == "" {
		response.Response(c, http.StatusBadRequest, 400, nil, "get请求错误")
		log.Println("没收到签到时间")
		return
	}

	classTimeStr, err := redis.String(rds.Do("HGET", redis.Args{}.Add("classTime").Add(signInDate)...))
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常！")
		log.Printf("获取上课时间时出错，err：%v", err)
		return
	}
	// 判断今天上不上课
	if classTimeStr == "" {
		response.Response(c, http.StatusBadRequest, 400, nil, "今天不上课！")
		log.Println("今天不上课！")
		return
	}
	classBeginTimeStr := strings.Split(classTimeStr, "-")[0]

	// 判断是否已经签到
	isSignIn, _ := redis.String(rds.Do("HGET", redis.Args{}.Add("studentSign:"+sid+":IsSignIn").Add(signInDate)...))
	if isSignIn == "1" {
		response.Response(c, http.StatusBadRequest, 400, nil, "你已经签到过了！")
		log.Println("用户已经签到过了！")
		return
	}

	// 字符串转换为时间
	layout := "15:04:05"
	signInTime, err := time.Parse(layout, signInTimeStr)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常！")
		log.Printf("时间字符串转换为时间时出错，err：%v", err)
		return
	}
	classBeginTime, _ := time.Parse(layout, classBeginTimeStr)

	// 将签到时间和上课时间进行比较
	if signInTime.After(classBeginTime) {
		// 迟到次数+1
		countStr, _ := redis.String(rds.Do("HGET", redis.Args{}.Add("student:"+sid).Add("lateCount")...))
		count, _ := strconv.Atoi(countStr)
		count++
		countStr = strconv.Itoa(count)
		rds.Do("HSET", redis.Args{}.Add("student:"+sid).Add("lateCount").Add(countStr)...)
	}
	rds.Do("HSET", redis.Args{}.Add("studentSign:"+sid+":IsSignIn").Add(signInDate).Add("1")...)

	response.Success(c, nil, "签到成功！")
}

// SignOut 签退
func SignOut(c *gin.Context) {
	rds := common.GetRDS()
	sid := c.Query("SID")
	signOutDate := c.Query("signOutDate")
	signOutTimeStr := c.Query("signOutTime")

	if signOutDate == "" || signOutTimeStr == "" {
		response.Response(c, http.StatusBadRequest, 400, nil, "get请求错误")
		log.Println("没收到签到时间")
		return
	}

	classTimeStr, err := redis.String(rds.Do("HGET", redis.Args{}.Add("classTime").Add(signOutDate)...))
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常！")
		log.Printf("获取上课时间时出错，err：%v", err)
		return
	}
	// 判断今天上不上课
	if classTimeStr == "" {
		response.Response(c, http.StatusBadRequest, 400, nil, "今天不上课！")
		log.Println("今天不上课！")
		return
	}
	classOverTimeStr := strings.Split(classTimeStr, "-")[1]

	// 判断是否已经签退
	isSignOut, _ := redis.String(rds.Do("HGET", redis.Args{}.Add("studentSign:"+sid+":IsSignOut").Add(signOutDate)...))
	if isSignOut == "1" {
		response.Response(c, http.StatusBadRequest, 400, nil, "你已经签退过了！")
		log.Println("用户已经签退过了！")
		return
	}

	// 字符串转换为时间
	layout := "15:04:05"
	signOutTime, err := time.Parse(layout, signOutTimeStr)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常！")
		log.Printf("时间字符串转换为时间时出错，err：%v", err)
		return
	}
	classOverTime, _ := time.Parse(layout, classOverTimeStr)

	// 将签到时间和上课时间进行比较
	if signOutTime.Before(classOverTime) {
		// 早退次数+1
		countStr, _ := redis.String(rds.Do("HGET", redis.Args{}.Add("student:"+sid).Add("earlyCount")...))
		count, _ := strconv.Atoi(countStr)
		count++
		countStr = strconv.Itoa(count)
		rds.Do("HSET", redis.Args{}.Add("student:"+sid).Add("earlyCount").Add(countStr)...)
	} else {
		// 正常上课次数+1
		countStr, _ := redis.String(rds.Do("HGET", redis.Args{}.Add("student:"+sid).Add("normalCount")...))
		count, _ := strconv.Atoi(countStr)
		count++
		countStr = strconv.Itoa(count)
		rds.Do("HSET", redis.Args{}.Add("student:"+sid).Add("normalCount").Add(countStr)...)
	}
	// 设置已经当天已经签退
	rds.Do("HSET", redis.Args{}.Add("studentSign:"+sid+":IsSignOut").Add(signOutDate).Add("1")...)

	response.Success(c, nil, "签退成功！")
}

// GetAbsence 获取妹签到的人
func GetAbsence(c *gin.Context) {
	rds := common.GetRDS()
	classDate := c.Query("classDate")

	keys, err := redis.Strings(rds.Do("KEYS", "student:*"))
	if err != nil {
		log.Printf("获取全部学生签到信息时查询redis发生错误，err: %v", err)
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常！")
		return
	}

	for _, key := range keys {
		sid := key[8:]

		isSignIn, _ := redis.String(rds.Do("HGET", redis.Args{}.Add("studentSign:"+sid+":IsSignIn").Add(classDate)...))
		if isSignIn != "1" {
			// 查看是否请假
			length, _ := redis.Int(rds.Do("LLEN", redis.Args{}.Add("studentLeave:"+sid+":Date")...))
			index := -1
			for i := 0; i < length; i++ {
				dateInfo, _ := redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+sid+":Date").Add(i)...))
				if dateInfo == classDate {
					index = i
					break
				}
			}
			// 没请假
			if index == -1 {
				// 将旷课次数+1
				countStr, _ := redis.String(rds.Do("HGET", redis.Args{}.Add("student:"+sid).Add("absenceCount")...))
				count, _ := strconv.Atoi(countStr)
				count++
				countStr = strconv.Itoa(count)
				rds.Do("HSET", redis.Args{}.Add("student:"+sid).Add("absenceCount").Add(countStr)...)
				continue
			}
			isApproved, _ := redis.String(rds.Do("LINDEX", redis.Args{}.Add("studentLeave:"+sid+":IsApproved").Add(index)...))
			if isApproved == "1" {
				continue
			}
			// 将旷课次数+1
			countStr, _ := redis.String(rds.Do("HGET", redis.Args{}.Add("student:"+sid).Add("absenceCount")...))
			count, _ := strconv.Atoi(countStr)
			count++
			countStr = strconv.Itoa(count)
			rds.Do("HSET", redis.Args{}.Add("student:"+sid).Add("absenceCount").Add(countStr)...)
		}
	}
	response.Success(c, nil, "查询所有妹签到的人成功！")
}

func isSIDExist(rds redis.Conn, sid string) bool {
	reply, err := rds.Do("HGET", redis.Args{}.Add("student:"+sid).Add("sid")...)
	if err != nil {
		panic("判断学号是否存在时查询学号出错, err: " + err.Error())
	}
	if reply == nil {
		return false
	}
	return true
}

func isSIDExist2(rds redis.Conn, tid string) bool {
	reply, err := rds.Do("HGET", redis.Args{}.Add("teacher:"+tid).Add("tid")...)
	if err != nil {
		panic("判断教工号是否存在时查询教工号出错, err: " + err.Error())
	}
	if reply == nil {
		return false
	}
	return true
}
