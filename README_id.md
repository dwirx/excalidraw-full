# Excalidraw Full: Platform Kolaborasi Self-Hosted yang Siap Cloud

Excalidraw Full telah berevolusi. Ini bukan lagi sekadar pembungkus sederhana dari Excalidraw, melainkan platform kolaborasi yang kuat dan self-hosted, mengadopsi filosofi "Bawa Cloud Sendiri" (BYOC - Bring Your Own Cloud). Platform ini menyediakan autentikasi pengguna, manajemen multi-kanvas, serta kemampuan unik untuk terhubung langsung dari frontend ke cloud storage Anda sendiri.

Ide utamanya adalah membiarkan backend menangani identitas pengguna, sementara Anda (pengguna) memiliki kontrol penuh atas lokasi penyimpanan data.

## Perbedaan Utama dengan Excalidraw Resmi

- **Kolaborasi dan Berbagi Self-Hosted Lengkap**: Berbeda dengan Excalidraw resmi, semua fitur kolaborasi real-time dan berbagi ditangani oleh layanan backend yang Anda deploy sendiri, memastikan privasi dan kontrol data.
- **Manajemen Multi-Kanvas yang Kuat**: Anda dapat dengan mudah membuat, menyimpan, dan mengelola beberapa kanvas. Data dapat disimpan di backend server (seperti SQLite, S3), atau frontend dapat terhubung langsung ke cloud storage Anda sendiri (seperti Cloudflare KV), mencapai "kedaulatan data" yang sebenarnya.
- **Fitur AI Siap Pakai**: Tanpa konfigurasi klien yang rumit, Anda dapat langsung menggunakan fitur OpenAI terintegrasi (seperti GPT-4 Vision) setelah login, dengan API key yang dikelola dengan aman oleh backend, frontend hanya bertanggung jawab untuk memanggil.


![Manajemen Multi-Kanvas](./img/PixPin_2025-07-06_16-07-27.png)

![Pilihan Storage Multi](./img/PixPin_2025-07-06_16-08-29.png)

![Login Oauth2](./img/PixPin_2025-07-06_16-09-24.png)

![Fitur AI](./img/PixPin_2025-07-06_16-09-55.png)


## Fitur Utama

- **Autentikasi GitHub**: Login aman menggunakan GitHub OAuth.
- **Manajemen Multi-Kanvas**: Pengguna dapat membuat, menyimpan, dan mengelola beberapa kanvas gambar.
- **Penyimpanan Data Fleksibel (BYOC)**:
    - **Penyimpanan Backend Default**: Dukungan siap pakai untuk menyimpan kanvas di storage server (SQLite, file system, S3).
    - **Koneksi Cloud Langsung**: Frontend dapat terhubung langsung ke layanan cloud Anda sendiri, seperti **Cloudflare KV** atau **Amazon S3**, untuk mencapai kedaulatan data tertinggi. Kredensial Anda tidak akan pernah menyentuh server kami.
- **Kolaborasi Real-time**: Dukungan penuh untuk fitur kolaborasi real-time Excalidraw klasik.
- **Proxy OpenAI yang Aman**: Proxy backend opsional untuk menggunakan fitur GPT-4 Vision OpenAI, memastikan API key Anda aman.
- **File Binary Terintegrasi**: Seluruh aplikasi, termasuk frontend yang sudah di-patch dan server backend, dikompilasi menjadi satu file binary Go, memudahkan deployment.

## Strategi Penyimpanan Kanvas Frontend

- **IndexedDB**: Storage key-value yang cepat, aman, dan dapat diskalakan. Tidak memerlukan konfigurasi apapun, juga tidak perlu login.
- **Penyimpanan Backend**: Backend dapat menyimpan kanvas ke storage server (SQLite, file system, S3). Dapat disinkronkan antar perangkat.
- **Cloudflare KV**: Storage key-value yang cepat, aman, dan dapat diskalakan. Ini memerlukan Anda untuk deploy Worker pendamping di akun Cloudflare Anda sendiri. Lihat [**Panduan Deploy Cloudflare Worker**](./cloudflare-worker/README.md) untuk instruksi detail.
- **Amazon S3**: Layanan object storage yang andal, dapat diskalakan, dan ekonomis.

## Instalasi dan Menjalankan

