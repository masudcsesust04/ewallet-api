# Ewallet API application


## Application overview
This project provies a RESTful API server to implement a ewallet.

## Project Setup
- Go 1.23.2 or higher 
- A PostgreSQL database

### Installation
1. Clone the code repository:
```bash
git clone <repository-url>
cd ewallet-api
```

2. Install dependencies:
```bash
go mod download
```

3. Database setup
- Create a database in your PostgreSQL server.
- Run the SQL schema file `schem.sql` to create necessary table and indices.

4. Set `DATABASE_URL` to environment variable:
```bash
export DATABASE_URL=postgres://user_name:password@127.0.0.1:5432/database_name?sslmode=disable
```

## Running the server:
To start the server, run:
```bash
go run cmd/server/main.go
```

The server will start on port `8080`

API End point testing:
1. Create user:
```bash
curl -X POST http://localhost:8080/users \
-H "Content-Type: application/json" \
-d '{"first_name": "masud", "last_name": "rana", "phone_number": "8098908080", "email": "masud@gmail.com", "password": "example123", "status": "active"}'
```

2. List of user:
```bash
curl -X GET http://localhost:8080/users
```

3. Get user by id:
```bash
curl -X GET http://localhost:8080/users/1
```

4. Update user:
```bash
curl -X PUT http://localhost:8080/users/1 \
-d '{"phone_number": "8098908080"}'
```
5. Delete id:
```bash
curl -X DELETE http://localhost:8080/users/1
```

## Wallet & Transactions
1. Create new wallet:
```bash
curl -X POST http://localhost:8080/wallets/new \
-H "Content-Type: application/json" \
-d '{"user_id": 1, "balance": 200.0, "currency": "USD"}'
```

2. Deposit to wallet:
```bash
curl -X POST http://localhost:8080/wallets/deposit \
-H "Content-Type: application/json" \
-d '{"user_id": 1, "amount": 50.0}'
```

3. Withdraw from wallet:
```bash
curl -X POST http://localhost:8080/wallets/withdraw \
-H "Content-Type: application/json" \
-d '{"user_id": 1, "amount": 25.0}'
```

4. Fund trunsfer from one wallet to another:
```bash
curl -X POST http://localhost:8080/wallets/transfer \
-H "Content-Type: application/json" \
-d '{"from_wallet_id": 1, "to_wallet_id": 2, "amount": 25.0}'
```

5. Check wallet balance:
```bash
curl -X GET 'http://localhost:8080/wallets/balance?user_id=1'
```

6. Transactions history:
```bash
curl -X GET 'http://localhost:8080/wallets/transactions?wallet_id=1'
```

