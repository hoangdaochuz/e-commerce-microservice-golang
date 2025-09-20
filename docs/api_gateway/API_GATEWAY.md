# API Gateway Documentation

## Architecture Overview

The API Gateway serves as a central entry point for all client requests in our microservices architecture. It handles routing, request/response transformation, and communication with backend services via NATS messaging system.

## Components

### 1. API Gateway Core
- Main gateway struct (`APIGateway`) that holds:
  - NATS connection
  - HTTP server
  - Request timeout settings
  - Router configuration

### 2. Middleware Layer
The gateway implements several middleware functions:
- Logging middleware - Logs request method, path and timing
- CORS middleware - Handles cross-origin resource sharing
- Content-Type middleware - Sets JSON content type headers

### 3. NATS Communication Layer
- Transforms HTTP requests to NATS messages
- Routes messages to appropriate services
- Handles responses and transforms them back to HTTP

## Request Flow

1. **Client Request**
   - Client sends HTTP request to API Gateway
   - Request passes through middleware chain

2. **Request Transformation**
   - HTTP request is converted to NATS format
   - Headers, body, and metadata are preserved
   - Subject is determined for routing

3. **NATS Communication**
   - Request is published to appropriate NATS subject
   - Gateway waits for response with configured timeout
   - Backend services process request and send response

4. **Response Handling**
   - NATS response is received
   - Response is transformed back to HTTP format
   - Headers and status codes are mapped
   - Response is sent back to client

## Error Handling

The gateway handles several types of errors:
- Request transformation errors
- NATS communication timeouts
- Backend service errors
- Response transformation errors

All errors are converted to appropriate HTTP status codes and error messages.

## Configuration

Key configuration parameters:
- NATS connection settings
- Service timeout duration (default 30 seconds)
- CORS settings
- Content type defaults

## Security

The gateway implements:
- CORS protection
- Request validation
- Error message sanitization
- Standard security headers

## Service Discovery

Services are discovered through NATS subjects, allowing for:
- Dynamic service registration
- Load balancing
- Service redundancy
- Automatic failover

## Benefits

1. **Centralized Control**
   - Single entry point for all API requests
   - Consistent security and monitoring
   - Unified error handling

2. **Service Decoupling**
   - Backend services can be modified without affecting clients
   - Protocol translation between HTTP and NATS
   - Independent scaling of gateway and services

3. **Enhanced Security**
   - Centralized authentication
   - Request validation
   - Response sanitization

4. **Monitoring and Logging**
   - Centralized request logging
   - Performance monitoring
   - Error tracking
