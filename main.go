package main

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

var roomManager *RoomManager

func main() {

	roomManager = NewRoomManager()

	r := gin.Default()
	r.LoadHTMLGlob("static/*")

	roomGroup := r.Group("/room")
	roomGroup.GET("/:roomid", roomGet)
	roomGroup.GET("/:roomid/stream", roomStream)
	roomGroup.POST("/:roomid", roomPostMessage)

	r.Run(":3001")
}

func roomGet(c *gin.Context) {
	roomid := c.Param("roomid")
	userid := fmt.Sprint(rand.Int31())
	c.HTML(http.StatusOK, "chat_room.html", gin.H{
		"roomid": roomid,
		"userid": userid,
	})
}

func roomStream(c *gin.Context) {
	roomid := c.Param("roomid")

	listener := roomManager.OpenListener(roomid)
	defer roomManager.CloseListener(roomid, listener)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")

	c.Writer.Flush()

	for {
		select {
		case m := <-listener:
			c.SSEvent("message", m)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func roomPostMessage(c *gin.Context) {
	roomid := c.Param("roomid")

	json := struct {
		Userid string `json:"userid"`
		Text   string `json:"text"`
	}{}

	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	roomManager.Submit(json.Userid, roomid, json.Text)
}
