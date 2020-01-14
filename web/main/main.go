package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
)

func main() {
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	defer conn.Close()
	r := gin.Default()
	r.GET("/lock", func(c *gin.Context) {
		conn.Do("PUBLISH", "cmd1", "lock")
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/shutdown", func(c *gin.Context) {
		conn.Do("PUBLISH", "cmd1", "shutdown")
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
