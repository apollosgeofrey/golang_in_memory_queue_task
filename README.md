# In-Memory Queue Broker (Go)

A minimal FIFO message queue exposed as an HTTP web service. Built with Go standard library only.

## Quick Start

```bash
# Run the server on port 8080
go run main.go 8080

# Or build and run
go build -o queue-broker main.go
./queue-broker 8080
```

The port is **required** and passed as the first command-line argument.

---

## API Reference

### PUT `/{queue}?v={message}`

Enqueue a message into the named queue.

| Case | Response |
|------|----------|
| Success | `200 OK`, empty body |
| Missing `v` parameter | `400 Bad Request`, empty body |

**Examples:**
```bash
curl -X PUT "http://127.0.0.1:8080/pet?v=cat"
curl -X PUT "http://127.0.0.1:8080/pet?v=dog"
curl -X PUT "http://127.0.0.1:8080/role?v=manager"
```

---

### GET `/{queue}?timeout={seconds}`

Dequeue the next message (FIFO). Returns the message as plain text in the response body.

| Case | Response |
|------|----------|
| Message available | `200 OK`, body = message text |
| Queue empty, no timeout | `404 Not Found`, empty body |
| Queue empty, timeout expires | `404 Not Found`, empty body |
| Invalid timeout value | `400 Bad Request`, empty body |

**`timeout`** (optional): wait up to N **seconds** for a message. If a message arrives during the wait, returns immediately with `200`.

**Examples:**
```bash
curl "http://127.0.0.1:8080/pet"              # immediate: cat (or 404 if empty)
curl "http://127.0.0.1:8080/pet?timeout=5"    # wait up to 5 seconds
curl "http://127.0.0.1:8080/pet"              # 404 when empty
```

---

## Full Spec Walkthrough

Run these in order after starting the server on port 8080:

```bash
# Enqueue messages
curl -X PUT "http://127.0.0.1:8080/pet?v=cat"
curl -X PUT "http://127.0.0.1:8080/pet?v=dog"
curl -X PUT "http://127.0.0.1:8080/role?v=manager"
curl -X PUT "http://127.0.0.1:8080/role?v=executive"

# Dequeue from pet (FIFO)
curl "http://127.0.0.1:8080/pet"    # → cat
curl "http://127.0.0.1:8080/pet"    # → dog
curl "http://127.0.0.1:8080/pet"    # → 404

# Dequeue from role
curl "http://127.0.0.1:8080/role"   # → manager
curl "http://127.0.0.1:8080/role"   # → executive
curl "http://127.0.0.1:8080/role"   # → 404
```

---

## Postman Setup

### 1. Start the server

```bash
go run main.go 8080
```

### 2. Create a collection (optional but helpful)

In Postman, create a new collection called **Queue Broker** with base URL variable:

| Variable | Value |
|----------|-------|
| `baseUrl` | `http://127.0.0.1:8080` |

### 3. PUT request — enqueue a message

| Field | Value |
|-------|-------|
| Method | `PUT` |
| URL | `{{baseUrl}}/pet?v=cat` |
| Body | None |
| Headers | None required |

Click **Send**. Expected: status **200 OK**, empty body.

To enqueue another message, change `v=cat` to `v=dog`, etc.

**Test missing parameter (400):**
- URL: `{{baseUrl}}/pet` (no `?v=...`)
- Expected: **400 Bad Request**

### 4. GET request — dequeue a message

| Field | Value |
|-------|-------|
| Method | `GET` |
| URL | `{{baseUrl}}/pet` |
| Body | None |

Click **Send**. Expected: status **200 OK**, body = `cat` (first message enqueued).

Repeat GET on `/pet` → `dog`, then **404** when empty.

### 5. GET with timeout — wait for a message

Use two Postman tabs (or Terminal + Postman):

**Tab 1 — GET (wait 10 seconds):**
| Field | Value |
|-------|-------|
| Method | `GET` |
| URL | `{{baseUrl}}/orders?timeout=10` |

Send this first. It will hang until a message arrives or 10 seconds pass.

**Tab 2 — PUT (within 10 seconds):**
| Field | Value |
|-------|-------|
| Method | `PUT` |
| URL | `{{baseUrl}}/orders?v=pizza` |

Send Tab 2 while Tab 1 is waiting.

**Result:** Tab 1 completes with **200 OK**, body = `pizza`.

If you never send the PUT, Tab 1 returns **404** after 10 seconds.

### 6. Verify FIFO ordering among waiters

1. Send GET `{{baseUrl}}/race?timeout=30` in Tab 1 (starts waiting).
2. Send GET `{{baseUrl}}/race?timeout=30` in Tab 2 (starts waiting second).
3. Send PUT `{{baseUrl}}/race?v=first` — Tab 1 gets `first`.
4. Send PUT `{{baseUrl}}/race?v=second` — Tab 2 gets `second`.

The first GET request always receives the first available message.

---

## Postman Tips

- **No request body** is needed for either method; everything is in the URL.
- Queue name is the **path segment** after the host: `/pet`, `/role`, `/orders`, etc.
- Response body on success is **plain text**, not JSON.
- For timeout tests, increase Postman request timeout: **Settings → General → Request timeout** (set to 0 or a high value so Postman doesn't abort before the server responds).

---

## Verify Locally (before submitting)

```bash
go test -v ./...
```

All 5 tests should pass:
- Spec walkthrough (PUT/GET FIFO, 404 on empty)
- Missing `v` → 400
- Timeout + late PUT → message delivered
- Timeout alone → 404
- Two waiters → first GET gets first message

---

## What to Submit

The company asked for **one file**: `main.go`. You can push the whole repo to GitHub, but the solution itself is just `main.go` (~140 lines, standard library only).

Share the GitHub link in the HH chat and mention how long you spent.
