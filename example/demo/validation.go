package demo

import entityApi "github.com/hecc-blot/framework/entity/api"

// ===== 请求参数与校验 =====
// 演示：binding tag 自动校验（required/min/max/email）、自定义错误信息 GetMessages()

// AddAccountRequest 新增账户 — 展示多种校验规则
type AddAccountRequest struct {
	AccountName string `json:"account_name" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	Email       string `json:"email" binding:"required,email"`
	Age         int    `json:"age" binding:"min=1,max=150"`
}

func (a AddAccountRequest) GetMessages() entityApi.Messages {
	return entityApi.Messages{
		"AccountName.required": "用户名不能为空",
		"Password.required":    "密码不能为空",
		"Password.min":         "密码长度不能少于6位",
		"Email.required":       "邮箱不能为空",
		"Email.email":          "邮箱格式不正确",
		"Age.min":              "年龄不能小于1",
		"Age.max":              "年龄不能大于150",
	}
}
