package api

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"github.com/rep2recall/plugin-zh/db"
)

func Sentence(app *fiber.App) {
	type sentence struct {
		Cmn string `json:"cmn"`
		Eng string `json:"eng"`
	}

	r := app.Group("/sentence")

	r.Get("/q", func(c *fiber.Ctx) error {
		var query struct {
			Q string `query:"q" validate:"required"`
		}

		if e := c.QueryParser(&query); e != nil {
			return fiber.ErrBadRequest
		}

		rows, err := db.DB.Query(`
		SELECT cmn, eng
		FROM sentences
		WHERE cmn LIKE '%'||?||'%'
		ORDER BY RANDOM()
		LIMIT 5
		`, query.Q)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		result := struct {
			Result []sentence `json:"result"`
		}{
			Result: make([]sentence, 0),
		}

		for rows.Next() {
			o := sentence{}

			if err := rows.Scan(&o.Cmn, &o.Eng); err != nil {
				panic(err)
			}

			result.Result = append(result.Result, o)
		}

		if len(result.Result) < 5 {
			err := func() error {
				client := http.Client{
					Timeout: 10 * time.Second,
				}

				resp, err := client.Get(fmt.Sprintf("http://www.jukuu.com/search.php?q=%s", url.QueryEscape(query.Q)))
				if err != nil {
					return nil
				}

				doc, err := goquery.NewDocumentFromReader(resp.Body)
				if err != nil {
					return nil
				}

				moreResult := make([]sentence, 10)
				maxResult := 0

				doc.Find("table tr.c td:last-child").Each(func(i int, item *goquery.Selection) {
					if i < len(moreResult) {
						moreResult[i].Cmn = item.Text()
					}
					maxResult = i
				})

				doc.Find("table tr.e td:last-child").Each(func(i int, item *goquery.Selection) {
					if i < len(moreResult) {
						moreResult[i].Eng = item.Text()
					}
				})

				moreResult = moreResult[:maxResult]

				for _, r := range moreResult {
					if _, err := db.DB.Exec(`
					INSERT INTO sentences (source, cmn, eng) VALUES ('jukuu', ?, ?)
					ON CONFLICT DO NOTHING
					`, r.Cmn, r.Eng); err != nil {
						return err
					}

					result.Result = append(result.Result, r)
				}

				if len(result.Result) > 5 {
					result.Result = result.Result[:5]
				}

				return nil
			}()

			if err != nil {
				panic(err)
			}
		}

		return c.JSON(result)
	})
}
