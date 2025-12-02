package main

import (
	"fmt"
	"os"

	"github.com/ainyava/proxyceptionist/internal"
	"github.com/gin-gonic/gin"
)

func main() {
	internal.InitLog()
	defer internal.LogFile.Close()

	r := gin.Default()
	r.Any("/*proxyPath", internal.ProxyReq)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	fmt.Printf("Running on localhost:%s\n", port)
	r.Run(":" + port)
}
