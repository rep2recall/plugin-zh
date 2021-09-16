package api

import (
	"database/sql"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/rep2recall/plugin-zh/db"
)

func Vocab(app *fiber.App) {
	type cedict struct {
		Simplified  string   `json:"simplified"`
		Traditional *string  `json:"traditional"`
		Reading     string   `json:"reading"`
		English     []string `json:"english"`
	}

	r := app.Group("/vocab")

	r.Get("/match", func(c *fiber.Ctx) error {
		var query struct {
			Q string `query:"q" validate:"required"`
		}

		if e := c.QueryParser(&query); e != nil {
			return fiber.ErrBadRequest
		}

		rows, err := db.DB.Query(`
		SELECT simplified, traditional, reading, english
		FROM cedict
		WHERE simplified = ? OR traditional = ?
		`, query.Q, query.Q)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		result := struct {
			Result []cedict `json:"result"`
		}{
			Result: make([]cedict, 0),
		}

		for rows.Next() {
			o := cedict{}
			var trad sql.NullString
			eng := ""

			if err := rows.Scan(&o.Simplified, &trad, &o.Reading, &eng); err != nil {
				panic(err)
			}

			if trad.Valid {
				o.Traditional = &trad.String
			}

			if err := json.Unmarshal([]byte(eng), &o.English); err != nil {
				panic(err)
			}

			result.Result = append(result.Result, o)
		}

		return c.JSON(result)
	})
}
