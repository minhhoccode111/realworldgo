package utils

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/gosimple/slug"
)

func GenerateUniqueSlug(ctx context.Context, db *sql.DB, title string) (string, error) {
	baseSlug := slug.Make(title)
	finalSlug := baseSlug
	for i := 0; ; i++ {
		var exists bool
		err := db.QueryRowContext(ctx, `
			select exists(
			select 1 from articles where slug = $1
			)
			`,
			finalSlug,
		).Scan(&exists)

		if err != nil {
			return "", err
		}

		if !exists {
			break
		}

		finalSlug = baseSlug + "-" + strconv.Itoa(i)
	}

	return finalSlug, nil
}
