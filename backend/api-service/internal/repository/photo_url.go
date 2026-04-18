package repository

import "strings"

// minioEndpoint задаётся из конфига при старте (SetMinioEndpoint).
// Используется для переписывания внутренних URL фото (http://minio:9000/...) в
// относительный путь same-origin (/minio/...), чтобы браузер подгружал фото
// через nginx-проксирование, не видя внутреннего docker-хоста.
var minioEndpoint string

// SetMinioEndpoint вызывается один раз при старте сервиса.
func SetMinioEndpoint(endpoint string) {
	minioEndpoint = endpoint
}

// rewritePhotoURL превращает http://minio:9000/bucket/key → /minio/bucket/key.
// Если URL не соответствует паттерну — возвращается как есть.
func rewritePhotoURL(url string) string {
	if minioEndpoint == "" || url == "" {
		return url
	}
	for _, prefix := range []string{"http://" + minioEndpoint, "https://" + minioEndpoint} {
		if strings.HasPrefix(url, prefix) {
			return "/minio" + strings.TrimPrefix(url, prefix)
		}
	}
	return url
}

// rewritePhotoURLPtr — то же, но для указателя (когда поле nullable).
func rewritePhotoURLPtr(url *string) {
	if url != nil && *url != "" {
		s := rewritePhotoURL(*url)
		*url = s
	}
}
