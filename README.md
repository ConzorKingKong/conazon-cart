# Conazon Cart

This is the cart endpoint for the Conazon project.

## Quickstart

To test locally, setup a `.env` file in the root directory with the following variables:

`DATABASEURL` - Url to postgres database. REQUIRED
`SECRET` - JWT secret. Must match the secret used in the auth service REQUIRED
`PORT` - Port to run server on. Defaults to 8082

Datbase url should be formatted like this if using `docker-compose up` - 'host=postgres port=5432 user=postgres dbname=conazon sslmode=disable'

Then run:

`docker-compose up`

## Endpoints (later will have swagger)

- /

GET - generic hello world. useless endpoint

- /cart

POST - Add items to cart

{"product_id": INT, "quantity": INT}

- /cart/{id}

GET - Returns cart id entry

PATCH - Updates item in cart

{"quantity": INT}

DELETE - Deletes item from cart (sets status to deleted)

- /cart/user/{id}

GET - get all active items in users cart

PUT - Move all active cart items to purchased after purchase
