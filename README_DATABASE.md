# Database Integration - Quick Start 🚀

## ✅ What's Been Set Up

1. ✅ **Prisma Schema** - Database schema definition
2. ✅ **GORM Models** - Go database models
3. ✅ **Database Connection** - Supabase PostgreSQL
4. ✅ **Auto-migration** - Tables created automatically
5. ✅ **gRPC Integration** - Server uses database

---

## 🎯 Quick Setup (2 Steps)

### Step 1: Create .env file

```bash
./setup-db.sh
```

Enter your Supabase password when prompted.

**OR manually create `.env`:**

```bash
DATABASE_URL="postgresql://postgres.bvtsauqbkrsnyfrrayuh:[YOUR-PASSWORD]@aws-1-ap-northeast-2.pooler.supabase.com:6543/postgres?pgbouncer=true"
DIRECT_URL="postgresql://postgres.bvtsauqbkrsnyfrrayuh:[YOUR-PASSWORD]@aws-1-ap-northeast-2.pooler.supabase.com:5432/postgres"
```

### Step 2: Run migrations

```bash
./migrate-db.sh
```

**OR the server will auto-migrate on startup!**

---

## 🚀 Start Server

```bash
cd server
go run main.go database.go
```

You should see:
```
✅ Connected to database successfully
✅ Database migration completed
gRPC server listening on :8080
✅ Database connected and ready
```

---

## 📊 What Happens Now

### When you call Unary API:

1. Server receives name (e.g., "Rohith")
2. **Finds or creates user** in `users` table
3. **Creates greeting record** in `greetings` table
4. Returns response

### Data is persisted:
- ✅ Users are saved
- ✅ Greetings are tracked
- ✅ Relationships maintained

---

## 🔍 Verify It Works

### Test the API:

```bash
# From frontend or using curl
curl -X POST http://localhost:8081/api/unary \
  -H "Content-Type: application/json" \
  -d '{"name":"Rohith"}'
```

### Check database:

Go to Supabase SQL Editor and run:
```sql
SELECT * FROM users;
SELECT * FROM greetings;
```

You should see your data!

---

## 📁 Files Created

```
prisma/
└── schema.prisma          # Schema definition

server/
├── database.go            # Database models & connection
├── migrate.go             # Migration script
└── main.go                # Updated with DB integration

.env                       # Database credentials (create this)
setup-db.sh                # Setup script
migrate-db.sh              # Migration script
```

---

## 🎉 You're Done!

Your gRPC server now:
- ✅ Connects to Supabase
- ✅ Creates tables automatically
- ✅ Saves user data
- ✅ Tracks greetings
- ✅ Uses connection pooling (fast!)

**Next:** Test your APIs and see data in Supabase! 🚀

