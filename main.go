/*
 * @Author: Vincent Young
 * @Date: 2023-02-07 03:35:39
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2024-09-08 21:56:09
 * @FilePath: /WeiboSearcher/main.go
 * @Telegram: https://t.me/missuo
 *
 * Copyright © 2023 by Vincent, All Rights Reserved.
 */

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

type User struct {
	Uid    string
	Mobile string
}

type Set struct {
	ClickhouseConf    ClickhouseConf    `yaml:"clickhouse"`
	WeiboSearcherConf WeiboSearcherConf `yaml:"weiboSearcher"`
}

type ClickhouseConf struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Dbname   string `yaml:"dbname"`
}

type WeiboSearcherConf struct {
	ListenAddress string `yaml:"listenAddress"`
	ListenPort    string `yaml:"listenPort"`
}

func getConfigFromEnvOrFile() *Set {
	var set Set

	// Try to get configuration from environment variables first
	if dbHost := os.Getenv("CLICKHOUSE_HOST"); dbHost != "" {
		set.ClickhouseConf.Host = dbHost
		set.ClickhouseConf.Port = getEnvOrDefault("CLICKHOUSE_PORT", "9009")
		set.ClickhouseConf.Username = getEnvOrDefault("CLICKHOUSE_USERNAME", "default")
		set.ClickhouseConf.Password = getEnvOrDefault("CLICKHOUSE_PASSWORD", "")
		set.ClickhouseConf.Dbname = getEnvOrDefault("CLICKHOUSE_DBNAME", "default")
		set.WeiboSearcherConf.ListenAddress = getEnvOrDefault("LISTEN_ADDRESS", "0.0.0.0")
		set.WeiboSearcherConf.ListenPort = getEnvOrDefault("LISTEN_PORT", "8080")
		return &set
	}

	// Fall back to config file if environment variables are not set
	yamlFile, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		fmt.Println("Error reading config file:", err.Error())
		// Return default values if both env vars and config file fail
		set.ClickhouseConf.Host = "localhost"
		set.ClickhouseConf.Port = "9009"
		set.ClickhouseConf.Username = "default"
		set.ClickhouseConf.Password = ""
		set.ClickhouseConf.Dbname = "default"
		set.WeiboSearcherConf.ListenAddress = "0.0.0.0"
		set.WeiboSearcherConf.ListenPort = "8080"
		return &set
	}
	err = yaml.Unmarshal(yamlFile, &set)
	if err != nil {
		fmt.Println("Error parsing config file:", err.Error())
	}
	return &set
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Get Configuration
	set := getConfigFromEnvOrFile()
	var clickhouseConf = set.ClickhouseConf
	var appConf = set.WeiboSearcherConf
	dbHost := clickhouseConf.Host
	dbPort := clickhouseConf.Port
	dbUsername := clickhouseConf.Username
	dbPassword := clickhouseConf.Password
	dbName := clickhouseConf.Dbname
	appAddress := appConf.ListenAddress
	appPort := appConf.ListenPort

	// Connect Clickhouse
	dsn := "clickhouse://" + dbUsername + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName
	db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())
	r.GET("/", func(c *gin.Context) {
		// Index Page
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "This is Weibo SGK. Made by Vincent.",
			"usage":   "GET/POST to /wb with parameter u",
		})

	})

	r.Any("/wb", func(c *gin.Context) {
		re := regexp.MustCompile(`\d+`)
		u := c.Query("u")
		key := re.FindString(u)
		var result User
		if len(key) == 10 {
			// Input Weibo Uid
			db.Raw("SELECT * FROM wb WHERE uid = ? LIMIT 1", key).Scan(&result)
		} else if len(key) == 11 {
			// Input User Mobile Number
			db.Raw("SELECT * FROM wbm WHERE mobile = ? LIMIT 1", key).Scan(&result)
		} else {
			// Bad Parameters
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "Bad Parameters",
			})
			return
		}
		// No Results
		if result.Uid == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "Data Not Found",
			})

		} else {
			c.JSON(http.StatusOK, gin.H{
				"code":   http.StatusOK,
				"uid":    result.Uid,
				"mobile": result.Mobile,
			})
		}

	})

	// Catch-all route to handle undefined paths
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "Path not found",
		})
	})

	r.Run(appAddress + ":" + appPort)
}
