package life

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/tools"
)

type ScheduleOption struct {
	Year  string `form:"year" json:"year" xml:"year"  binding:"required"`
	Month string `form:"month" json:"month" xml:"month"  binding:"required"`
}

func GetSchedulesWithOptions(c *gin.Context) {
	var optionData ScheduleOption
	if err := c.ShouldBindJSON(&optionData); err != nil {
		c.String(http.StatusBadRequest,
			`Empty or malformed option data.
			비어 있거나 올바르지 않은 조건 데이터 입니다.`)
		return
	}
	targetURL := fmt.Sprintf("%s/calendar/calendar_list_1.aspx?strYear=%s&strMonth=%s",
		consts.SkhuURL, optionData.Year, optionData.Month)

	client := &http.Client{}
	req, err := http.NewRequest("GET", targetURL, nil)
	res, err := client.Do(req)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(tools.EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		log.Fatal(err)
	}

	schedules := []gin.H{}
	doc.Find("div.info > table > tbody > tr").Each(func(i int, item *goquery.Selection) {
		schedules = append(schedules, gin.H{
			"period":  item.Children().Eq(0).Text(),
			"content": item.Children().Eq(1).Text(),
		})
	})

	c.JSON(http.StatusOK, gin.H{
		"schedules": schedules,
	})
}
