package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	//"overrustlelogs/common"
	"github.com/MemeLabs/overrustlelogs/common"
	"github.com/gin-gonic/gin"
)

func init() {
	configPath := flag.String("config", "/logger/overrustlelogs.toml", "config path")
	flag.Parse()
	common.SetupConfig(*configPath)
}

type Channel struct {
	Channel string `uri:"channel" binding:"required"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	logs := NewChatLogs()

	twitchLogHandler := func(m <-chan *common.Message) {
		NewLogger(NewChatLogs()).TwitchLog(m)
	}

	tl := NewTwitchLogger(twitchLogHandler)
	go tl.Start()

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.GET("/whitelist", func(c *gin.Context) {
		c.File("/logger/channels.json")
	})

	r.POST("/join/:channel", func(c *gin.Context) {
		var channel Channel
		if err := c.ShouldBindUri(&channel); err != nil {
			c.JSON(400, gin.H{"error": err})
			return
		}
		err := tl.join(channel.Channel, true)
		if err != nil {
			c.JSON(500, gin.H{"error": err})
		} else {
			c.JSON(200, gin.H{"msg": "good to go"})
		}
	})

	r.DELETE("/leave/:channel", func(c *gin.Context) {
		var channel Channel
		if err := c.ShouldBindUri(&channel); err != nil {
			c.JSON(400, gin.H{"error": err})
			return
		}
		err := tl.leave(channel.Channel)
		if err != nil {
			c.JSON(500, gin.H{"error": err})
		} else {
			c.JSON(200, gin.H{"msg": "good to go"})
		}
	})

	r.Run(":8080")

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint
	logs.Close()
	tl.Stop()
	log.Println("i love you guys, be careful")
	os.Exit(0)
}
