package handlers

import (
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/babon21/avito-url-shortened/db"
	"github.com/babon21/avito-url-shortened/utils"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strings"
)

type UrlShortHandler struct {
	BaseUrl string
	Db      db.DbUrlClient
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

func (e Error) Error() string {
	return e.Detail
}

func (h *UrlShortHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Receive request with URL: " + r.URL.Path)

	if r.URL.Path != "/api" {
		h.redirect(w, r)
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.handleShortUrl(w, r)
	case http.MethodPatch:
		h.handleCustomUrl(w, r)
	default:
		writeResponseWithError(w, Error{
			Status: http.StatusBadRequest,
			Detail: "Invalid HTTP method",
		})
		log.Error().Msg("Invalid HTTP method")
	}
}

func (h *UrlShortHandler) handleShortUrl(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Handle url shortener endpoint")

	request, err := decodeShortUrlRequest(r.Body)
	if err != nil {
		writeResponseWithError(w, *err)
		return
	}

	log.Info().Msg("Request origin URL: " + request.OriginUrl)

	err = validateShortUrlRequest(request)
	if err != nil {
		writeResponseWithError(w, *err)
	}

	shortUrl, err := h.shortenUrl(request)
	if err != nil {
		writeResponseWithError(w, *err)
		return
	}

	writeSuccessResponse(w, http.StatusCreated, UpdateCustomUrl{ShortUrl: shortUrl})
}

func validateShortUrlRequest(request *ShortUrlRequest) *Error {
	if !govalidator.IsURL(request.OriginUrl) {
		log.Error().Msg("origin url param isn't URL: " + request.OriginUrl)

		return &Error{
			Status: http.StatusBadRequest,
			Detail: "Origin url is not valid.",
		}
	}

	return nil
}

func decodeShortUrlRequest(body io.ReadCloser) (*ShortUrlRequest, *Error) {
	var request ShortUrlRequest

	err := json.NewDecoder(body).Decode(&request)
	if err != nil {
		log.Err(err).Send()

		return nil, &Error{
			Status: http.StatusBadRequest,
			Detail: err.Error(),
		}
	}

	return &request, nil
}

func (h *UrlShortHandler) handleCustomUrl(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Handle update custom url endpoint.")

	var request, err = decodeUpdateCustomUrlRequest(r.Body)
	if err != nil {
		writeResponseWithError(w, *err)
		return
	}

	err = validateUpdateCustomURlRequest(request)
	if err != nil {
		writeResponseWithError(w, *err)
		return
	}

	err = h.UpdateCustomUrl(request)
	if err != nil {
		writeResponseWithError(w, *err)
		return
	}

	writeSuccessResponse(w, http.StatusOK, UpdateCustomUrl{
		ShortUrl:  request.ShortUrl,
		CustomUrl: request.CustomUrl,
	})
}

func (h *UrlShortHandler) UpdateCustomUrl(request *UpdateCustomUrl) *Error {
	num, err := h.Db.UpdateCustomUrl(request.CustomUrl, request.ShortUrl)
	if err != nil {
		log.Err(err).Send()

		return &Error{
			Status: http.StatusUnprocessableEntity,
			Detail: "This custom url is already taken.",
		}
	}

	if num != 1 {
		log.Err(err).Send()

		return &Error{
			Status: http.StatusNotFound,
			Detail: "Short url doesn't exist.",
		}
	}

	return nil
}

func (h *UrlShortHandler) redirect(w http.ResponseWriter, r *http.Request) {
	resultUrl := h.BaseUrl + r.URL.Path
	originUrl := h.Db.GetOriginUrl(resultUrl)

	if !strings.HasPrefix(originUrl, "http://") && !strings.HasPrefix(originUrl, "https://") {
		originUrl = "http://" + originUrl
	}

	log.Info().Msg("Redirect to " + originUrl + " by " + resultUrl)
	http.Redirect(w, r, originUrl, http.StatusSeeOther)
}

func decodeUpdateCustomUrlRequest(body io.ReadCloser) (*UpdateCustomUrl, *Error) {
	var request UpdateCustomUrl
	err := json.NewDecoder(body).Decode(&request)
	if err != nil {
		log.Err(err).Send()

		return nil, &Error{
			Status: http.StatusBadRequest,
			Detail: err.Error(),
		}
	}

	return &request, nil
}

func validateUpdateCustomURlRequest(request *UpdateCustomUrl) *Error {
	if !govalidator.IsURL(request.ShortUrl) {
		log.Error().Msg("short url param isn't URL: " + request.ShortUrl)

		return &Error{
			Status: http.StatusBadRequest,
			Detail: "Short url is not valid.",
		}
	}

	if !govalidator.IsURL(request.CustomUrl) {
		log.Error().Msg("custom url param isn't URL: " + request.ShortUrl)

		return &Error{
			Status: http.StatusBadRequest,
			Detail: "Custom url is not valid.",
		}
	}

	return nil
}

func (h *UrlShortHandler) generateUrl(id uint64) string {
	encodedId := utils.ToBase62(id)
	return h.BaseUrl + "/" + encodedId
}

func (h *UrlShortHandler) shortenUrl(request *ShortUrlRequest) (string, *Error) {
	id, err := h.Db.AddUrl(request.OriginUrl)
	if err != nil {
		log.Err(err).Send()

		return "", &Error{
			Status: http.StatusInternalServerError,
			Detail: "Error while adding origin url to db.",
		}
	}

	shortUrl := h.generateUrl(id)
	err = h.Db.AddShortUrl(shortUrl, id)

	if err != nil {
		log.Err(err).Send()

		return "", &Error{
			Status: http.StatusInternalServerError,
			Detail: "Error while adding generated short url to db.",
		}
	}

	return shortUrl, nil
}

func writeSuccessResponse(w http.ResponseWriter, responseCode int, r UpdateCustomUrl) {
	response := make(map[string]interface{})
	response["data"] = [1]UpdateCustomUrl{r}
	bytes, _ := json.MarshalIndent(response, "", "    ")

	w.WriteHeader(responseCode)
	w.Write(bytes)
}

func writeResponseWithError(w http.ResponseWriter, e Error) {
	response := make(map[string]interface{})
	response["errors"] = [1]Error{e}
	bytes, _ := json.MarshalIndent(response, "", "    ")

	w.WriteHeader(e.Status)
	w.Write(bytes)
}