Jalankan Docker satu klik [Excalidraw-Full](https://github.com/dwirx/excalidraw-full).

```bash
# Contoh Linux
git clone https://github.com/dwirx/excalidraw-full.git
cd excalidraw-full
mv .env.example .env
touch ./excalidraw.db  # Penting: inisialisasi database SQLite, jika tidak tidak bisa start
docker compose up -d
```

Server akan start, dan Anda dapat mengakses aplikasi di `http://localhost:3002`.

## Konfigurasi

Konfigurasi dikelola melalui variabel environment. Untuk template lengkap, lihat bagian `.env.example` di bawah.

### 1. Konfigurasi Backend (Wajib)

Anda harus mengkonfigurasi GitHub OAuth dan JWT secret agar aplikasi dapat berjalan dengan normal.

- `GITHUB_CLIENT_ID`: Client ID dari GitHub OAuth App Anda.
- `GITHUB_CLIENT_SECRET`: Client Secret dari GitHub OAuth App Anda.
- `GITHUB_REDIRECT_URL`: URL callback. Untuk testing lokal, ini adalah `http://localhost:3002/auth/github/callback`.
- `JWT_SECRET`: String random yang kuat untuk menandatangani session token. Gunakan `openssl rand -base64 32` untuk generate satu.
- `OPENAI_API_KEY`: Secret key OpenAI Anda.
- `OPENAI_BASE_URL`: (Opsional) Untuk menggunakan API yang kompatibel, seperti Azure OpenAI.

### 2. Penyimpanan Default (Opsional, tapi Direkomendasikan)

Ini mengkonfigurasi storage built-in server, digunakan secara default.

- `STORAGE_TYPE`: `memory` (default), `sqlite`, `filesystem`, atau `s3`.
- `DATA_SOURCE_NAME`: Path database SQLite (contoh, `excalidraw.db`).
- `LOCAL_STORAGE_PATH`: Direktori untuk file system storage.
- `S3_BUCKET_NAME`, `AWS_REGION`, dll: Untuk S3 storage.

### 3. Proxy OpenAI (Opsional)

Untuk mengaktifkan fitur AI, set OpenAI API key Anda.

- `OPENAI_API_KEY`: Secret key OpenAI Anda.
- `OPENAI_BASE_URL`: (Opsional) Untuk menggunakan API yang kompatibel, seperti Azure OpenAI.

### 4. Konfigurasi Frontend

Storage adapter frontend (seperti Cloudflare KV, S3) dikonfigurasi langsung di UI pengaturan aplikasi setelah Anda login. Ini dirancang secara sengaja: kredensial cloud pribadi Anda hanya disimpan di session browser, tidak akan pernah dikirim ke server backend.

### Contoh `.env.example`

Buat file `.env` di root direktori project, dan tambahkan konten berikut, isi dengan nilai Anda sendiri.

```env
# Konfigurasi server backend
# Dapatkan dari https://github.com/settings/developers
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_REDIRECT_URL=http://localhost:3002/auth/github/callback

# Generate dengan perintah: openssl rand -base64 32
JWT_SECRET=your_super_secret_jwt_string

# Penyimpanan default (SQLite)
STORAGE_TYPE=sqlite
DATA_SOURCE_NAME=excalidraw.db

# Proxy OpenAI opsional
OPENAI_API_KEY=sk-your_openai_api_key
```

## Build dari Source Code

Prosesnya mirip dengan sebelumnya, tapi sekarang perlu build Go backend.

### Menggunakan Docker (Direkomendasikan)

```bash
# Clone repository dan submodule-nya
git clone https://github.com/dwirx/excalidraw-full.git --recursive
cd excalidraw-full

# Build Docker image
# Proses ini akan menangani build frontend, patching, dan kompilasi Go backend.
docker build -t excalidraw-complete -f excalidraw-complete.Dockerfile .

# Jalankan container, dan berikan variabel environment
docker run -p 3002:3002 \
  -e GITHUB_CLIENT_ID="your_id" \
  -e GITHUB_CLIENT_SECRET="your_secret" \
  -e GITHUB_REDIRECT_URL="http://localhost:3002/auth/github/callback" \
  -e JWT_SECRET="your_jwt_secret" \
  -e STORAGE_TYPE="sqlite" \
  -e DATA_SOURCE_NAME="excalidraw.db" \
  -e OPENAI_API_KEY="your_openai_api_key" \
  excalidraw-complete
```

### Build Manual

1.  **Build Frontend**: Ikuti langkah-langkah di README asli, patch dan build Excalidraw frontend di dalam submodule `excalidraw/`.
2.  **Copy Frontend**: Pastikan file frontend yang sudah di-build dari `excalidraw/excalidraw-app/build` di-copy ke direktori `frontend/` di root.
3.  **Build Go Backend**:
    ```bash
    go build -o excalidraw-complete main.go
    ```
4.  **Jalankan**:
    ```bash
    # Pertama set variabel environment
    ./excalidraw-complete
    ```
---

Excalidraw adalah tool yang luar biasa. Project ini bertujuan untuk membuat semua orang dapat menggunakan versi yang powerful dan aman secara data. Kontribusi sangat disambut! 