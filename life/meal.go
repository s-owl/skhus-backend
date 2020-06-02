package life

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/tools"
)

func GetMealURLs(c *gin.Context) {
	targetURL := consts.SkhuURL + "/uni_zelkova/uni_zelkova_4_3_list.aspx"

	client := &http.Client{}
	req, err := http.NewRequest("GET", targetURL, nil)
	res, err := client.Do(req)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(tools.EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		log.Fatal(err)
	}

	mealUrls := []gin.H{}
	doc.Find("table.board_list > tbody > tr").Each(func(i int, item *goquery.Selection) {
		mealUrls = append(mealUrls, gin.H{
			"title": item.Children().Eq(1).Find("a").Text(),
			"url":   consts.SkhuURL + "/uni_zelkova/" + item.Children().Eq(1).Find("a").AttrOr("href", ""),
			"date":  item.Children().Eq(3).Text(),
		})
	})
	c.JSON(http.StatusOK, gin.H{
		"urls": mealUrls,
	})
	return
}

type MealOption struct {
	URL string `form:"url" json:"url" xml:"url"  binding:"required"`
}

const theadSelector string = `thead > tr:nth-child(%d) > th:nth-child(%d)`
const tbodySelector string = `tbody > tr:nth-child(%d) > td:nth-child(%d)`

func GetMealData(c *gin.Context) {
	defaultURL := consts.SkhuURL + "/uni_zelkova/uni_zelkova_4_3_first.aspx"

	var targetURL string
	var optionData MealOption
	if err := c.ShouldBindJSON(&optionData); err != nil {
		targetURL = defaultURL
	} else {
		targetURL = optionData.URL
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", targetURL, nil)
	res, err := client.Do(req)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(tools.EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		log.Fatal(err)
	}

	meal := []gin.H{}
	mealTable := doc.Find("table.cont_c")

	for i := 0; i < 5; i++ {
		meal = append(meal, gin.H{
			"day":  mealTable.Find(fmt.Sprintf(theadSelector, 1, i+2)).Text(),
			"date": mealTable.Find(fmt.Sprintf(theadSelector, 2, i+3)).Text(),
			"lunch": gin.H{
				"a": gin.H{
					"diet":    processDietData(mealTable, 1, i+3),
					"calorie": mealTable.Find(fmt.Sprintf(tbodySelector, 2, i+3)).Text(),
				},
				"b": gin.H{
					"diet":    processDietData(mealTable, 3, i+2),
					"calorie": mealTable.Find(fmt.Sprintf(tbodySelector, 4, i+2)).Text(),
				},
				"c": gin.H{
					"diet":    processDietData(mealTable, 5, i+2),
					"calorie": mealTable.Find(fmt.Sprintf(tbodySelector, 6, i+2)).Text(),
				},
			},
			"dinner": gin.H{
				"a": gin.H{
					"diet":    processDietData(mealTable, 7, i+3),
					"calorie": mealTable.Find(fmt.Sprintf(tbodySelector, 8, i+3)).Text(),
				},
			},
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data": meal,
	})
}

func processDietData(sel *goquery.Selection, trIndex int, tdIndex int) string {
	item := sel.Find(fmt.Sprintf(tbodySelector, trIndex, tdIndex))
	htmlContent, _ := item.Html()
	content := strings.ReplaceAll(htmlContent, "<br/>", "\n")
	return content
}
