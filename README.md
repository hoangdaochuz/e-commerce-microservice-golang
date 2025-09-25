# E-commerce Microservice in Go

This is a microservices-based e-commerce application built with Go, implementing modern cloud-native patterns and best practices.

## Features

- Microservices architecture
- gRPC for inter-service communication
- NATS for asynchronous messaging
- PostgreSQL database with sqlx
- API Gateway pattern
- Frontend code generation for TypeScript/React Query

## Project Structure

## Project Architecture
We have designed the database schema in folder `docs/database-design` and the schema is in file `docs/database-design/schema/Ecommerce-db.sql`
Base on the schema, We applied DDD (Domain Driven Design) to design the project.
We have 11 main services that serve for specific domain business:
- `apps/product`: Product service
- `apps/shop`: Shop service
- `apps/order`: Order service
- `apps/user`: User service
- `apps/address`: Address service
- `apps/settings`: Settings service
- `apps/cart`: Cart service
- `apps/voucher`: Voucher service
- `apps/comment`: Comment service
- `apps/notification`: Notification service
- `apps/payment`: Payment service

Based on DDD(Domain Driven Design), we will aggregate the tables that have relation to business domain into a single service. 

We use pattern `Database per Service` to design the project. So each service has its own database. Let read for more detail:

### Product service:
#### Tables:
- `Products`
- `Category`
- `Category_product`
- `Inventory`
- `Product_voucher`

#### Database type:
- SQL (PostgreSQL)
#### Why?
- The complexity of the relation between tables and we need many join query to get the data.
- We need the transaction to ensure the data consistency. Ex: Update the inventory when user buy a product.

### Shop service:
#### Tables:
- `Shop_info`
#### Database type:
- SQL (PostgreSQL)
#### Why?
- The shop service needs to maintain ACID properties for shop information updates
- Shop data requires complex queries and joins with products and users
- SQL provides better data consistency and integrity which is important for shop management
- Shop data is structured and relationships need to be strictly enforced

### Order service:
#### Tables:
- `Order`
- `Order_items`
#### Database type:
- SQL (PostgreSQL)
#### Why?
- Order data requires ACID transactions to maintain data consistency when processing orders
- Complex joins needed between orders, order items, products and users
- SQL provides better support for handling financial transactions and order history
- Order data is highly structured with clear relationships that need to be enforced
- Need strong consistency guarantees for order processing and payment handling

### User service:
#### Tables:
- `User`
- `Profile`
#### Database type:
- NoSQL (MongoDB)
#### Why?
- User data is unstructured and requires flexible schema for different user types
- User data is highly dynamic and requires flexible schema for different user types

### Address service:
#### Tables:
- `Address`
#### Database type:
- SQL (PostgreSQL)
#### Why?
- Address data is highly structured and requires clear relationships between users and addresses

### Settings service:
#### Tables:
- `Setting`
#### Database type:
- NoSQL (MongoDB)
#### Why?
- Settings data need to be stored in a flexible schema to support different user types
- It's easy to change the schema of the settings data

### Cart service:
#### Tables:
- `Shopping_carts`
#### Database type:
- NoSQL (Redis) 
- SQL (PostgreSQL)
#### Why?
- The cart change frequently and we need to update the cart data frequently.
- But, We also need store the backup of the cart data for the user to analyze the cart data and the behavior of the user -> we also to support suggestion service for user.

### Voucher service:
#### Tables:
- `Voucher`
#### Database type:
- SQL (PostgreSQL)
#### Why?
- The relation between product and voucher is many to many.

### Comment service:
#### Tables:
- `Comment`
#### Database type:
- NoSQL (MongoDB)
#### Why?
- Comments often include unstructured data like text, images, and ratings
- Comments can be nested (replies to comments) which is well-suited for document databases
- Comments don't require complex joins with other data
- MongoDB's flexible schema allows easy addition of new comment features
- High write throughput needed for popular products with many comments

### Notification service:
#### Tables:
- `Notification`
#### Database type:
- NoSQL (MongoDB)
#### Why?
- Notifications are often sent to users and require flexible schema for different user types
- No need transaction for notification

### Payment service:
#### Tables:
- `Payment`
#### Database type:
- SQL (PostgreSQL)
#### Why?
- Payment is a financial transaction and we need to ensure the data consistency.
- Need ensure the data consistency between payment and order.
## Run docker
- From root project
```
docker compose -f ./infra/docker-compose.yml -p e-commerce-microservice-golang up
```

## Run the project
```
go run cmd/api_gateway/main.go
```


```
go run cmd/api_gateway/main.go
```

## How to generate .pb.go from .proto file
```
task backend:codegen -- apps/order/proto/order.proto
```