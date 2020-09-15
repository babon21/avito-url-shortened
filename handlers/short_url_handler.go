package handlers

import (
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/babon21/avito-url-shortened/utils"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

type UrlShortHandler struct {
	BaseUrl string
	Db      *sqlx.DB
}

type UpdateCustomUrl struct {
	ShortUrl  string `json:"short_url"`
	CustomUrl string `json:"custom_url,omitempty"`
}

type ShortUrlRequest struct {
	OriginUrl string `json:"origin_url"`
}

type Error struct {
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

func (h *UrlShortHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Receive request with URL: " + r.URL.Path)

	if r.URL.Path != "/api" {
		var originUrl string
		resultUrl := h.BaseUrl + r.URL.Path
		h.Db.QueryRow("SELECT origin_url FROM urls WHERE short_url = $1 OR custom_url = $2", resultUrl, resultUrl).Scan(&originUrl)

		if !strings.HasPrefix(originUrl, "http://") && !strings.HasPrefix(originUrl, "https://") {
			originUrl = "http://" + originUrl
		}

		log.Info().Msg("Redirect to " + originUrl + " by " + resultUrl)
		http.Redirect(w, r, originUrl, http.StatusSeeOther)
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.HandleShortUrl(w, r)
	case http.MethodPatch:
		h.HandleCustomUrl(w, r)
	default:
		WriteResponseWithError(w, Error{
			Status: http.StatusBadRequest,
			Detail: "Invalid HTTP method",
		})
		log.Error().Msg("Invalid HTTP method")
	}
}

func (h *UrlShortHandler) generateUrl(id uint64) string {
	encodedId := utils.ToBase62(id)
	return h.BaseUrl + "/" + encodedId
}

func (h *UrlShortHandler) HandleShortUrl(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Handle url shortener endpoint")

	var request ShortUrlRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Err(err).Send()

		WriteResponseWithError(w, Error{
			Status: http.StatusBadRequest,
			Detail: err.Error(),
		})
		return
	}
	log.Info().Msg("Request origin URL: " + request.OriginUrl)

	if !govalidator.IsURL(request.OriginUrl) {
		log.Error().Msg("origin url param isn't URL: " + request.OriginUrl)

		WriteResponseWithError(w, Error{
			Status: http.StatusBadRequest,
			Detail: "Origin url is not valid.",
		})
		return
	}

	var id uint64
	err = h.Db.QueryRow("INSERT INTO urls(origin_url) VALUES ($1) RETURNING id", request.OriginUrl).Scan(&id)
	if err != nil {
		log.Err(err).Send()

		WriteResponseWithError(w, Error{
			Status: http.StatusInternalServerError,
			Detail: "Error while adding origin url to db.",
		})
		return
	}

	generatedUrl := h.generateUrl(id)

	userInsert := "UPDATE urls SET short_url = $1 WHERE id = $2"
	_, err = h.Db.Exec(userInsert, generatedUrl, id)

	if err != nil {
		log.Err(err).Send()

		WriteResponseWithError(w, Error{
			Status: http.StatusInternalServerError,
			Detail: "Error while adding generated short url to db.",
		})
		return
	}

	WriteSuccessResponse(w, http.StatusCreated, UpdateCustomUrl{ShortUrl: generatedUrl})
}

func (h *UrlShortHandler) HandleCustomUrl(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Handle update custom url endpoint.")

	var request UpdateCustomUrl
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Err(err).Send()

		WriteResponseWithError(w, Error{
			Status: http.StatusBadRequest,
			Detail: err.Error(),
		})
		return
	}

	if !govalidator.IsURL(request.ShortUrl) {
		log.Error().Msg("origin url param isn't URL: " + request.ShortUrl)

		WriteResponseWithError(w, Error{
			Status: http.StatusBadRequest,
			Detail: "Short url is not valid.",
		})
		return
	}

	customUrl := request.CustomUrl
	if !strings.HasPrefix(request.CustomUrl, h.BaseUrl) {
		customUrl = h.BaseUrl + "/" + request.CustomUrl
	}

	if !govalidator.IsURL(customUrl) {
		log.Error().Msg("Invalid custom_url param: " + request.ShortUrl)

		WriteResponseWithError(w, Error{
			Status: http.StatusBadRequest,
			Detail: "Custom url is not valid.",
		})
		return
	}

	userInsert := `UPDATE urls SET custom_url = $1 WHERE short_url = $2;`
	result, err := h.Db.Exec(userInsert, customUrl, request.ShortUrl)

	if err != nil {
		log.Err(err).Send()

		WriteResponseWithError(w, Error{
			Status: http.StatusUnprocessableEntity,
			Detail: "This custom url is already taken.",
		})
		return
	}

	num, err := result.RowsAffected()
	if err != nil {
		log.Err(err).Send()

		WriteResponseWithError(w, Error{
			Status: http.StatusInternalServerError,
			Detail: "Error while getting id from database.",
		})

		return
	}

	if num == 1 {
		WriteSuccessResponse(w, http.StatusOK, UpdateCustomUrl{
			ShortUrl:  request.ShortUrl,
			CustomUrl: customUrl,
		})

		return
	}

	log.Error().Msg("Database is in the incorrect state.")
	WriteResponseWithError(w, Error{
		Status: http.StatusInternalServerError,
		Detail: "Database is in the incorrect state.",
	})
}

func WriteSuccessResponse(w http.ResponseWriter, responseCode int, r UpdateCustomUrl) {
	response := make(map[string]interface{})
	response["data"] = [1]UpdateCustomUrl{r}
	bytes, _ := json.MarshalIndent(response, "", "    ")

	w.WriteHeader(responseCode)
	w.Write(bytes)
}

func WriteResponseWithError(w http.ResponseWriter, e Error) {
	response := make(map[string]interface{})
	response["errors"] = [1]Error{e}
	bytes, _ := json.MarshalIndent(response, "", "    ")

	w.WriteHeader(e.Status)
	w.Write(bytes)
}
