# Conazon Cart

This is the cart endpoint for the Conazon project.

## Quickstart

To test locally, setup a `.env` file in the root directory with the following variables:

```
JWTSECRET - JWT secret. Must match the secret used in the auth service REQUIRED
DATABASEURL - Url to postgres database. REQUIRED
PORT - Port to run server on. Defaults to 8082
```

Datbase url should be formatted like this if using `docker-compose` - `'host=postgres port=5432 user=postgres dbname=conazon sslmode=disable'`

Then run:

`docker-compose up`

## Endpoints (later will have swagger)

- /

GET - Catch all 404

- /cart

GET - get all active carts for user PROTECTED

POST - create new cart/item combo PROTECTED

{"product_id": INT, "quantity": INT}

PATCH - mark all active items as purchased PROTECTED (This is very illegal and will be changed later to take a userId and be done by the checkout microservice)

DELETE - delete all active cart/item combos for user PROTECTED

- /cart/{id}

GET - Returns cart id entry (see if cart/item combo is active or inactive) PROTECTED

PATCH - Updates item quantity in cart PROTECTED

{"quantity": INT}

DELETE - Deletes item from cart (sets status to deleted)