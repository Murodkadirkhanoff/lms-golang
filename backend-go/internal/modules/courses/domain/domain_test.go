package domain

import (
	"testing"

	"github.com/chashma/lms/internal/modules/courses/contract"
	"github.com/chashma/lms/internal/platform/web"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"My Category!":     "my-category",
		"  Hello  World  ": "hello-world",
		"Go & Rust":        "go-rust",
		"---":              "",
		"Ứзбек":            "",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestThumbnailColorDeterministic(t *testing.T) {
	if ThumbnailColor(1) != ThumbnailColor(1) {
		t.Fatal("ThumbnailColor must be deterministic")
	}
	// id+3 offset: id 3 -> index (3+3)%6 = 0
	if ThumbnailColor(3) != "bg-indigo-200" {
		t.Fatalf("unexpected color for id 3: %s", ThumbnailColor(3))
	}
}

func TestValidateCourse(t *testing.T) {
	v := web.NewValidator()
	c := &contract.CourseView{
		Title: "Valid", Lang: "uz", Price: 10,
		Modules: []contract.Module{{
			Title: "M1",
			Lessons: []contract.Lesson{
				{Title: "L1", Type: "video", Price: 0, IsFree: true},
			},
		}},
	}
	ValidateCourse(v, c)
	if !v.Valid() {
		t.Fatalf("expected valid, got %v", v.Errors)
	}

	bad := web.NewValidator()
	badCourse := &contract.CourseView{
		Title: "", Lang: "fr", Price: -1,
		Modules: []contract.Module{{
			Title:   "",
			Lessons: []contract.Lesson{{Title: "", Type: "audio", IsFree: true, Price: 5}},
		}},
	}
	ValidateCourse(bad, badCourse)
	for _, key := range []string{"title", "lang", "price", "modules[0].title",
		"modules[0].lessons[0].title", "modules[0].lessons[0].type", "modules[0].lessons[0].price"} {
		if _, ok := bad.Errors[key]; !ok {
			t.Errorf("expected error for %q", key)
		}
	}
}
