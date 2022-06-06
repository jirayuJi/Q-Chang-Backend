# Q-Chang-Backend
how to run
- open terminal run command "go build" and "./SGC-test-2 --port 8080"

path
[GET]
  - /healthcheck
  - /cashiers \n
    ```curl --location --request GET 'localhost:8080/v1/cashiers?is_active=true&limit=10&page=1'```
  - /cash_log
  - /order_log
  - /cash_log
[POST]
  - /cashier
  - /top_up
  - /payment

