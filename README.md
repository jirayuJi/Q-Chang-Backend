# Q-Chang-Backend
how to run
- open terminal run command "go build" and "./SGC-test-2 --port 8080"

path
[GET]
  - /healthcheck
  - /cashiers <br>
    ```curl --location --request GET 'localhost:8080/v1/cashiers?is_active=true&limit=10&page=1'```
  - /cash_log <br> action = top-up / order-payment  <br>
    ```curl --location --request GET 'localhost:8080/v1/cash_log?action=top-up'```
  - /order_log <br>
    ```curl --location --request GET 'localhost:8082/v1/order_log?limit=10&page=1'```<br>
[POST]
  - /cashier <br> for create cashier<br>
    ```curl --location --request POST 'localhost:8080/v1/cashier?cashier_id=12345&location=Ton%20Son%20Tower' --header 'Content-Type: application/x-www-form-urlencoded' --data-urlencode 'cashier_id=12346' --data-urlencode 'location=computerlogy' --data-urlencode 'is_active=true'```
  - /top_up <br> top up to cashier store <br>
    ```curl --location --request POST 'localhost:8080/v1/top_up' --header 'Content-Type: application/x-www-form-urlencoded' --data-urlencode 'cashier_id=12345' --data-urlencode 'balance=1,1,2' --data-urlencode 'type=bank_note,bank_note' --data-urlencode 'value=1000,500,100'```
  - /payment <br>
  ```curl --location --request POST 'localhost:8080/v1/payment' --header 'Content-Type: application/x-www-form-urlencoded' --data-urlencode 'cashier_id=12345' --data-urlencode 'product_id=1' --data-urlencode 'quantity=1' --data-urlencode 'receive_cash=1000'```

