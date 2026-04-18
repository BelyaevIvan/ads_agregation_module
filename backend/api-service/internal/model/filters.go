package model

// SizeFilters содержит все уникальные размеры, встречающиеся в активных
// (is_hidden = FALSE) объявлениях. Отсортированы численно по возрастанию.
type SizeFilters struct {
	SizeRus []string `json:"size_rus"`
	SizeEU  []string `json:"size_eu"`
	SizeUS  []string `json:"size_us"`
}
