package demo

import "gorm.io/plugin/soft_delete"

// ===== Model 定义 =====
// 演示：实现 IDbModel 接口（GetID），定义表名（TableName），支持多 Model

// AccountModel 账户模型
type AccountModel struct {
	ID          int                   `json:"id" gorm:"primaryKey"`
	AccountName string                `json:"account_name"`
	Password    string                `json:"password"`
	Email       string                `json:"email"`
	Balance     float64               `json:"balance"`
	CreatedAt   int64                 `json:"created_at"`
	UpdatedAt   int64                 `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"deleted_at"`
}

func (a AccountModel) TableName() string { return "account" }

func (a AccountModel) GetID() int { return a.ID }

// OrderModel 订单模型 — 演示多 Model 场景
type OrderModel struct {
	ID        int     `json:"id" gorm:"primaryKey"`
	AccountID int     `json:"account_id"`
	Product   string  `json:"product"`
	Amount    float64 `json:"amount"`
	CreatedAt int64   `json:"created_at"`
}

func (o OrderModel) TableName() string { return "order" }

func (o OrderModel) GetID() int { return o.ID }
