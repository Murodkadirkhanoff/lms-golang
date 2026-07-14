package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"lms.chashma.uz/pkg/jsonutil"
	"lms.chashma.uz/pkg/svcclient"
)

// adminStatsHandler — GET /v1/admin/stats (frontend AdminStats shakli).
// totalUsers lokal, qolgani course/enrollment servislaridan yig'iladi.
// Boshqa servis ishlamasa 0 bilan davom etamiz — panel yiqilmasin.
func (app *application) adminStatsHandler(w http.ResponseWriter, r *http.Request) {
	totalUsers, err := app.models.Users.Count()
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	stats := struct {
		TotalUsers        int     `json:"totalUsers"`
		TotalCourses      int     `json:"totalCourses"`
		Revenue           float64 `json:"revenue"`
		ActiveInstructors int     `json:"activeInstructors"`
	}{TotalUsers: totalUsers}

	var courseStats struct {
		TotalCourses      int `json:"totalCourses"`
		ActiveInstructors int `json:"activeInstructors"`
	}
	if err := app.courseClient.Get(r.Context(), "/internal/stats", &courseStats); err != nil {
		app.logger.Warn("admin stats: course-service unavailable", "error", err.Error())
	}
	stats.TotalCourses = courseStats.TotalCourses
	stats.ActiveInstructors = courseStats.ActiveInstructors

	var enrollmentStats struct {
		Revenue float64 `json:"revenue"`
	}
	if err := app.enrollmentClient.Get(r.Context(), "/internal/stats", &enrollmentStats); err != nil {
		app.logger.Warn("admin stats: enrollment-service unavailable", "error", err.Error())
	}
	stats.Revenue = enrollmentStats.Revenue

	err = jsonutil.WriteJSON(w, http.StatusOK, stats, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// fetchCounts boshqa servisdan {counts: {userId: n}} shaklidagi javobni oladi.
// Xato bo'lsa bo'sh map — admin ro'yxati baribir ko'rinadi.
func (app *application) fetchCounts(ctx context.Context, client *svcclient.Client, path string, ids []int64) map[int64]int {
	counts := map[int64]int{}
	if len(ids) == 0 {
		return counts
	}

	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatInt(id, 10))
	}

	var response struct {
		Counts map[string]int `json:"counts"`
	}

	err := client.Get(ctx, path+"?ids="+strings.Join(parts, ","), &response)
	if err != nil {
		app.logger.Warn("failed to fetch counts", "path", path, "error", err.Error())
		return counts
	}

	for key, count := range response.Counts {
		if id, err := strconv.ParseInt(key, 10, 64); err == nil {
			counts[id] = count
		}
	}

	return counts
}
