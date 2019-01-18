package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetFoobar(c *gin.Context) {
	fmt.Fprintf(c.Writer, "foobar")
}

func main() {
	router := gin.Default()
	router.GET("/foobar", GetFoobar)
	router.Run(":8080")
}
