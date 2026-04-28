package category

import "time"

type CreateCategoryRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Kind  string `json:"kind" validate:"required,oneof=income expense"`
	Color string `json:"color" validate:"required,hexcolor"`
}

type CategoryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
}

type CategoriesResponse struct {
	Items []CategoryResponse `json:"items"`
}

func ToCategoryResponse(c *Category) CategoryResponse {
	return CategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		Kind:      c.Kind,
		Color:     c.Color,
		CreatedAt: c.CreatedAt.UTC().Format(time.RFC3339),
	}
}
