# ThaID Authentication - Next.js + Golang

โปรเจกต์ตัวอย่างการเชื่อมต่อ **ThaID** (ระบบยืนยันตัวตนของรัฐบาลไทย) โดยใช้ **Next.js** (Frontend) และ **Golang** (Backend) แทน Flask

## 🏗️ โครงสร้างโปรเจกต์

```
thaid-nextjs-golang/
├── backend/                 # Golang Backend API
│   ├── config/
│   │   └── config.go       # การตั้งค่า
│   ├── .env.example        # ตัวอย่าง environment variables
│   ├── go.mod              # Go modules
│   └── main.go             # Entry point
├── frontend/               # Next.js Frontend
│   ├── app/
│   │   ├── layout.tsx      # Root layout
│   │   ├── page.tsx        # Home page
│   │   └── dashboard/
│   │       └── page.tsx    # Dashboard page
│   ├── .env.example        # ตัวอย่าง environment variables
│   ├── next.config.js      # Next.js config
│   ├── package.json
│   └── tsconfig.json
└── README.md
```

## 🔄 เปรียบเทียบ Flask vs Next.js + Go

| ฟีเจอร์ | Flask (Python) | Next.js + Go |
|---------|----------------|--------------|
| **Router** | `@app.route()` | Next.js File-based routing |
| **Session** | Flask Session | `gin-contrib/sessions` (Cookie) |
| **Template** | Jinja2 | React Components |
| **OAuth2** | Authlib | Custom implementation |
| **CORS** | flask-cors | gin-contrib/cors |

## 🚀 การติดตั้งและรัน

### 1. Backend (Golang)

```bash
cd backend

# ติดตั้ง dependencies
go mod tidy

# สร้างไฟล์ .env
cp .env.example .env
# แก้ไข .env ใส่ THAID_CLIENT_ID และ THAID_CLIENT_SECRET

# รัน server
go run main.go
```

Backend จะรันที่ `http://localhost:8080`

### 2. Frontend (Next.js)

```bash
cd frontend

# ติดตั้ง dependencies
npm install

# สร้างไฟล์ .env
cp .env.example .env.local

# รัน development server
npm run dev
```

Frontend จะรันที่ `http://localhost:3000`

## 📋 API Endpoints

| Endpoint | Method | คำอธิบาย |
|----------|--------|---------|
| `/api/auth/login` | GET | ขอ Authorization URL |
| `/api/auth/callback` | GET | รับ Callback จาก ThaID |
| `/api/auth/logout` | GET | ออกจากระบบ |
| `/api/auth/me` | GET | ดึงข้อมูลผู้ใช้ |
| `/api/auth/introspect` | POST | ตรวจสอบ Token |

## 🔑 การตั้งค่า ThaID

1. ไปที่ [ThaID Developer Portal](https://developers.thaid.
2. สร้าง Application ใหม่
3. ตั้งค่า **Redirect URI**: `http://localhost:8080/api/auth/callback`
4. คัดลอก **Client ID** และ **Client Secret** ไปใส่ใน `.env`

## 🔒 Security Notes

- ใน Production ควรเปลี่ยน `Secure: false` เป็น `true` ใน `main.go`
- ใช้ HTTPS สำหรับ Production
- เก็บ SESSION_SECRET ให้ปลอดภัย
- ใช้ Redis หรือ Database แทน Cookie Store สำหรับ Production
