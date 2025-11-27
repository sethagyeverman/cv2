package minio

type UploadFile struct {
	Filename string `json:"filename"`
	Content  []byte `json:"content"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
}
