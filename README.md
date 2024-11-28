# Numeris Book Backend Test

The test application is implemented in to handle invoicing service for the test.
The test application is implemented in such a way that it can be run with no external dependencies
This test application is build with Mux, and Mongodb.

### User Interface
The user interface application is located here:
[https://www.figma.com/design/HydD9vrj5ScAlC2MPeB7ku/Assessment?node-id=1-7889&node-type=frame&t=GPXIg163hidjLLdz-0](https://www.figma.com/design/HydD9vrj5ScAlC2MPeB7ku/Assessment?node-id=1-7889&node-type=frame&t=GPXIg163hidjLLdz-0)


### Setup & Installation

clone the project and run:
```sh
go mod tidy
```

You can also install MongoDB.

Once the project is installed, create a `.env` file at the root of the project, with the following variables:

| Key                  | Description                 |          |
|:---------------------|:----------------------------|:---------|
| MONGO_URL            | Your mongodb connection url | Required |
| DB_NAME              | Database name               | Required |

To run the application, execute the following command from the root directory of the application:
```sh
    go run main.go
```

### Running Test

ensure your docker is running at the background because of the test db

To run test cases for the application, execute the following command from the root directory of the application:
```sh
    go test ./... -v
```

### Completed Tasks
The rest endpoints of the app can be tested as a standalone using `postman`.
This backend application implements the following tasks, per the ui requirements:

- [x] **Go Mux Setup
    - [x] Initialize a new Mux project.
    - [x] Install required dependencies (mux, mongodb).
    - [x] Set up a basic project structure.

- [x] **Implement a numericbookcore
    - [x] Implement a core system that generate a crud request automatically.
    - [x] Implement a core repository system that generate a repository operations like create, read, update and delete.

- [x] **Authentication**
    - [x] Implement JWT authentication/authorization middleware.
    - [x] Secure all endpoints with this authentication.

- [x] **Customer Management**
    - [x] Implemented endpoints to register, fetch, list, all user's customers.

- [x] **Invoice Configuration**
    - [x] Implemented endpoints to create invoices configuration.
    - [x] Implemented endpoints to update invoices configuration.

- [x] **Invoice management**
    - [x] Implemented endpoints to create invoices.
    - [x] Implemented functionalities to log invoice actions.
    - [x] Implemented functionalities to create invoice reminder time.
    - [x] Implemented functionalities to fetch audit logs for invoices.
    - [x] Implemented functionalities to fetch metrics

### API ROUTE
    - [x] /v1/auth/login --- login endpoint
    - [x] /v1/auth/signup --- signup endpoint
    - [x] v1/customer/register --- register customer
    - [x] v1/customer --- fetch customer
    - [x] v1/invoice --- fetch invoice
    - [x] v1/invoice/register--- register customer
    - [x] /v1/activity-log --- fetch all logs for a particular user
    - [x] /v1/activity-log?filter_by={"invoice_id": {"$eq": "674882e58ef79ee9bad633ce"}} --- filter activity log by specific invoice 
    - [x] /v1/metric/invoice --- dashboard metrics
