# Routes API Documentation

This document describes all currently registered server routes.

## Base

- Root health route:
  - `GET /`
  - Response: plain text `hello from homesever`

## Authentication

Routes under `/auth` do not require JWT.

### Signup

- Method: `POST`
- Path: `/auth/signup`
- Body:

```json
{
  "name": "Siddharth",
  "username": "sid",
  "password": "secret"
}
```

- Success (`201`):

```json
{
  "token": "<jwt>"
}
```

- Errors:
  - `400`: `{ "error": "Invalid request body" }`
  - `500`: `{ "error": "Failed to create user" }` or `{ "error": "Failed to issue token" }`

### Login

- Method: `POST`
- Path: `/auth/login`
- Body:

```json
{
  "username": "sid",
  "password": "secret"
}
```

- Success (`200`):

```json
{
  "token": "<jwt>"
}
```

- Errors:
  - `400`: `{ "error": "Invalid request body" }`
  - `401`: `{ "error": "Invalid username or password" }`
  - `500`: `{ "error": "Failed to issue token" }`

## Protected Routes

All routes below require auth middleware and a valid bearer token.

## Nodes (`/nodes`)

### List nodes

- Method: `GET`
- Path: `/nodes/`
- Query params:
  - `parent_id` (optional UUID)

- Success (`200`):

```json
{
  "nodes": [
    {
      "id": "<uuid>",
      "parent_id": null,
      "user_id": "<uuid>",
      "name": "docs",
      "type": "directory",
      "size": 0,
      "storage_id": null
    }
  ]
}
```

- Errors:
  - `400`: `{ "error": "invalid parent_id" }`
  - `401`: `{ "error": "unauthorized" }`
  - `500`: `{ "error": "failed to list nodes" }`

### Rename node

- Method: `PATCH`
- Path: `/nodes/:id`
- Body:

```json
{
  "new_name": "new-file-name.txt"
}
```

- Success (`200`):

```json
{
  "id": "<uuid>",
  "name": "new-file-name.txt",
  "message": "node renamed successfully"
}
```

- Errors:
  - `400`: `{ "error": "invalid node id" }`, `{ "error": "invalid request body" }`, `{ "error": "name cannot be empty" }`
  - `401`: `{ "error": "unauthorized" }`
  - `404`: `{ "error": "node not found" }`
  - `409`: `{ "error": "a node with this name already exists in this directory" }`
  - `500`: `{ "error": "failed to rename node" }`

### Delete node

- Method: `DELETE`
- Path: `/nodes/:id`

- Success (`200`):

```json
{
  "message": "node deleted successfully"
}
```

- Errors:
  - `400`: `{ "error": "invalid node id" }`
  - `401`: `{ "error": "unauthorized" }`
  - `404`: `{ "error": "node not found" }`
  - `500`: `{ "error": "failed to delete node" }`

## Directories (`/directories`)

### Create directory

- Method: `POST`
- Path: `/directories/`
- Body:

```json
{
  "name": "photos",
  "parent_id": "<optional-uuid>"
}
```

- Success (`201`):

```json
{
  "id": "<uuid>",
  "name": "photos",
  "parent_id": "<uuid-or-null>",
  "type": "directory",
  "created_at": 1711111111
}
```

- Errors:
  - `400`: `{ "error": "invalid request body" }`, `{ "error": "name cannot be empty" }`, `{ "error": "invalid user id" }`
  - `401`: `{ "error": "unauthorized" }`
  - `500`: `{ "error": "failed to create directory" }`

## Files (`/files`)

### Upload file (root)

- Method: `POST`
- Path: `/files/upload`
- Content-Type: `multipart/form-data`
- Form fields:
  - `file` (required)

### Upload file (inside directory)

- Method: `POST`
- Path: `/files/upload/:parentID`
- Content-Type: `multipart/form-data`
- Path params:
  - `parentID` (UUID or `null`)
- Form fields:
  - `file` (required)

Both upload endpoints return:

- Success (`201`):

```json
{
  "id": "<uuid>",
  "name": "myfile.pdf",
  "parent_id": "<uuid-or-null>",
  "type": "file",
  "size": 12345,
  "storage_id": "<uuid>"
}
```

- Errors:
  - `400`: `{ "error": "invalid parent id" }`, `{ "error": "file is required" }`, `{ "error": "invalid file" }`, `{ "error": "file name cannot be empty" }`, `{ "error": "invalid user id" }`
  - `401`: `{ "error": "unauthorized" }`
  - `409`: `{ "error": "no active storage available" }`, `{ "error": "insufficient disk space" }`
  - `500`: `{ "error": "failed to upload file" }`

### Download file

- Method: `GET`
- Path: `/files/download/:fileID`
- Path params:
  - `fileID` (UUID)

- Success (`200`):
  - Binary file stream
  - `Content-Disposition` attachment with original filename

- Errors:
  - `400`: `{ "error": "invalid file id" }`, `{ "error": "node is not a file" }`
  - `401`: `{ "error": "unauthorized" }`
  - `404`: `{ "error": "file not found" }`, `{ "error": "file storage not found" }`, `{ "error": "file is missing on disk" }`
  - `500`: `{ "error": "failed to download file" }`

## Quick Curl Examples

```bash
# signup
curl -X POST http://localhost:3000/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"name":"Siddharth","username":"sid","password":"secret"}'

# login
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"sid","password":"secret"}'

# list nodes
curl http://localhost:3000/nodes/ \
  -H "Authorization: Bearer <token>"

# create directory
curl -X POST http://localhost:3000/directories/ \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"photos"}'

# upload file
curl -X POST http://localhost:3000/files/upload \
  -H "Authorization: Bearer <token>" \
  -F "file=@/path/to/file.pdf"

# download file
curl -L http://localhost:3000/files/download/<file-uuid> \
  -H "Authorization: Bearer <token>" \
  -o downloaded-file
```
