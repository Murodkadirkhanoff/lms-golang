// Package uidefaults frontend talab qiladigan, lekin DB'da saqlanmaydigan
// UI maydonlar (avatarColor, thumbnailColor) uchun deterministik defaultlar.
// Palitra frontend mock ma'lumotlaridagi Tailwind klasslariga mos.
package uidefaults

var palette = []string{
	"bg-indigo-200",
	"bg-amber-200",
	"bg-rose-200",
	"bg-emerald-200",
	"bg-sky-200",
	"bg-fuchsia-200",
}

func AvatarColor(id int64) string {
	return palette[id%int64(len(palette))]
}

func ThumbnailColor(id int64) string {
	return palette[(id+3)%int64(len(palette))]
}
