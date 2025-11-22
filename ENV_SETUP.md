# Environment Setup Guide

## 📍 Where to Put .env File

### ✅ Correct Location: Project Root

```
grpc-example/
├── .env              ← Put it HERE (project root)
├── server/
│   └── main.go       ← Server reads from parent directory
├── gateway/
└── frontend/
```

### ❌ Wrong Location: server/ directory

```
grpc-example/
├── server/
│   ├── .env          ← DON'T put here
│   └── main.go
```

---

## 🚀 Quick Setup

### Step 1: Create .env in Project Root

```bash
cd "/media/rohith/New Volume/rohith/grpc-example"
nano .env
```

Add:
```bash
DATABASE_URL="postgresql://postgres.bvtsauqbkrsnyfrrayuh:[YOUR-PASSWORD]@aws-1-ap-northeast-2.pooler.supabase.com:6543/postgres?pgbouncer=true"
DIRECT_URL="postgresql://postgres.bvtsauqbkrsnyfrrayuh:[YOUR-PASSWORD]@aws-1-ap-northeast-2.pooler.supabase.com:5432/postgres"
```

### Step 2: Start Server

```bash
cd server
go run main.go database.go
```

**That's it!** The server will:
- ✅ Load .env automatically
- ✅ Connect to database
- ✅ Auto-migrate tables
- ✅ Start gRPC server

---

## ❌ Do You Need `npx prisma generate`?

### Answer: NO

**Why?**
- We're using **GORM** (Go ORM), not Prisma Client
- Prisma Client is for Node.js/TypeScript
- GORM doesn't need code generation
- Prisma schema is just for **reference/documentation**

### What We're Using:

```
Go Server → GORM → PostgreSQL
```

**NOT:**
```
Go Server → Prisma Client → PostgreSQL  ❌
```

---

## 🔄 How It Works

### 1. Server Starts
```bash
cd server
go run main.go database.go
```

### 2. Loads .env
- Looks for `.env` in project root (parent directory)
- Loads `DATABASE_URL` environment variable

### 3. Connects to Database
- Uses GORM to connect to Supabase
- Connection pooling enabled (fast!)

### 4. Auto-Migrates
- Creates `users` table if not exists
- Creates `greetings` table if not exists
- Sets up indexes and relationships

### 5. Ready!
- Server starts on port 8080
- Database ready for queries

---

## 📝 Environment Variables

### Required:
```bash
DATABASE_URL="postgresql://..."  # For queries (with pooling)
```

### Optional:
```bash
DIRECT_URL="postgresql://..."    # For migrations (if needed)
```

---

## 🧪 Test It

### 1. Create .env
```bash
echo 'DATABASE_URL="postgresql://postgres.bvtsauqbkrsnyfrrayuh:YOUR_PASSWORD@aws-1-ap-northeast-2.pooler.supabase.com:6543/postgres?pgbouncer=true"' > .env
```

### 2. Start Server
```bash
cd server
go run main.go database.go
```

### 3. Expected Output:
```
✅ Loaded .env file
✅ Connected to database successfully
✅ Database migration completed
gRPC server listening on :8080
✅ Database connected and ready
```

---

## 🚨 Troubleshooting

### Error: "DATABASE_URL not set"

**Solution:**
1. Check `.env` is in project root (not `server/`)
2. Check `.env` has `DATABASE_URL=...`
3. Check password is correct (no brackets)

### Error: "Could not load .env file"

**Solution:**
- Server will still work if you export environment variable:
```bash
export DATABASE_URL="postgresql://..."
cd server
go run main.go database.go
```

### Error: "Connection refused"

**Solution:**
- Check Supabase password is correct
- Verify Supabase project is active
- Check network/firewall

---

## ✅ Summary

1. ✅ Put `.env` in **project root** (not `server/`)
2. ✅ **NO need** for `npx prisma generate` (using GORM)
3. ✅ Server **auto-loads** .env file
4. ✅ Server **auto-migrates** on startup
5. ✅ Just start server: `go run main.go database.go`

**That's all you need!** 🚀

