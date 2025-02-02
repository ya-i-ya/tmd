
## Quick Overview

- **Telegram**: Authenticates via phone number (and optional 2FA).
- **Postgres**: Holds all message and media references.
- **MinIO**: Stores files (photos, documents, audio).
- **Docker Compose**: Provides out-of-the-box containers for Postgres & MinIO.
---

## Prerequisites

1. **Go** (1.19+ recommended)
2. **Docker** & **Docker Compose**
3. **Telegram** account with API credentials from [my.telegram.org](https://my.telegram.org/).
---

## Configuration

All **instructions** and **fields** are already documented in [`example.config.yaml`](./example.config.yaml).  
Please open it and edit the following sections to match your environment:

- **`telegram`**: Phone number, `api_id`, `api_hash`, and optional `password` for 2FA.
- **`download`**: Where media is temporarily loaded to ram before uploading to MinIO.
- **`logging`**: Path for logs, file rotation, log level, etc.
- **`fetching`**: Dialog/message limits.
- **`minio`**: Host, credentials, bucket name, SSL usage.
- **`database`**: Postgres connection details (host, port, user, password, database name).
- **`rename example.config.yaml to config.yaml`**
---

## Getting Started

   cd docker
   docker-compose up -d