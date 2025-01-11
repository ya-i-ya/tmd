
- **Telegram**: The app authenticates via phone number and optionally 2FA.
- **Media Downloads**: Files (photos, documents, audio, etc.) are fetched and saved locally.
- **MinIO Upload**: Local files are uploaded to MinIO for off-box storage, and a link is stored in the database.
- **Postgres Database**: Stores message data, including references to the media in MinIO.

## Prerequisites

1. **Telegram API credentials** (phone number, API ID, API hash). Obtain them at [my.telegram.org](https://my.telegram.org/) if you haven’t already.
2. **Docker**: Docker & Docker Compose, for MinIO and postgres DB.
