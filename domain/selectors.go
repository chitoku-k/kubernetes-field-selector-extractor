package domain

type FieldSelector struct {
	File   string   `json:"file"`
	Kind   string   `json:"kind"`
	Labels []string `json:"labels"`
}
