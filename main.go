/*
 * @Author: Vincent Young
 * @Date: 2023-02-07 03:35:39
 * @LastEditors: Vincent Young
 * @LastEditTime: 2023-02-07 04:54:58
 * @FilePath: /wb/main.go
 * @Telegram: https://t.me/missuo
 *
 * Copyright © 2023 by Vincent, All Rights Reserved.
 */

package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

type User struct {
	Uid    string
	Mobile string
}

func main() {
	// Connect Clickhouse
	dsn := "clickhouse://default:@192.168.36.134:29000/weibo"
	db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		// Index Page
		c.JSON(200, gin.H{
			"code":    200,
			"message": "This is Weibo SGK. Made by Vincent.",
			"usage":   "GET/POST to /wb with parameter u",
		})

	})

	r.Any("/wb", func(c *gin.Context) {
		key := c.Query("u")
		var result User
		if len(key) == 10 {
			// Input Weibo Uid
			db.Raw("SELECT * FROM wb WHERE uid = ? LIMIT 1", key).Scan(&result)
		} else if len(key) == 11 {
			// Input User Mobile Number
			db.Raw("SELECT * FROM wbm WHERE mobile = ? LIMIT 1", key).Scan(&result)
		} else {
			// Bad Parameters
			c.JSON(400, gin.H{
				"code":    400,
				"message": "Bad Parameters",
			})
			return
		}
		// No Results
		if result.Uid == "" {
			c.JSON(404, gin.H{
				"code":    404,
				"message": "Data Not Found",
			})

		} else {
			c.JSON(200, gin.H{
				"code":   200,
				"uid":    result.Uid,
				"mobile": result.Mobile,
			})
		}

	})
	r.Run(":11119")
}
