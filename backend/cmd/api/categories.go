package main

import (
	"errors"
	"fmt"
	"net/http"

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
		app.handleCategoryWriteError(w, r, v, err)
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

	category, err := app.models.Categories.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"category": category}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	category, err := app.models.Categories.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		NameUz   *string `json:"name_uz"`
		NameRu   *string `json:"name_ru"`
		NameEn   *string `json:"name_en"`
		ParentID *int    `json:"parent_id"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.NameUz != nil {
		category.NameUz = *input.NameUz
	}
	if input.NameRu != nil {
		category.NameRu = *input.NameRu
	}
	if input.NameEn != nil {
		category.NameEn = *input.NameEn
	}
	if input.ParentID != nil {
		category.ParentID = input.ParentID
	}

	v := validator.New()

	if data.ValidateCategory(v, category); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if input.NameEn != nil {
		category.Slug = data.Slugify(category.NameEn)
		if category.Slug == "" {
			v.AddError("name_en", "must contain at least one latin letter or digit for slug generation")
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
	}

	err = app.models.Categories.Update(category)
	if err != nil {
		app.handleCategoryWriteError(w, r, v, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"category": category}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Categories.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "category successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// handleCategoryWriteError translates errors from Insert/Update into client responses.
func (app *application) handleCategoryWriteError(w http.ResponseWriter, r *http.Request, v *validator.Validator, err error) {
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
	case errors.Is(err, data.ErrEditConflict):
		app.editConflictResponse(w, r)
	default:
		app.serverErrorResponse(w, r, err)
	}
}
