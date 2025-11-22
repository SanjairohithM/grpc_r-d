# Fix Database Connection Error

## 🚨 Error: "invalid userinfo"

This error occurs when your password contains special characters that need URL encoding.

---

## ✅ Solution: URL-Encode Your Password

### Step 1: Check Your Password

If your Supabase password contains special characters like:
- `@`, `#`, `$`, `%`, `&`, `+`, `=`, `/`, `?`, `:`, `;`

They need to be URL-encoded.

### Step 2: URL-Encode Your Password

**Option A: Use Online Tool**
1. Go to: https://www.urlencoder.org/
2. Paste your password
3. Copy the encoded version

**Option B: Use Python (quick)**
```bash
python3 -c "import urllib.parse; print(urllib.parse.quote('YOUR_PASSWORD'))"
```

**Option C: Use Node.js**
```bash
node -e "console.log(encodeURIComponent('YOUR_PASSWORD'))"
```

### Step 3: Update .env File

Replace `[YOUR-PASSWORD]` with the **URL-encoded** password:

```bash
DATABASE_URL="postgresql://postgres.bvtsauqbkrsnyfrrayuh:URL_ENCODED_PASSWORD@aws-1-ap-northeast-2.pooler.supabase.com:6543/postgres?pgbouncer=true"
```

---

## 📝 Example

### If your password is: `MyP@ss#123`

**URL-encoded:** `MyP%40ss%23123`

**Updated DATABASE_URL:**
```bash
DATABASE_URL="postgresql://postgres.bvtsauqbkrsnyfrrayuh:MyP%40ss%23123@aws-1-ap-northeast-2.pooler.supabase.com:6543/postgres?pgbouncer=true"
```

---

## 🔍 Common Special Characters Encoding

| Character | Encoded |
|-----------|---------|
| `@` | `%40` |
| `#` | `%23` |
| `$` | `%24` |
| `%` | `%25` |
| `&` | `%26` |
| `+` | `%2B` |
| `=` | `%3D` |
| `/` | `%2F` |
| `?` | `%3F` |
| `:` | `%3A` |
| `;` | `%3B` |

---

## 🚀 Quick Fix Script

Create a file `encode-password.sh`:

```bash
#!/bin/bash
echo "Enter your Supabase password:"
read -s PASSWORD
ENCODED=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$PASSWORD'))")
echo ""
echo "URL-encoded password: $ENCODED"
echo ""
echo "Add this to your .env file:"
echo "DATABASE_URL=\"postgresql://postgres.bvtsauqbkrsnyfrrayuh:${ENCODED}@aws-1-ap-northeast-2.pooler.supabase.com:6543/postgres?pgbouncer=true\""
```

Run:
```bash
chmod +x encode-password.sh
./encode-password.sh
```

---

## ✅ After Fixing

1. Update `.env` with URL-encoded password
2. Restart server:
```bash
cd server
go run main.go database.go
```

You should see:
```
✅ Loaded .env file
✅ Connected to database successfully
✅ Database migration completed
gRPC server listening on :8080
```

---

## 🔐 Alternative: Change Supabase Password

If encoding is too complex, you can:
1. Go to Supabase Dashboard
2. Settings → Database
3. Reset database password
4. Use a password without special characters

---

## 📚 Reference

- URL Encoding: https://www.urlencoder.org/
- PostgreSQL Connection Strings: https://www.postgresql.org/docs/current/libpq-connect.html

