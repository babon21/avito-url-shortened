package db

import (
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type DbUrlClient interface {
	AddUrl(url string) (uint64, error)
	GetOriginUrl(inputUrl string) string
	AddShortUrl(url string, id uint64) error
	UpdateCustomUrl(customUrl string, shortUrl string) (int64, error)
}

type DbPostgreSQLUrlClient struct {
	Db *sqlx.DB
}

func (h *DbPostgreSQLUrlClient) GetOriginUrl(inputUrl string) string {
	var originUrl string
	h.Db.QueryRow("SELECT origin_url FROM urls WHERE short_url = $1 OR custom_url = $2", inputUrl, inputUrl).Scan(&originUrl)
	return originUrl
}

func (h *DbPostgreSQLUrlClient) AddUrl(url string) (uint64, error) {
	var id uint64
	err := h.Db.QueryRow("INSERT INTO urls(origin_url) VALUES ($1) RETURNING id", url).Scan(&id)
	return id, err
}

func (h *DbPostgreSQLUrlClient) AddShortUrl(url string, id uint64) error {
	userInsert := "UPDATE urls SET short_url = $1 WHERE id = $2"
	_, err := h.Db.Exec(userInsert, url, id)
	return err
}

func (h *DbPostgreSQLUrlClient) UpdateCustomUrl(customUrl string, shortUrl string) (int64, error) {
	userInsert := `UPDATE urls SET custom_url = $1 WHERE short_url = $2;`

	result, err := h.Db.Exec(userInsert, customUrl, shortUrl)
	if err != nil {
		log.Err(err).Send()
		return -1, err
	}

	num, err := result.RowsAffected()
	if err != nil {
		log.Err(err).Send()
		return -1, err
	}

	return num, nil
}
