package main

import (
	"calendar"
	"flag"
	"fmt"
	"heater"
	"log"
	"time"

	_ "docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

var setTemp chan heater.HeaterTemp
var sensorRequest chan heater.Sensors
var sensorResponse chan heater.Sensors
var he heater.Heater

type OvrType struct {
	Val time.Time `json:"Time"`
}

// @title Heater REST API
// @version 1.0
// @description Heater server
// @host 192.168.1.111
// @BasePath /heaterapi/v1

// Sensors godoc
// @Summary Retrieves current sensors
// @Produce json
// @Success 200
// @Router /sensors [get]
func Sensors(ctx *gin.Context) {
	var res heater.Sensors
	sensorRequest <- res
	res = <-sensorResponse
	ctx.JSON(200, gin.H{
		"temp":     res.Temp,
		"pressure": res.Pressure,
		"relay":    res.Relay})
	return
}

// GetDefaultTemp godoc
// @Summary Get default temperature
// @Produce json
// @Success 200
// @Router /defaulttemp [get]
func GetDefaultTemp(ctx *gin.Context) {
	ctx.JSON(200, he.DefaultTemp)
	return
}

// SetDefaultTemp godoc
// @Summary Set default temperature. Body should contain JSON { TLow: <tlow>, THigh: <thigh>, Warm: <warm> }
// @Accept json
// @Produce json
// @Success 204
// @Router /defaulttemp [put]
func SetDefaultTemp(ctx *gin.Context) {
	var newTemp heater.HeaterTemp
	if ctx.BindJSON(&newTemp) == nil {
		setTemp <- newTemp
		ctx.Status(204)
	} else {
		ctx.Status(400)
	}
	return
}

// GetCalendar godoc
// @Summary Retreives current calendars
// @Produce json
// @Success 200
// @Param id query string false "optional calendar id"
// @Router /calendar [get]
func GetCalendar(ctx *gin.Context) {
	cid := ctx.DefaultQuery("id", "")
	if cid == "" {
		ctx.JSON(200, calendar.AllCals.Cals)
	} else {
		ctx.JSON(200, calendar.AllCals.Cals[cid])
	}
	return
}

// SetCalendar godoc
// @Summary Update or create calendar body should contain JSON with data
// @Produce json
// @Success 204
// @Param id query string false "calendar ID"
// @Router /calendar [put]
func SetCalendar(ctx *gin.Context) {
	var newCals calendar.HeaterCalendar
	cid := ctx.DefaultQuery("id", "")
	if ctx.BindJSON(&newCals) != nil {
		ctx.Status(400)
		return
	}
	calendar.AllCals.AddCalendar(cid, newCals)
	ctx.Status(204)
	return
}

// DeleteCalendar godoc
// @Summary Delete calendar
// @Produce json
// @Success 204
// @Param id query string true "calendar ID"
// @Router /calendar [delete]
func DeleteCalendar(ctx *gin.Context) {
	cid := ctx.DefaultQuery("id", "")
	if cid == "" {
		ctx.Status(400)
		return
	}
	calendar.AllCals.RemoveCalendar(cid)
	ctx.Status(204)
	return
}

// SetOverride godoc
// @Summary set calendar override
// @Produce json
// @Success 204
// @Router /override [put]
func SetOverride(ctx *gin.Context) {
	var ovr OvrType
	if err := ctx.BindJSON(&ovr); err != nil {
		fmt.Println(err)
		ctx.Status(400)
		return
	}
	he.SetOverride(ovr.Val)
	ctx.Status(204)
	return
}

// GetOverride godoc
// @Summary Retreives calendar override
// @Produce json
// @Success 200
// @Router /override [get]
func GetOverride(ctx *gin.Context) {
	var override struct {
		Val time.Time `json:"time"`
	}
	override.Val = he.GetOverride()
	ctx.JSON(200, override)
	return
}

func main() {
	var tmp string
	var cls string
	flag.StringVar(&cls, "cals", "", "calendars")
	flag.StringVar(&tmp, "temp", "", "temps")
	flag.Parse()
	if tmp == "" || cls == "" {
		log.Fatal("-cals=<filename> -temp=<filename>")
	}
	calendar.AllCals.Init(cls)
	he.Init(1, 11, 0x76, tmp)
	defer he.Stop()
	setTemp = make(chan heater.HeaterTemp)
	sensorRequest = make(chan heater.Sensors)
	sensorResponse = make(chan heater.Sensors)
	go he.Runner(setTemp, sensorRequest, sensorResponse)
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	v1 := r.Group("/heaterapi/v1")
	{
		v1.GET("/sensors", Sensors)
		v1.GET("/defaulttemp", GetDefaultTemp)
		v1.PUT("/defaulttemp", SetDefaultTemp)
		v1.GET("/calendar", GetCalendar)
		v1.PUT("/calendar", SetCalendar)
		v1.DELETE("/calendar", DeleteCalendar)
		v1.GET("/override", GetOverride)
		v1.PUT("/override", SetOverride)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8080")
}
