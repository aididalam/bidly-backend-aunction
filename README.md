Root project: [bidly](https://github.com/aididalam/bidly)

# Endpoints

| Method | Path | Authentication |
|---|---|---|
| POST | `/api/products` | Bearer JWT |
| GET | `/api/products` | No |
| GET | `/api/products/{product_id}` | No |
| GET | `/api/me/products` | Bearer JWT |
| PUT | `/api/products/{product_id}` | Bearer JWT |
| DELETE | `/api/products/{product_id}` | Bearer JWT |
| POST | `/api/products/{product_id}/bids` | Bearer JWT |
| GET | `/api/products/{product_id}/bids` | Bearer JWT |
| POST | `/api/products/{product_id}/sell` | Bearer JWT |
| POST | `/api/uploads/presigned-url` | Bearer JWT |

# Structure

```text
auction/
├── cmd/api/
├── internal/
│   ├── config/
│   ├── handler/
│   ├── middleware/
│   ├── model/
│   ├── repository/
│   ├── service/
│   ├── token/
│   ├── upload/
│   └── worker/
└── migrations/
```
