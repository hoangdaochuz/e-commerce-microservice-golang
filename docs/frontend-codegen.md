# Frontend Code Generation

Hệ thống tự động generate TypeScript/React Query code từ các service routes của API Gateway.

## Tính năng

- ✅ Generate constants cho các API endpoints  
- ✅ Hỗ trợ nhiều services
- ✅ Tự động detect service routes từ API Gateway
- 🔄 Generate TypeScript types (đang phát triển)
- 🔄 Generate React Query hooks (đang phát triển)

## Cách sử dụng

### 1. Cơ bản
```bash
# Generate với settings mặc định
task frontend:codegen
```

### 2. Với tham số tùy chỉnh
```bash
# Chỉ định output directory
task frontend:codegen OUTDIR=./web/src/api

# Chỉ định base URL cho production
task frontend:codegen BASEURL=https://api.myapp.com

# Chỉ định package name
task frontend:codegen PACKAGE=my-ecommerce-api

# Kết hợp nhiều tham số
task frontend:codegen OUTDIR=./frontend/api BASEURL=https://api.production.com PACKAGE=prod-api
```

### 3. Xem help
```bash
task frontend:codegen:help
```

### 4. Demo workflow
```bash
# Chạy demo hoàn chỉnh (backend + frontend codegen)
task demo

# Dọn dẹp files demo
task clean:demo
```

## Tham số

| Tham số | Mặc định | Mô tả |
|---------|----------|-------|
| `outdir` | `./frontend/apis` | Thư mục output cho code được generate |
| `service` | `api-client` | Tên package cho code được generate |

## Output

Code được generate sẽ bao gồm:

### 1. Constants (constant.ts)
```typescript
// This is codegen - DO NOT EDIT
// Code generated at: 2024-01-15 10:30:45

export const order_CreateOrder_URL = "/api/v1/order/CreateOrder";
```

### 2. Types (types.ts) - Coming Soon
```typescript
export interface CreateOrderRequest {
  customer_id: string;
}

export interface CreateOrderResponse {
  order_id: string;
}
```

### 3. Service Client (client.ts) - Coming Soon  
```typescript
export class OrderServiceClient {
  async createOrder(req: CreateOrderRequest): Promise<CreateOrderResponse> {
    // Implementation
  }
}
```

### 4. Index (index.ts)
```typescript
export * from './constant';
export * from './types';
export * from './client';
```

## Requirements

- NATS server đang chạy (cho service discovery)
- Các services đã được register trong API Gateway
- Go version 1.24+

## Troubleshooting

### "No service routes found"
- Đảm bảo NATS server đang chạy
- Kiểm tra config trong `configs/config.yaml`
- Đảm bảo các services đã được register đúng cách

### "Failed to connect to NATS"
- Kiểm tra NATS server status: `docker-compose ps`
- Verify credentials trong config
- Đảm bảo port 4222 không bị block

## Ví dụ sử dụng

```bash
# 1. Generate cho tất cả service
task frontend:codegen:all

# 3. Generate cho từng service cụ thể
task frontend:codegen:service service=ecommerce-client outdir=./src/services
``` 