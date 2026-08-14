-- name: CreateProduct :one
INSERT INTO products (name, price, stock, category_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProducts :many
SELECT * FROM products;

-- name: GetProductById :one
SELECT * FROM products WHERE id = $1;

-- name: UpdateProduct :exec
UPDATE products SET name = $1, price = $2, stock = $3, category_id = $4, updated_at = now();

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;

-- name: AddProductToCategory :exec
UPDATE products SET category_id = $1 WHERE id = $2;