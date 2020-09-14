package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/babon21/avito-url-shortened/utils"
	"github.com/jmoiron/sqlx"
	"net/http"
)

type UrlShortHandler struct {
	BaseUrl string
	Db      *sqlx.DB
}

type Result struct {
	ShortUrl  string `json:"short_url"`
	CustomUrl string `json:"custom_url,omitempty"`
}

type Error struct {
	Status int
	Detail string
}

func (h *UrlShortHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// TODO logging
	//fmt.Println("Serve this url!")

	if r.URL.Path != "/api" {
		// TODO logging
		// TODO получить origin url из базы по short url или custom url
		http.Redirect(w, r, "mail.ru", http.StatusSeeOther)
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

func generateUrl(id uint64) string {
	return utils.ToBase62(id)
}

func (h *UrlShortHandler) HandleShortUrl(w http.ResponseWriter, r *http.Request) {
	// TODO logging

	// validate request url

	url := r.FormValue("origin_url")

	var id uint64
	err := h.Db.QueryRow("INSERT INTO urls(origin_url) VALUES ($1) RETURNING id", url).Scan(&id)
	if err != nil {
		fmt.Println(err)
		// TODO logging
		WriteResponseWithError(w, Error{
			Status: http.StatusInternalServerError,
			Detail: "Error while adding origin url to db.",
		})
		return
	}

	generatedUrl := generateUrl(id)

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

	WriteSuccessResponse(w, http.StatusCreated, Result{ShortUrl: generatedUrl})
}

func (h *UrlShortHandler) HandleCustomUrl(w http.ResponseWriter, r *http.Request) {

	// validate request url

	shortUrl := r.FormValue("url_shortener")
	customUrl := r.FormValue("custom_url")
	// validate post param

	userInsert := `UPDATE urls SET custom_url = $1 WHERE short_url = $2;`
	result, err := h.Db.Exec(userInsert, customUrl, shortUrl)

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
		fmt.Println("num == 1")
		WriteSuccessResponse(w, http.StatusOK, Result{
			ShortUrl:  shortUrl,
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

func WriteSuccessResponse(w http.ResponseWriter, responseCode int, r Result) {
	response := make(map[string]interface{})
	response["data"] = [1]Result{r}
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
