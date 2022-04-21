// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-19
package webserver

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HandleIndex index file
func HandleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.twig", gin.H{
		"port":    Session.WebPort(),
		"streams": Session.Streams,
		"version": time.Now().String(),
		"page":    "index",
	})
}

func HandlePageDocumentation(c *gin.Context) {
	c.HTML(http.StatusOK, "documentation.twig", gin.H{
		"port":    Session.WebPort(),
		"streams": Session.Streams,
		"version": time.Now().String(),
		"page":    "documentation",
	})
}

func HandlePageStreamList(c *gin.Context) {
	c.HTML(http.StatusOK, "stream_list.twig", gin.H{
		"port":    Session.WebPort(),
		"streams": Session.Streams,
		"version": time.Now().String(),
		"page":    "stream_list",
	})
}

func HandlePagePlayHLS(c *gin.Context) {
	c.HTML(http.StatusOK, "play_hls.twig", gin.H{
		"port":    Session.WebPort(),
		"streams": Session.Streams,
		"version": time.Now().String(),
		"page":    "play_hls",
		"uuid":    c.Param("uuid"),
		"channel": c.Param("channel"),
	})
}
func HandlePagePlayMSE(c *gin.Context) {
	c.HTML(http.StatusOK, "play_mse.twig", gin.H{
		"port":    Session.WebPort(),
		"streams": Session.Streams,
		"version": time.Now().String(),
		"page":    "play_mse",
		"uuid":    c.Param("uuid"),
		"channel": c.Param("channel"),
	})
}
func HandlePagePlayWebRTC(c *gin.Context) {
	c.HTML(http.StatusOK, "play_webrtc.twig", gin.H{
		"port":    Session.WebPort(),
		"streams": Session.Streams,
		"version": time.Now().String(),
		"page":    "play_webrtc",
		"uuid":    c.Param("uuid"),
		"channel": c.Param("channel"),
	})
}
func HandlePageAddStream(c *gin.Context) {
	c.HTML(http.StatusOK, "add_stream.twig", gin.H{
		"port":    Session.WebPort(),
		"streams": Session.Streams,
		"version": time.Now().String(),
		"page":    "add_stream",
	})
}
func HandlePageEditStream(c *gin.Context) {
	c.HTML(http.StatusOK, "edit_stream.twig", gin.H{
		"port":    Session.WebPort(),
		"streams": Session.Streams,
		"version": time.Now().String(),
		"page":    "edit_stream",
		"uuid":    c.Param("uuid"),
	})
}

func HandlePageMultiview(c *gin.Context) {
	c.HTML(http.StatusOK, "multiview.twig", gin.H{
		"port":    Session.WebPort(),
		"streams": Session.Streams,
		"version": time.Now().String(),
		"page":    "multiview",
	})
}

func HandlePagePlayAll(c *gin.Context) {
	c.HTML(http.StatusOK, "play_all.twig", gin.H{
		"port":    Session.WebPort(),
		"streams": Session.Streams,
		"version": time.Now().String(),
		"page":    "play_all",
		"uuid":    c.Param("uuid"),
		"channel": c.Param("channel"),
	})
}

type MultiViewOptions struct {
	Grid   int                             `json:"grid"`
	Player map[string]MultiViewOptionsGrid `json:"player"`
}
type MultiViewOptionsGrid struct {
	UUID       string `json:"uuid"`
	Channel    int    `json:"channel"`
	PlayerType string `json:"playerType"`
}

func HandlePageFullScreenMultiView(c *gin.Context) {
	var createParams MultiViewOptions
	err := c.ShouldBindJSON(&createParams)
	if err != nil {
		log.Printf("bindjson has an error: %s", err)
	}

	log.Printf("bindjson has an error: %s", err)

	log.Printf("options: %v", createParams)
	c.HTML(http.StatusOK, "fullscreenmulti.twig", gin.H{
		"port":    Session.WebPort(),
		"streams": Session.Streams,
		"version": time.Now().String(),
		"options": createParams,
		"page":    "fullscreenmulti",
		"query":   c.Request.URL.Query(),
	})
}
