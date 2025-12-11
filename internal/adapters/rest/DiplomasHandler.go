package rest

import (
	"encoding/json"
	"gosmol/internal/apperror"
	"gosmol/internal/domain"
	"gosmol/pkg/logging"
	"net/http"
	"strconv"
	"fmt"

	"github.com/julienschmidt/httprouter"
)

type DiplomasService interface {
	GetResources(limits int64) ([]domain.Diploma, error)
	GetResource(id int64) (domain.Diploma, error)
	CreateResource(diploma domain.Diploma) (domain.Diploma, error)
    UpdateResource(id int64, diploma domain.Diploma) (domain.Diploma, error)
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
	resourceURL = "/api/resource/:id"
	resourcesURL = "/api/resources"
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

    fmt.Printf("DEBUG HANDLER DIPLOMA CREATE: Received - Title: %s\n", diploma.Title)

    createdDiploma, err := d.service.CreateResource(diploma)
    if err != nil {
        d.logger.Error("Failed to create resource: " + err.Error())
        http.Error(w, err.Error(), http.StatusBadRequest)
        return err
    }

    fmt.Printf("DEBUG HANDLER DIPLOMA CREATE: Diploma created with ID: %d\n", createdDiploma.ID)
    
    w.WriteHeader(http.StatusCreated)
    return json.NewEncoder(w).Encode(createdDiploma)
}

func (d *DiplomasHandler) put(w http.ResponseWriter, r *http.Request) error {
    fmt.Printf("DEBUG HANDLER DIPLOMA PUT: Method called\n")
    
    w.Header().Set("Content-Type", "application/json")

    params := httprouter.ParamsFromContext(r.Context())
    fmt.Printf("DEBUG HANDLER DIPLOMA PUT: Params: %+v\n", params)
    
    idParams, err := strconv.Atoi(params.ByName("id"))
    if err != nil {
        fmt.Printf("DEBUG HANDLER DIPLOMA PUT: Param error: %v\n", err)
        d.logger.Error("Failed to params: " + err.Error())
        http.Error(w, err.Error(), http.StatusBadRequest)
        return err
    }

    id := int64(idParams)
    fmt.Printf("DEBUG HANDLER DIPLOMA PUT: Updating diploma ID: %d\n", id)

    var diploma domain.Diploma
    if err := json.NewDecoder(r.Body).Decode(&diploma); err != nil {
        fmt.Printf("DEBUG HANDLER DIPLOMA PUT: JSON decode error: %v\n", err)
        d.logger.Error("Failed to decode JSON: " + err.Error())
        http.Error(w, err.Error(), http.StatusBadRequest)
        return err
    }
    defer r.Body.Close()

    fmt.Printf("DEBUG HANDLER DIPLOMA PUT: Received diploma - Title: %s\n", diploma.Title)

    updatedDiploma, err := d.service.UpdateResource(id, diploma)
    if err != nil {
        fmt.Printf("DEBUG HANDLER DIPLOMA PUT: Service error: %v\n", err)
        d.logger.Error("Failed to update resource: " + err.Error())
        http.Error(w, err.Error(), http.StatusBadRequest)
        return err
    }

    fmt.Printf("DEBUG HANDLER DIPLOMA PUT: Successfully updated diploma ID: %d\n", updatedDiploma.ID)
    
    return json.NewEncoder(w).Encode(updatedDiploma)
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

    diploma, err := d.service.GetResource(id)
    if err != nil {
        d.logger.Error("Failed to get resource for deletion: " + err.Error())
        http.Error(w, err.Error(), http.StatusBadRequest)
        return err
    }

    if err := d.service.DeleteResource(id); err != nil {
        d.logger.Error("Failed to delete resource: " + err.Error())
        http.Error(w, err.Error(), http.StatusBadRequest)
        return err
    }

    fmt.Printf("DEBUG HANDLER DIPLOMA DELETE: Diploma deleted with ID: %d\n", id)
    
    return json.NewEncoder(w).Encode(diploma)
}