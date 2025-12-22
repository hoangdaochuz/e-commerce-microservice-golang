---
name: Test Infrastructure Strategy
overview: Implement a comprehensive test infrastructure with unit, integration, and E2E tests using Mockery for mock generation and GitHub Actions for CI/CD integration.
todos:
  - id: setup-mockery
    content: Install Mockery, create mockery.yaml config, extract interfaces from concrete types
    status: completed
  - id: testutil-package
    content: Create testutil/ package with shared helpers, fixtures, and gRPC test utilities
    status: completed
  - id: unit-tests
    content: Implement unit tests for middleware, handlers, services, and shared packages
    status: completed
  - id: integration-tests
    content: Implement gRPC integration tests using bufconn for order and auth services
    status: completed
  - id: e2e-tests
    content: Implement E2E tests via API Gateway for authentication and order flows
    status: completed
  - id: ci-pipeline
    content: Create GitHub Actions workflow for automated testing and coverage reporting
    status: completed
  - id: taskfile-commands
    content: Add test-related tasks to taskfile.yml (unit, integration, e2e, coverage, mocks)
    status: completed
---

# Test Infrastructure Implementation Plan

## Overview

Implement a multi-layered test strategy covering unit tests (with Mockery-generated mocks), integration tests (gRPC/HTTP), and E2E tests (full system), with GitHub Actions CI/CD pipeline.

## Architecture

```mermaid
flowchart TB
    subgraph TestLayers [Test Layers]
        Unit[Unit Tests]
        Integration[Integration Tests]
        E2E[E2E Tests]
    end

    subgraph UnitScope [Unit Test Scope]
        Handlers[Handlers]
        Services[Services]
        Repos[Repositories]
        Middleware[Middleware]
        Pkg[Shared Packages]
    end

    subgraph IntegrationScope [Integration Test Scope]
        gRPC[gRPC Services]
        DB[Database Layer]
        Cache[Redis Cache]
    end

    subgraph E2EScope [E2E Test Scope]
        Gateway[API Gateway]
        FullFlow[Full Request Flow]
    end

    Unit --> UnitScope
    Integration --> IntegrationScope
    E2E --> E2EScope
```



## Directory Structure

```javascript
├── mocks/                         # Auto-generated mocks (gitignored)
├── testutil/                      # Shared test utilities
│   ├── fixtures/                  # Test data fixtures
│   ├── helpers.go                 # Common test helpers
│   └── grpc_helpers.go            # gRPC testing utilities
├── apps/
│   ├── order/
│   │   ├── handler/order/order_test.go
│   │   ├── services/order/order_service_test.go
│   │   └── repository/repo_test.go
│   └── auth/
│       ├── handler/auth/auth_test.go
│       └── services/auth/auth_test.go
├── api_gateway/
│   ├── middleware_test.go
│   └── integration_test.go
└── .github/workflows/test.yml     # CI/CD pipeline
```



## Implementation Steps

### 1. Setup Mockery and Generate Interfaces

Install Mockery and create a `mockery.yaml` config at the project root. Extract interfaces from concrete types where needed (e.g., `OrderRepository` interface from the struct).Key files to modify:

- [`apps/order/repository/repo.go`](apps/order/repository/repo.go) - Add `OrderRepositoryInterface`
- [`apps/order/services/order/order_service.go`](apps/order/services/order/order_service.go) - Add `OrderServiceInterface`

### 2. Create Test Utilities Package

Create `testutil/` with shared helpers:

- `helpers.go` - Common assertions, context builders
- `grpc_helpers.go` - gRPC test server setup, mock clients
- `fixtures/` - JSON/Go test data fixtures

### 3. Implement Unit Tests

**Priority targets:**

- [`api_gateway/middleware.go`](api_gateway/middleware.go) - Test each middleware in isolation
- [`apps/order/handler/order/order.go`](apps/order/handler/order/order.go) - Mock service layer
- [`apps/order/services/order/order_service.go`](apps/order/services/order/order_service.go) - Mock repository
- [`pkg/circuitbreaker/`](pkg/circuitbreaker/) - Expand existing tests
- [`pkg/cache/`](pkg/cache/) - Test cache implementations

### 4. Implement Integration Tests

Test real gRPC connections using `bufconn` (in-memory gRPC):

- Order service gRPC endpoint tests
- Auth service gRPC endpoint tests
- API Gateway routing integration

### 5. Implement E2E Tests

Full system tests via API Gateway HTTP endpoints:

- Authentication flow (Login -> Callback -> Protected endpoint)
- Order creation flow
- Rate limiting behavior

### 6. CI/CD Pipeline (GitHub Actions)

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
    - name: Generate mocks
        run: go install github.com/vektra/mockery/v2@latest && mockery
    - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...
    - name: Upload coverage
        uses: codecov/codecov-action@v4
```



### 7. Add Taskfile Commands

Extend [`taskfile.yml`](taskfile.yml) with test tasks:

- `task test:unit` - Run unit tests only
- `task test:integration` - Run integration tests
- `task test:e2e` - Run E2E tests