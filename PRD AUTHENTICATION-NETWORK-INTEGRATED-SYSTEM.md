Sebagai Senior Software Engineer, merombak sistem untuk skala **10 juta user** bukan lagi soal "bisa login", tapi soal **Availability, Security, dan Latency**. Di skala ini, database bottleneck dan session management adalah musuh utama.

Jika saya memimpin proyek ini, saya tidak akan menaruh logika auth di dalam monolith backend utama. Saya akan memisahkannya menjadi **Centralized Identity Provider (IdP)**.

Berikut adalah rancangan arsitektur dan stack yang saya pilih.

---

### 1. High-Level Architecture: Centralized Identity Service

Kita akan beralih ke standar industri: **OAuth 2.0 & OpenID Connect (OIDC)**.

* **Pemisahan Concern:** Auth Service hanya mengurus *who you are* (Authentication). Backend lain (Order, Product, dll) mengurus *what you can do* (Authorization) berdasarkan token yang valid.
* **Stateless Authentication:** Menggunakan **JWT (JSON Web Token)** untuk Access Token agar service lain tidak perlu bolak-balik tanya ke database untuk memvalidasi user setiap request.
* **Stateful Session:** Menggunakan **Refresh Token** (Opaque token) yang disimpan di database/cache untuk kemampuan *revoke* akses (misal: "Log out from all devices").

### 2. Technology Stack (The "Performance" Stack)

Untuk menangani 10 juta user dengan *concurrent login* yang tinggi, saya memilih stack yang *low-overhead* dan *high-throughput*.

| Komponen | Pilihan Teknologi | Alasan Senior Engineer |
| --- | --- | --- |
| **Language** | **Go (Golang)** | Concurrency model (Goroutines) sangat efisien untuk menangani ribuan request auth per detik dengan memori kecil. Jauh lebih cepat dibanding Python/Node.js untuk *computational task* seperti hashing. |
| **Core Protocol** | **Ory Hydra / Custom OIDC** | Jangan tulis *crypto* sendiri. Gunakan engine yang sudah diaudit seperti **Ory Hydra** (Go-based) atau library `osin` jika ingin custom wrapper. Atau gunakan **Keycloak** jika ingin fitur "siap pakai" (tapi perlu tuning JVM yang berat). |
| **Database** | **PostgreSQL** | ACID compliant. Fitur JSONB bagus untuk menyimpan user attributes yang dinamis. Wajib menggunakan **Partitioning** untuk tabel user. |
| **Caching** | **Redis Cluster** | Wajib. Digunakan untuk menyimpan session ephemeral, rate limiting, dan *token blacklist*. |
| **Message Broker** | **Kafka / RabbitMQ** | Untuk *User Events* (misal: `UserRegistered`, `UserLoginFailed`). Service lain bisa *subscribe* untuk kirim email welcome atau deteksi fraud, tanpa memperlambat proses login. |
| **Infrastructure** | **Kubernetes (K8s)** | Auto-scaling pod Auth Service saat traffic spike (misal: saat promo flash sale). |

---

### 3. Deep Dive: Strategi Database & Security

Menangani 10 juta user di database butuh strategi khusus:

#### A. Database Optimization (PostgreSQL)

* **Connection Pooling:** Gunakan **PgBouncer**. Jangan biarkan aplikasi Go membuka ribuan koneksi langsung ke Postgres.
* **Table Partitioning:** Jangan simpan 10 juta user di satu tabel fisik. Partisi tabel `users` dan `audit_logs` (misalnya berdasarkan `created_at` atau hash dari `user_id`).
* **Read Replicas:** Pisahkan traffic. Login (Write last login) ke Master, Validasi User (Read profile) ke Replica.

#### B. Security Measures

* **Password Hashing:** Gunakan **Argon2id**. Jangan MD5, SHA1, atau bahkan Bcrypt biasa (Argon2 lebih tahan terhadap serangan GPU/ASIC).
* **MFA (Multi-Factor Auth):** Wajib support TOTP (Google Authenticator) atau WebAuthn (TouchID/FaceID).
* **Rate Limiting:** Implementasikan "Leaky Bucket" algorithm di Redis untuk mencegah Brute Force attack pada endpoint login.

---

### 4. Workflow Autentikasi (Flow Chart Logic)

1. **User Login:**
* Client kirim `email` + `password`.
* Auth Service validasi hash (Argon2).
* Jika sukses  Generate **Access Token** (JWT, exp: 15 min) & **Refresh Token** (Random String, exp: 7 days).
* Refresh Token disimpan di Redis & DB (untuk persistensi).
* Return ke user.


2. **Access Resource (Misal: Get Profile):**
* Client kirim Request + `Header: Bearer <JWT>`.
* API Gateway / Service memvalidasi *signature* JWT (CPU bound, tanpa DB call).
* Jika valid  Process request.


3. **Token Refresh:**
* Saat JWT expired, Client kirim Refresh Token.
* Auth Service cek di Redis/DB: Apakah Refresh Token valid? Apakah user diban?
* Jika oke  Rotate Refresh Token (Ganti baru) & Issue JWT baru. (Rotation mencegah pencurian token jangka panjang).



---

### 5. Strategi Migrasi (The Hardest Part)

Anda tidak bisa mematikan sistem lama begitu saja. Gunakan pola **"Lazy Migration" (Double Hashing)**:

1. Import semua user dari DB lama ke DB baru.
2. Karena kita tidak tahu password asli user (hanya hash lama), di DB baru kita simpan hash lama di kolom sementara (misal: `legacy_password_hash`).
3. **Saat User Login:**
* Sistem cek: Apakah user punya password baru (Argon2)?
* **Jika TIDAK:** Cek input password dengan algoritma lama (misal: MD5).
* Jika cocok  Hash password input dengan Argon2, simpan ke kolom password baru, hapus hash lama. (User ter-migrasi otomatis).


* **Jika YA:** Validasi langsung dengan Argon2.



---

### Kesimpulan untuk Eksekusi

Untuk membangun ini, saya akan mulai dengan:

1. **Repo Setup:** Go (Golang) + Gin/Fiber Framework.
2. **Infrastructure:** Setup Docker Compose dengan Postgres & Redis.
3. **Core Logic:** Implementasi OIDC flow dan Argon2 hashing.

