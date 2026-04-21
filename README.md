# AI-Based Financial Application (Backend)

A financial application backend built using Go, adhering to Clean Architecture principles. It enables users to track finances, categorize transactions, generate statistics, and automatically parse bank statements (such as Kaspi Bank PDF statements).

## Features
- **User Authentication**: JWT-based secure login and registration.
- **Transaction Management**: Full CRUD operations to track income and expenses.
- **PDF Parsing**: Automated extraction and saving of bulk transactions directly from Kaspi Bank PDF statements.
- **Financial Statistics**: See balances and query time-period-based statistics (net flow, aggregate category info).
- **Architecture**: Designed with separation of concerns: Routers -> Handlers -> UseCases -> Repositories -> PostgreSQL mapping.

## Setup & Running

1. **Environment Config**: Create an `.env` file at the root of the project with your database settings:
   ```env
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=secret
   DB_NAME=finance_app
   JWT_SECRET=supersecret
   SERVER_PORT=8080
   ```
2. **Database Required**: A running PostgreSQL instance.
3. **Execute**:
   ```bash
   cd cmd/api
   go run main.go
   ```
   The backend API will start on `https://localhost:8080`.

---

## API Endpoints Reference

### Public Authentication
**Base Path:** `/api/auth`

- **POST** `/register` — Register a new account.
  ```json
  // Request Body
  {
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepassword123",
    "currency": "KZT"
  }
  ```

- **POST** `/login` — Login into the system to get a JWT token.
  ```json
  // Request Body
  {
    "email": "john@example.com",
    "password": "securepassword123"
  }
  ```

---

### Protected Routes
_All actions below require the HTTP Header to be passed: `Authorization: Bearer <your_jwt_token>`_

**Base Path:** `/api`

#### Profile
- **GET** `/profile` — Get the currently logged-in user's active profile and currency.
- **POST** `/profile` — Update user profile.

#### Transactions
- **POST** `/transactions` — Create a new transaction.
  ```json
  {
    "amount": 1500.50,
    "description": "Groceries",
    "category": "Food",
    "type": "expense",
    "created_at": "2023-10-12T15:00:00Z"
  }
  ```
- **GET** `/transactions` — Fetch all previous transactions of the user. 
  - Supports query param pagination: `?page=1&limit=10`
- **GET** `/transactions/:id` — Retrieve details for a single specific transaction by UUID.
- **PUT** `/transactions/:id` — Update transaction properties.
  ```json
  {
    "amount": 2000.00,
    "description": "Groceries and Snacks",
    "category": "Food",
    "created_at": "2023-10-12T15:00:00Z"
  }
  ```
- **DELETE** `/transactions/:id` — Permenantly remove a transaction.

#### Statement Integrations
- **POST** `/statements/upload` — Upload Kaspi PDF Bank statements.
  - **Type**: `multipart/form-data`
  - **Body Parameter**: `file` (Expecting `.pdf` type only)
  - _Description_: The internal PDF parser will comb through and extract all transaction rows and apply it sequentially to the user's transaction book natively.

#### Statistics & Reporting
- **GET** `/statistics/balance` — Fetches standard account current balances.
- **GET** `/statistics` — Get detailed financial insights and aggregations.
  - **Available Query Params**: `?period_start=2023-01-01&period_end=2023-12-31`
  - **Sample Response**:
    ```json
    {
      "balance": {
         "total": 120000.0,
         "currency": "KZT",
         "updated_at": "2023-10-12T00:00:00Z"
      },
      "income": 50000.0,
      "expenses": 15000.0,
      "net_flow": 35000.0,
      "expense_categories": [{"category": "Food", "amount": 10000.0}],
      "income_categories": [{"category": "Salary", "amount": 50000.0}],
      "period_start": "2023-01-01T00:00:00Z",
      "period_end": "2023-12-31T23:59:59Z"
    }
    ```

#### Goals Tracking
- **POST** `/goals` — Create a new financial goal.
  ```json
  {
    "title": "New Laptop",
    "target_amount": 1500.00,
    "deadline": "2024-12-31T00:00:00Z"
  }
  ```
- **GET** `/goals` — Fetch all goals for the logged-in user.
- **GET** `/goals/:id` — Retrieve details for a single specific goal.
- **PUT** `/goals/:id/contribute` — Add progress (money) to a specific goal.
  ```json
  {
    "amount": 50.00
  }
  ```
- **DELETE** `/goals/:id` — Permanently delete a goal.

#### Balance Management
- **PUT** `/statistics/balance` — Update the user's base opening balance.
  ```json
  {
    "total": 5000.00
  }
  ```
