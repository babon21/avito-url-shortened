package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/babon21/avito-url-shortened/utils"
	"github.com/jmoiron/sqlx"
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

	// TODO logging

	if r.URL.Path != "/api" {
		// TODO logging
		var originUrl string
		resultUrl := h.BaseUrl + r.URL.Path
		h.Db.QueryRow("SELECT origin_url FROM urls WHERE short_url = $1 OR custom_url = $2", resultUrl, resultUrl).Scan(&originUrl)

		if !strings.HasPrefix(originUrl, "http://") && !strings.HasPrefix(originUrl, "https://") {
			originUrl = "http://" + originUrl
		}

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
			Detail: "Uknown action.",
		})
		// TODO logging
		fmt.Println("Default case, unknown http method!")
	}
}

func (h *UrlShortHandler) generateUrl(id uint64) string {
	encodedId := utils.ToBase62(id)
	return h.BaseUrl + "/" + encodedId
}

func (h *UrlShortHandler) HandleShortUrl(w http.ResponseWriter, r *http.Request) {
	// TODO logging

	var request ShortUrlRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		WriteResponseWithError(w, Error{
			Status: http.StatusBadRequest,
			Detail: err.Error(),
		})
		return
	}

	if !govalidator.IsURL(request.OriginUrl) {
		// TODO logging
		WriteResponseWithError(w, Error{
			Status: http.StatusBadRequest,
			Detail: "Origin url is not valid.",
		})
		return
	}

	var id uint64
	err = h.Db.QueryRow("INSERT INTO urls(origin_url) VALUES ($1) RETURNING id", request.OriginUrl).Scan(&id)
	if err != nil {
		fmt.Println(err)
		// TODO logging
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
		fmt.Println(err)
		// TODO logging
		WriteResponseWithError(w, Error{
			Status: http.StatusInternalServerError,
			Detail: "Error while adding generated short url to db.",
		})
		return
	}

	WriteSuccessResponse(w, http.StatusCreated, UpdateCustomUrl{ShortUrl: generatedUrl})
}

func (h *UrlShortHandler) HandleCustomUrl(w http.ResponseWriter, r *http.Request) {
	var request UpdateCustomUrl
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		WriteResponseWithError(w, Error{
			Status: http.StatusBadRequest,
			Detail: err.Error(),
		})
		return
	}

	if !govalidator.IsURL(request.ShortUrl) {
		// TODO logging
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
		// TODO logging
		WriteResponseWithError(w, Error{
			Status: http.StatusBadRequest,
			Detail: "Custom url is not valid.",
		})
		return
	}

	userInsert := `UPDATE urls SET custom_url = $1 WHERE short_url = $2;`
	result, err := h.Db.Exec(userInsert, customUrl, request.ShortUrl)

	if err != nil {
		fmt.Println(err)
		// TODO logging

		WriteResponseWithError(w, Error{
			Status: http.StatusUnprocessableEntity,
			Detail: "This custom url is already taken.",
		})
		return
	}

	num, err := result.RowsAffected()
	if err != nil {
		fmt.Println(err)

		fmt.Println("Error while getting id from database.")
		WriteResponseWithError(w, Error{
			Status: http.StatusInternalServerError,
			Detail: "Error while getting id from database.",
		})
		// TODO logging

		return
	}

	if num == 1 {
		// TODO logging
		WriteSuccessResponse(w, http.StatusOK, UpdateCustomUrl{
			ShortUrl:  request.ShortUrl,
			CustomUrl: customUrl,
		})

		return
	}

	WriteResponseWithError(w, Error{
		Status: http.StatusInternalServerError,
		Detail: "Unknown error.",
	})
	// TODO logging
}

func WriteSuccessResponse(w http.ResponseWriter, responseCode int, r UpdateCustomUrl) {
	response := make(map[string]interface{})
	response["data"] = [1]UpdateCustomUrl{r}
	bytes, _ := json.MarshalIndent(response, "", "    ")

	w.WriteHeader(responseCode)
	w.Write(bytes)
	// TODO logging
}

func WriteResponseWithError(w http.ResponseWriter, e Error) {
	response := make(map[string]interface{})
	response["errors"] = [1]Error{e}
	bytes, _ := json.MarshalIndent(response, "", "    ")

	w.WriteHeader(e.Status)
	w.Write(bytes)
	// TODO logging
}
