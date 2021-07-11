package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type Handler struct {
	storage Storage
}

func NewHandler(storage Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

func (h Handler) AddBuilding(w http.ResponseWriter, r *http.Request) {
	var build Building
	if err := json.NewDecoder(r.Body).Decode(&build); err != nil {
		render.JSON(w, r, map[string]interface{} {
			"HttpStatus":http.StatusBadRequest,
			"Message": err.Error(),
			"Data": nil,
		})
		return
	}
	
	if err := h.storage.InsertBuilding(&build); err != nil {
		render.JSON(w, r, map[string]interface{} {
			"HttpStatus":http.StatusBadRequest,
			"Message": err.Error(),
			"Data": nil,
		})
		return
	}
	
	render.JSON(w, r, map[string]interface{} {
		"HttpStatus":http.StatusOK,
		"Message": "Success",
		"Data": nil,
	})
}

func (h Handler) GetCompaniesFromBuilding(w http.ResponseWriter, r *http.Request) {
	strID := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		render.JSON(w, r, map[string]interface{} {
			"HttpStatus":http.StatusBadRequest,
			"Message": err.Error(),
			"Data": nil,
		})
		return
	}
	comps, err := h.storage.GetCompaniesFromBuilding(id)
	if err != nil {
		render.JSON(w, r, map[string]interface{} {
			"HttpStatus":http.StatusBadRequest,
			"Message": err.Error(),
			"Data": nil,
		})
		return
	}

	jsoncomps := make([]JSONCompany, 0)
	i := 0
	jsonidx := 0

	for {
		if i >= len(comps) {
			break
		}
		jsoncomps = append(jsoncomps, JSONCompany{
			Name:    comps[i].Name,
			Phones:  comps[i].Phones,
			Address: comps[i].Address,
		})
		for j := i; j < len(comps); j++ {
			if comps[j].Name != jsoncomps[jsonidx].Name {
				jsonidx++
				break
			}
			jsoncomps[jsonidx].Rubrics = append(jsoncomps[jsonidx].Rubrics, comps[j].Rubric)
			i++
		}
	}

	render.JSON(w, r, map[string]interface{} {
		"HttpStatus":http.StatusOK,
		"Message": "Success",
		"Data": jsoncomps,
	})
}

func (h Handler) GetCompaniesFromRubric(w http.ResponseWriter, r *http.Request) {
	strID := chi.URLParam(r, "id")
	// fmt.Println(strID)
	id, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		render.JSON(w, r, map[string]interface{} {
			"HttpStatus":http.StatusBadRequest,
			"Message": err.Error(),
			"Data": nil,
		})
		return
	}
	comps, err := h.storage.GetCompaniesFromRubric(id)
	if err != nil {
		render.JSON(w, r, map[string]interface{} {
			"HttpStatus":http.StatusBadRequest,
			"Message": err.Error(),
			"Data": nil,
		})
		return
	}

	jsoncomps := make([]JSONCompany, 0)
	i := 0
	jsonidx := 0

	for {
		if i >= len(comps) {
			break
		}
		jsoncomps = append(jsoncomps, JSONCompany{
			Name:    comps[i].Name,
			Phones:  comps[i].Phones,
			Address: comps[i].Address,
		})
		for j := i; j < len(comps); j++ {
			if comps[j].Name != jsoncomps[jsonidx].Name {
				jsonidx++
				break
			}
			if comps[j].Name == jsoncomps[jsonidx].Name && comps[j].Address != jsoncomps[jsonidx].Address{
				jsonidx++
				break
			}
			jsoncomps[jsonidx].Rubrics = append(jsoncomps[jsonidx].Rubrics, comps[j].Rubric)
			i++
		}
	}

	render.JSON(w, r, map[string]interface{} {
		"HttpStatus":http.StatusOK,
		"Message": "Success",
		"Data": jsoncomps,
	})
}

func (h Handler) GetCompany(w http.ResponseWriter, r *http.Request) {
	strID := chi.URLParam(r, "id")
	// fmt.Println(strID)
	id, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		render.JSON(w, r, map[string]interface{} {
			"HttpStatus":http.StatusBadRequest,
			"Message": err.Error(),
			"Data": nil,
		})
		return
	}
	comps, err := h.storage.GetCompany(id)
	if err != nil {
		render.JSON(w, r, map[string]interface{} {
			"HttpStatus":http.StatusBadRequest,
			"Message": err.Error(),
			"Data": nil,
		})
		return
	}

	jsoncomp := JSONCompany{
		Name:    comps[0].Name,
		Phones:  comps[0].Phones,
		Address: comps[0].Address,
	}

	for _, comp := range comps {
		jsoncomp.Rubrics = append(jsoncomp.Rubrics, comp.Rubric)
	}

	render.JSON(w, r, map[string]interface{} {
		"HttpStatus":http.StatusOK,
		"Message": "Success",
		"Data": jsoncomp,
	})
}
