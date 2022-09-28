# saga-example

Example of system with orchestrated saga:
- [transaction manager](http://localhost:36789)
- [service: orders](http://localhost:8090)
- [service: customers](http://localhost:8091)

## How to run

```bash
# run system
docker compose up --build

# open transaction manager UI
open http://localhost:36789

# create new customer
curl -X POST http://127.0.0.1:8090/create

# create new order
# response example {"gid":"GVLr2EBtAfUbfvANGybpdL"}
curl -X POST http://127.0.0.1:8091/create -H 'Content-Type: application/json' -d '{  
  "idCustomer": 1,
  "amount": 10,
  "currency": "BTC"
}'

# see GVLr2EBtAfUbfvANGybpdL in transaction manager
```

## Links

- [saga pattern](https://microservices.io/patterns/data/saga.html)
- [dtm transaction manager](https://en.dtm.pub)
