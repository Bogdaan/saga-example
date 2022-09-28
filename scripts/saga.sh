#!/usr/bin/env bash

# create new customer
curl -X POST http://127.0.0.1:8090/create

# create valid order
curl -X POST http://127.0.0.1:8091/create -H 'Content-Type: application/json' -d '{
  "idCustomer": 1,
  "amount": 98,
  "currency": "BTC"
}'

# create invalid order
curl -X POST http://127.0.0.1:8091/create -H 'Content-Type: application/json' -d '{
  "idCustomer": 1,
  "amount": 1,
  "currency": "BTC"
}'

curl -X POST http://127.0.0.1:8091/create -H 'Content-Type: application/json' -d '{
  "idCustomer": 1,
  "amount": 1,
  "currency": "BTC"
}'

###
#
## create new customer
#curl -X POST http://127.0.0.1:8090/create
#
## create order
#curl -X POST http://127.0.0.1:8091/create -H 'Content-Type: application/json' -d '{
#  "idCustomer": 2,
#  "amount": 99,
#  "currency": "BTC"
#}'
#
## create order
#curl -X POST http://127.0.0.1:8091/create -H 'Content-Type: application/json' -d '{
#  "idCustomer": 2,
#  "amount": 99,
#  "currency": "BTC"
#}'
