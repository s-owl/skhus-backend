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

func GetMealURLs(c *gin.Context) {
	targetURL := fmt.Sprintf("%s/uni_zelkova/uni_zelkova_4_3_list.aspx", consts.SkhuURL)

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
			"url":   fmt.Sprintf("%s/uni_zelkova/%s", consts.SkhuURL, item.Children().Eq(1).Find("a").AttrOr("href", "")),
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

func GetMealData(c *gin.Context) {
	defaultURL := fmt.Sprintf("%s/uni_zelkova/uni_zelkova_4_3_first.aspx", consts.SkhuURL)

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
			"day":  mealTable.Find(fmt.Sprintf(`thead > tr:nth-child(1) > th:nth-child(%d)`, i+2)).Text(),
			"date": mealTable.Find(fmt.Sprintf(`thead > tr:nth-child(2) > th:nth-child(%d)`, i+3)).Text(),
			"lunch": gin.H{
				"a": gin.H{
					"diet":    mealTable.Find(fmt.Sprintf(`tbody > tr:nth-child(1) > td:nth-child(%d)`, i+3)).Text(),
					"calorie": mealTable.Find(fmt.Sprintf(`tbody > tr:nth-child(2) > td:nth-child(%d)`, i+3)).Text(),
				},
				"b": gin.H{
					"diet":    mealTable.Find(fmt.Sprintf(`tbody > tr:nth-child(3) > td:nth-child(%d)`, i+2)).Text(),
					"calorie": mealTable.Find(fmt.Sprintf(`tbody > tr:nth-child(4) > td:nth-child(%d)`, i+2)).Text(),
				},
				"c": gin.H{
					"diet":    mealTable.Find(fmt.Sprintf(`tbody > tr:nth-child(5) > td:nth-child(%d)`, i+2)).Text(),
					"calorie": mealTable.Find(fmt.Sprintf(`tbody > tr:nth-child(6) > td:nth-child(%d)`, i+2)).Text(),
				},
			},
			"dinner": gin.H{
				"a": gin.H{
					"diet":    mealTable.Find(fmt.Sprintf(`tbody > tr:nth-child(7) > td:nth-child(%d)`, i+3)).Text(),
					"calorie": mealTable.Find(fmt.Sprintf(`tbody > tr:nth-child(8) > td:nth-child(%d)`, i+3)).Text(),
				},
			},
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data": meal,
	})
}
