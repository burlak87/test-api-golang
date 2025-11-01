package rest

import (
	"encoding/json"
	"gosmol/internal/apperror"
	"gosmol/internal/domain"
	"gosmol/pkg/logging"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type DiplomasService interface {
	GetResources(limits int64) ([]domain.Diploma, error)
	GetResource(id int64) (domain.Diploma, error)
	CreateResource(diploma domain.Diploma) error
	UpdateResource(id int64, diploma domain.Diploma) error
	DeleteResource(id int64) error
}

type DiplomasHandler struct {
	service DiplomasService
	logger *logging.Logger
}

func NewDiplomasHandler(s DiplomasService, l *logging.Logger) *DiplomasHandler {
	return &DiplomasHandler{
		service: s,
		logger: l,
	}
}

const (
	resourceURL = "api/resource/{id}"
	resourcesURL = "api/resources"
)

var dip []domain.Diploma

func (d *DiplomasHandler) Register(router *httprouter.Router, jwtSecret string) {
	router.Handler(http.MethodGet, resourcesURL, apperror.JWTMiddleware(jwtSecret, http.HandlerFunc(apperror.Middleware(d.get))))
	router.Handler(http.MethodGet, resourceURL, apperror.JWTMiddleware(jwtSecret, http.HandlerFunc(apperror.Middleware(d.getById))))
	router.Handler(http.MethodPost, resourcesURL, apperror.JWTMiddleware(jwtSecret, http.HandlerFunc(apperror.Middleware(d.post))))
	router.Handler(http.MethodPut, resourceURL, apperror.JWTMiddleware(jwtSecret, http.HandlerFunc(apperror.Middleware(d.put))))
	router.Handler(http.MethodDelete, resourceURL, apperror.JWTMiddleware(jwtSecret, http.HandlerFunc(apperror.Middleware(d.delete))))
}

func (d *DiplomasHandler) get(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	var limits int64 = 25

	diplomas, err := d.service.GetResources(limits)
	if err != nil {
		d.logger.Error("Failed to search resources: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	json.NewEncoder(w).Encode(diplomas)
	return nil
}

func (d *DiplomasHandler) getById(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	params := httprouter.ParamsFromContext(r.Context())
	idParams, err := strconv.Atoi(params.ByName("id")) 
	if err != nil {
		d.logger.Error("Failed to params: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	id := int64(idParams)
	diplomas, err := d.service.GetResource(id)
	if err != nil {
		d.logger.Error("Failed to search resource: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	json.NewEncoder(w).Encode(diplomas)
	return nil
}

func (d *DiplomasHandler) post(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	var diploma domain.Diploma
	if err := json.NewDecoder(r.Body).Decode(&diploma); err != nil {
		d.logger.Error("Failed to decode JSON: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	defer r.Body.Close()

	if err := d.service.CreateResource(diploma); err != nil {
		d.logger.Error("Failed to create resource: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	json.NewEncoder(w).Encode(diploma)
	return nil
}

func (d *DiplomasHandler) put(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	params := httprouter.ParamsFromContext(r.Context())
	idParams, err := strconv.Atoi(params.ByName("id"))

	if err != nil {
		d.logger.Error("Failed to params: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	id := int64(idParams)
	var diploma domain.Diploma

	_ = json.NewDecoder(r.Body).Decode(&diploma)
	defer r.Body.Close()

	if err := d.service.UpdateResource(id, diploma); err != nil {
		d.logger.Error("Failed to update resource: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	json.NewEncoder(w).Encode(diploma)
	return nil
}

func (d *DiplomasHandler) delete(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	params := httprouter.ParamsFromContext(r.Context())
	idParams, err := strconv.Atoi(params.ByName("id")) 

	if err != nil {
		d.logger.Error("Failed to params: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	id := int64(idParams)

	diplomas, err := d.service.GetResource(id)
	if err != nil {
		d.logger.Error("Failed to delete resource: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	json.NewEncoder(w).Encode(diplomas)
	return nil
}