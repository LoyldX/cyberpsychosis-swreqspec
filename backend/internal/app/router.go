// Package app ประกอบ Gin application แยกจากการเปิดเซิร์ฟเวอร์ เพื่อให้ test import ได้โดยไม่ listen พอร์ต (CON-TECH-01).
package app

import "github.com/gin-gonic/gin"

// NewRouter สร้าง *gin.Engine ของ API แล้วให้แต่ละ registrar ลง route ของตน (CON-TECH-01); เรียกโดยไม่ส่งอะไรได้ ผลคือยังไม่มี endpoint.
func NewRouter(registrars ...func(gin.IRouter)) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	for _, register := range registrars {
		register(r)
	}
	return r
}
