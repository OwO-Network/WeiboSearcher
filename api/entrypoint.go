/*
 * @Author: Vincent Yang
 * @Date: 2025-09-03 01:03:00
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2025-09-04 01:10:22
 * @FilePath: /WeiboSearcher/api/entrypoint.go
 * @Telegram: https://t.me/missuo
 * @GitHub: https://github.com/missuo
 *
 * Copyright © 2025 by Vincent, All Rights Reserved.
 */

package api

import (
	"log"
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
		log.Printf("Using environment variables for configuration")
		set.ClickhouseConf.Host = dbHost
		set.ClickhouseConf.Port = getEnvOrDefault("CLICKHOUSE_PORT", "9000")
		set.ClickhouseConf.Username = getEnvOrDefault("CLICKHOUSE_USERNAME", "default")
		set.ClickhouseConf.Password = getEnvOrDefault("CLICKHOUSE_PASSWORD", "")
		set.ClickhouseConf.Dbname = getEnvOrDefault("CLICKHOUSE_DBNAME", "default")
		set.WeiboSearcherConf.ListenAddress = getEnvOrDefault("LISTEN_ADDRESS", "0.0.0.0")
		set.WeiboSearcherConf.ListenPort = getEnvOrDefault("LISTEN_PORT", "8080")
		return &set
	}

	// Fall back to config file if environment variables are not set
	log.Printf("Environment variables not found, trying config file")
	yamlFile, err := os.ReadFile("./config.yml")
	if err != nil {
		log.Printf("Error reading config file: %v", err)
		log.Printf("Using default configuration values")
		// Return default values if both env vars and config file fail
		set.ClickhouseConf.Host = "localhost"
		set.ClickhouseConf.Port = "9000"
		set.ClickhouseConf.Username = "default"
		set.ClickhouseConf.Password = ""
		set.ClickhouseConf.Dbname = "default"
		set.WeiboSearcherConf.ListenAddress = "0.0.0.0"
		set.WeiboSearcherConf.ListenPort = "8080"
		return &set
	}
	err = yaml.Unmarshal(yamlFile, &set)
	if err != nil {
		log.Printf("Error parsing config file: %v", err)
	}
	log.Printf("Configuration loaded from config file")
	return &set
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var (
	app *gin.Engine
	db  *gorm.DB
)

func init() {
	// Get Configuration
	set := getConfigFromEnvOrFile()
	var clickhouseConf = set.ClickhouseConf
	dbHost := clickhouseConf.Host
	dbPort := clickhouseConf.Port
	dbUsername := clickhouseConf.Username
	dbPassword := clickhouseConf.Password
	dbName := clickhouseConf.Dbname

	// Connect Clickhouse
	dsn := "clickhouse://" + dbUsername + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName
	var err error
	db, err = gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		log.Printf("DSN: %s", dsn)
		// Don't panic in serverless environment, let the function handle errors gracefully
		db = nil
	}

	gin.SetMode(gin.ReleaseMode)
	app = gin.Default()
	app.Use(cors.Default())

	app.GET("/", func(c *gin.Context) {
		// Index Page
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "This is Weibo SGK. Made by Vincent.",
			"usage":   "GET/POST to /wb with parameter u",
		})
	})

	app.Any("/wb", func(c *gin.Context) {
		// Check if database is available
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Database connection not available",
			})
			return
		}

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
	app.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "Path not found",
		})
	})
}

// Entrypoint is the serverless function handler for Vercel
func Entrypoint(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}

// Type assertion to ensure our function matches http.HandlerFunc
var _ http.HandlerFunc = Entrypoint
