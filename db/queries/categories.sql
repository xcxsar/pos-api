-- name: CreateCategory :one
INSERT INTO categories (name)
VALUES ($1)
RETURNING *;

-- name: GetCategories :many
SELECT * FROM categories;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1;

-- name: UpdateCategory :one
UPDATE categories SET name = $1, updated_at = now()
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;

-- name: GetProductsByCategory :many
SELECT * FROM products WHERE category_id = $1;