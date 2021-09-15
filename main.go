package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/patarapolw/atexit"
	"github.com/rep2recall/duolog"
	"github.com/rep2recall/plugin-zh/api"
	"github.com/rep2recall/plugin-zh/db"
	"github.com/yanyiwu/gojieba"
)

var jieba *gojieba.Jieba

func Tokenize(s string) []string {
	return jieba.CutForSearch(s, true)
}

func main() {
	jieba = gojieba.NewJieba()
	atexit.Register(jieba.Free)
	atexit.Listen()
	defer jieba.Free()

	if len(os.Args) > 1 {
		fmt.Println(strings.Join(Tokenize(os.Args[1]), " "))
	} else {
		db.Connect()

		d := duolog.Duolog{
			NoColor: true,
		}
		d.New()

		app := fiber.New()
		app.Use(recover.New(recover.Config{
			EnableStackTrace: true,
		}))
		app.Use(logger.New(logger.Config{
			Output: d,
			Format: "[${time}] :${port} ${status} - ${latency} ${method} ${path} ${queryParams}\n",
		}))

		app.Get("/tokenize", func(c *fiber.Ctx) error {
			var query struct {
				Q string `query:"q" validate:"required"`
			}

			if e := c.QueryParser(&query); e != nil {
				return fiber.ErrBadRequest
			}

			return c.JSON(map[string]interface{}{
				"result": Tokenize(query.Q),
			})
		})

		api.Vocab(app)
		api.Sentence(app)

		port := os.Getenv("PORT")
		if port == "" {
			port = "27002"
		}
		app.Listen(":" + port)
	}
}
