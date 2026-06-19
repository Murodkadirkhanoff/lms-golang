package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"lms.chashma.uz/internal/data"
	"lms.chashma.uz/internal/validator"
)

func (app *application) createCategoryHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		NameUz   string `json:"name_uz"`
		NameRu   string `json:"name_ru"`
		NameEn   string `json:"name_en"`
		ParentID *int   `json:"parent_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	category := &data.Category{
		NameUz:   input.NameUz,
		NameRu:   input.NameRu,
		NameEn:   input.NameEn,
		ParentID: input.ParentID,
	}

	v := validator.New()

	if data.ValidateCategory(v, category); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	category.Slug = data.Slugify(category.NameEn)
	if category.Slug == "" {
		v.AddError("name_en", "must contain at least one latin letter or digit for slug generation")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Categories.Insert(category)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateSlug):
			v.AddError("name_en", "a category with this name already exists")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrInvalidParent):
			v.AddError("parent_id", "parent category does not exist")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrMaxDepthExceeded):
			v.AddError("parent_id", "category nesting is too deep (max 2 levels)")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/categories/%d", category.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"category": category}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	category := data.Category{
		ID:        int64(id),
		CreatedAt: time.Now(),
		Slug:      "Category_1",
		NameUz:    "Name UZ",
		NameRu:    "Name RU",
		NameEn:    "Name EN",
		Version:   1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"category": category}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
