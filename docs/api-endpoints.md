# UniNotes API Dokümantasyonu

Bu doküman, UniNotes platformundaki tüm API endpoint'lerini ve kullanımlarını açıklar.

## İçindekiler

- [Genel Bilgiler](#genel-bilgiler)
- [Kimlik Doğrulama (Auth) API](#kimlik-doğrulama-auth-api)
- [Not (Note) API](#not-note-api)
- [PDF API](#pdf-api)
- [Beğeni (Like) API](#beğeni-like-api)
- [Yorum (Comment) API](#yorum-comment-api)
- [Davet Bağlantısı (Invite) API](#davet-bağlantısı-invite-api)
- [Görüntüleme Takip (View) API](#görüntüleme-takip-view-api)

## Genel Bilgiler

### Temel URL
```
https://api.uninotes.com
```

### API Versiyonu
Tüm endpoint'ler `/api/v1` öneki ile başlar.

### Kimlik Doğrulama
Çoğu endpoint JWT tabanlı kimlik doğrulama gerektirir. Token, `Authorization` header'ında `Bearer` şeması ile gönderilmelidir:

```
Authorization: Bearer <token>
```

### Hata Yanıtları
API, aşağıdaki HTTP durum kodlarını kullanarak hata durumlarını bildirir:

- `400 Bad Request`: Geçersiz istek formatı veya parametreler
- `401 Unauthorized`: Kimlik doğrulama başarısız
- `403 Forbidden`: Yetkilendirme başarısız (kimlik doğrulanmış ancak yetkisiz)
- `404 Not Found`: İstenen kaynak bulunamadı
- `500 Internal Server Error`: Sunucu hatası

### Sayfalama
Çoğu liste endpoint'i sayfalama destekler. Sayfalama için aşağıdaki sorgu parametreleri kullanılabilir:

- `limit`: Sayfa başına öğe sayısı (varsayılan: 10)
- `offset`: Atlanacak öğe sayısı (varsayılan: 0)

## Kimlik Doğrulama (Auth) API

### Kayıt Olma

**Endpoint:** `POST /api/v1/register`

**Kimlik Doğrulama:** Gerekli değil

**İstek Gövdesi:**
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepassword",
  "firstName": "John",
  "lastName": "Doe",
  "university": "Örnek Üniversitesi",
  "department": "Bilgisayar Mühendisliği",
  "class": "3. Sınıf"
}
```

**Başarılı Yanıt (201 Created):**
```json
{
  "message": "Kullanıcı başarıyla kaydedildi"
}
```

### Giriş Yapma

**Endpoint:** `POST /api/v1/login`

**Kimlik Doğrulama:** Gerekli değil

**İstek Gövdesi:**
```json
{
  "email": "john@example.com",
  "password": "securepassword"
}
```

**Başarılı Yanıt (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Profil Bilgilerini Getirme

**Endpoint:** `GET /api/v1/profile`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Başarılı Yanıt (200 OK):**
```json
{
  "id": 42,
  "username": "johndoe",
  "email": "john@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "university": "Örnek Üniversitesi",
  "department": "Bilgisayar Mühendisliği",
  "class": "3. Sınıf",
  "createdAt": "2025-03-20T10:15:30Z"
}
```

### Profil Bilgilerini Güncelleme

**Endpoint:** `PUT /api/v1/profile`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**İstek Gövdesi:**
```json
{
  "firstName": "John",
  "lastName": "Doe",
  "university": "Yeni Üniversite",
  "department": "Bilgisayar Mühendisliği",
  "class": "4. Sınıf"
}
```

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "Profil başarıyla güncellendi"
}
```

### Şifre Değiştirme

**Endpoint:** `POST /api/v1/change-password`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**İstek Gövdesi:**
```json
{
  "oldPassword": "securepassword",
  "newPassword": "newsecurepassword"
}
```

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "Şifre başarıyla değiştirildi"
}
```

## Not (Note) API

### Not Oluşturma

**Endpoint:** `POST /api/v1/notes`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**İstek Gövdesi:**
```json
{
  "title": "Veri Yapıları Notları",
  "content": "# Veri Yapıları\n\n## Diziler\n\nDiziler, aynı türdeki verileri...",
  "tags": ["bilgisayar", "algoritma"],
  "isPublic": true
}
```

**Başarılı Yanıt (201 Created):**
```json
{
  "id": 123,
  "title": "Veri Yapıları Notları",
  "content": "# Veri Yapıları\n\n## Diziler\n\nDiziler, aynı türdeki verileri...",
  "userId": 42,
  "tags": ["bilgisayar", "algoritma"],
  "isPublic": true,
  "createdAt": "2025-03-22T15:30:45Z",
  "updatedAt": "2025-03-22T15:30:45Z"
}
```

### Not Güncelleme

**Endpoint:** `PUT /api/v1/notes/{id}`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**İstek Gövdesi:**
```json
{
  "title": "Güncellenmiş Veri Yapıları Notları",
  "content": "# Veri Yapıları\n\n## Diziler ve Bağlı Listeler\n\nDiziler, aynı türdeki verileri...",
  "tags": ["bilgisayar", "algoritma", "veri yapıları"],
  "isPublic": true
}
```

**Başarılı Yanıt (200 OK):**
```json
{
  "id": 123,
  "title": "Güncellenmiş Veri Yapıları Notları",
  "content": "# Veri Yapıları\n\n## Diziler ve Bağlı Listeler\n\nDiziler, aynı türdeki verileri...",
  "userId": 42,
  "tags": ["bilgisayar", "algoritma", "veri yapıları"],
  "isPublic": true,
  "createdAt": "2025-03-22T15:30:45Z",
  "updatedAt": "2025-03-22T16:45:20Z"
}
```

### Not Silme

**Endpoint:** `DELETE /api/v1/notes/{id}`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "Not başarıyla silindi"
}
```

### Not Getirme

**Endpoint:** `GET /api/v1/notes/{id}`

**Kimlik Doğrulama:** Opsiyonel (Herkese açık notlar için gerekli değil)

**Başarılı Yanıt (200 OK):**
```json
{
  "id": 123,
  "title": "Veri Yapıları Notları",
  "content": "# Veri Yapıları\n\n## Diziler\n\nDiziler, aynı türdeki verileri...",
  "userId": 42,
  "tags": ["bilgisayar", "algoritma"],
  "isPublic": true,
  "createdAt": "2025-03-22T15:30:45Z",
  "updatedAt": "2025-03-22T15:30:45Z",
  "likeCount": 15
}
```

### Kullanıcının Notlarını Getirme

**Endpoint:** `GET /api/v1/notes/my`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Sorgu Parametreleri:**
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 123,
    "title": "Veri Yapıları Notları",
    "content": "# Veri Yapıları\n\n## Diziler\n\nDiziler, aynı türdeki verileri...",
    "userId": 42,
    "tags": ["bilgisayar", "algoritma"],
    "isPublic": true,
    "createdAt": "2025-03-22T15:30:45Z",
    "updatedAt": "2025-03-22T15:30:45Z",
    "likeCount": 15
  },
  // ... diğer notlar
]
```

### Herkese Açık Notları Getirme

**Endpoint:** `GET /api/v1/notes`

**Kimlik Doğrulama:** Gerekli değil

**Sorgu Parametreleri:**
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 123,
    "title": "Veri Yapıları Notları",
    "content": "# Veri Yapıları\n\n## Diziler\n\nDiziler, aynı türdeki verileri...",
    "userId": 42,
    "tags": ["bilgisayar", "algoritma"],
    "isPublic": true,
    "createdAt": "2025-03-22T15:30:45Z",
    "updatedAt": "2025-03-22T15:30:45Z",
    "likeCount": 15
  },
  // ... diğer notlar
]
```

### Not Arama

**Endpoint:** `GET /api/v1/notes/search`

**Kimlik Doğrulama:** Gerekli değil

**Sorgu Parametreleri:**
- `q` (gerekli): Arama sorgusu
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 123,
    "title": "Veri Yapıları Notları",
    "content": "# Veri Yapıları\n\n## Diziler\n\nDiziler, aynı türdeki verileri...",
    "userId": 42,
    "tags": ["bilgisayar", "algoritma"],
    "isPublic": true,
    "createdAt": "2025-03-22T15:30:45Z",
    "updatedAt": "2025-03-22T15:30:45Z",
    "likeCount": 15
  },
  // ... diğer notlar
]
```

### Etikete Göre Not Getirme

**Endpoint:** `GET /api/v1/notes/tag/{tag}`

**Kimlik Doğrulama:** Gerekli değil

**Sorgu Parametreleri:**
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 123,
    "title": "Veri Yapıları Notları",
    "content": "# Veri Yapıları\n\n## Diziler\n\nDiziler, aynı türdeki verileri...",
    "userId": 42,
    "tags": ["bilgisayar", "algoritma"],
    "isPublic": true,
    "createdAt": "2025-03-22T15:30:45Z",
    "updatedAt": "2025-03-22T15:30:45Z",
    "likeCount": 15
  },
  // ... diğer notlar
]
```

### Nota Yorum Ekleme

**Endpoint:** `POST /api/v1/notes/{id}/comments`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**İstek Gövdesi:**
```json
{
  "content": "Harika bir not, teşekkürler!"
}
```

**Başarılı Yanıt (201 Created):**
```json
{
  "id": 456,
  "noteId": 123,
  "userId": 42,
  "content": "Harika bir not, teşekkürler!",
  "createdAt": "2025-03-22T17:30:45Z"
}
```

### Not Yorumlarını Getirme

**Endpoint:** `GET /api/v1/notes/{id}/comments`

**Kimlik Doğrulama:** Opsiyonel (Özel notlar için gerekli)

**Sorgu Parametreleri:**
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 456,
    "contentId": 123,
    "userId": 42,
    "username": "johndoe",
    "fullName": "John Doe",
    "content": "Harika bir not, teşekkürler!",
    "createdAt": "2025-03-22T17:30:45Z",
    "updatedAt": "2025-03-22T17:30:45Z"
  },
  // ... diğer yorumlar
]
```

### Not Beğenme

**Endpoint:** `POST /api/v1/notes/{id}/like`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "Not başarıyla beğenildi"
}
```

### Not Beğenisini Kaldırma

**Endpoint:** `DELETE /api/v1/notes/{id}/like`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "Not beğenisi başarıyla kaldırıldı"
}
```

### Beğenilen Notları Getirme

**Endpoint:** `GET /api/v1/notes/liked`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Sorgu Parametreleri:**
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 123,
    "title": "Veri Yapıları Notları",
    "content": "# Veri Yapıları\n\n## Diziler\n\nDiziler, aynı türdeki verileri...",
    "userId": 56,
    "tags": ["bilgisayar", "algoritma"],
    "isPublic": true,
    "createdAt": "2025-03-22T15:30:45Z",
    "updatedAt": "2025-03-22T15:30:45Z",
    "likeCount": 15
  },
  // ... diğer notlar
]
```

## PDF API

### PDF Yükleme

**Endpoint:** `POST /api/v1/pdfs`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**İstek Formatı:** `multipart/form-data`

**Form Alanları:**
- `file`: PDF dosyası
- `title`: PDF başlığı
- `description`: PDF açıklaması
- `tags`: Etiketler (JSON dizisi olarak)
- `isPublic`: Herkese açık mı (boolean)

**Başarılı Yanıt (201 Created):**
```json
{
  "id": 456,
  "title": "Makine Öğrenmesi Ders Notları",
  "description": "2025 Bahar Dönemi Makine Öğrenmesi dersi notları",
  "userId": 42,
  "tags": ["yapay zeka", "veri bilimi"],
  "isPublic": true,
  "createdAt": "2025-03-22T18:30:45Z",
  "updatedAt": "2025-03-22T18:30:45Z"
}
```

### PDF Güncelleme

**Endpoint:** `PUT /api/v1/pdfs/{id}`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**İstek Gövdesi:**
```json
{
  "title": "Güncellenmiş Makine Öğrenmesi Ders Notları",
  "description": "2025 Bahar Dönemi Makine Öğrenmesi dersi güncellenmiş notları",
  "tags": ["yapay zeka", "veri bilimi", "derin öğrenme"],
  "isPublic": true
}
```

**Başarılı Yanıt (200 OK):**
```json
{
  "id": 456,
  "title": "Güncellenmiş Makine Öğrenmesi Ders Notları",
  "description": "2025 Bahar Dönemi Makine Öğrenmesi dersi güncellenmiş notları",
  "userId": 42,
  "tags": ["yapay zeka", "veri bilimi", "derin öğrenme"],
  "isPublic": true,
  "createdAt": "2025-03-22T18:30:45Z",
  "updatedAt": "2025-03-22T19:45:30Z"
}
```

### PDF Silme

**Endpoint:** `DELETE /api/v1/pdfs/{id}`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "PDF başarıyla silindi"
}
```

### PDF Getirme

**Endpoint:** `GET /api/v1/pdfs/{id}`

**Kimlik Doğrulama:** Opsiyonel (Herkese açık PDF'ler için gerekli değil)

**Başarılı Yanıt (200 OK):**
```json
{
  "id": 456,
  "title": "Makine Öğrenmesi Ders Notları",
  "description": "2025 Bahar Dönemi Makine Öğrenmesi dersi notları",
  "userId": 42,
  "tags": ["yapay zeka", "veri bilimi"],
  "isPublic": true,
  "createdAt": "2025-03-22T18:30:45Z",
  "updatedAt": "2025-03-22T18:30:45Z",
  "likeCount": 10
}
```

### PDF İçeriğini Getirme

**Endpoint:** `GET /api/v1/pdfs/{id}/content`

**Kimlik Doğrulama:** Opsiyonel (Herkese açık PDF'ler için gerekli değil)

**Başarılı Yanıt (200 OK):**
PDF dosyası içeriği (application/pdf)

### Kullanıcının PDF'lerini Getirme

**Endpoint:** `GET /api/v1/pdfs/my`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Sorgu Parametreleri:**
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 456,
    "title": "Makine Öğrenmesi Ders Notları",
    "description": "2025 Bahar Dönemi Makine Öğrenmesi dersi notları",
    "userId": 42,
    "tags": ["yapay zeka", "veri bilimi"],
    "isPublic": true,
    "createdAt": "2025-03-22T18:30:45Z",
    "updatedAt": "2025-03-22T18:30:45Z",
    "likeCount": 10
  },
  // ... diğer PDF'ler
]
```

### Herkese Açık PDF'leri Getirme

**Endpoint:** `GET /api/v1/pdfs`

**Kimlik Doğrulama:** Gerekli değil

**Sorgu Parametreleri:**
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 456,
    "title": "Makine Öğrenmesi Ders Notları",
    "description": "2025 Bahar Dönemi Makine Öğrenmesi dersi notları",
    "userId": 42,
    "tags": ["yapay zeka", "veri bilimi"],
    "isPublic": true,
    "createdAt": "2025-03-22T18:30:45Z",
    "updatedAt": "2025-03-22T18:30:45Z",
    "likeCount": 10
  },
  // ... diğer PDF'ler
]
```

### PDF Arama

**Endpoint:** `GET /api/v1/pdfs/search`

**Kimlik Doğrulama:** Gerekli değil

**Sorgu Parametreleri:**
- `q` (gerekli): Arama sorgusu
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 456,
    "title": "Makine Öğrenmesi Ders Notları",
    "description": "2025 Bahar Dönemi Makine Öğrenmesi dersi notları",
    "userId": 42,
    "tags": ["yapay zeka", "veri bilimi"],
    "isPublic": true,
    "createdAt": "2025-03-22T18:30:45Z",
    "updatedAt": "2025-03-22T18:30:45Z",
    "likeCount": 10
  },
  // ... diğer PDF'ler
]
```

### Etikete Göre PDF Getirme

**Endpoint:** `GET /api/v1/pdfs/tag/{tag}`

**Kimlik Doğrulama:** Gerekli değil

**Sorgu Parametreleri:**
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 456,
    "title": "Makine Öğrenmesi Ders Notları",
    "description": "2025 Bahar Dönemi Makine Öğrenmesi dersi notları",
    "userId": 42,
    "tags": ["yapay zeka", "veri bilimi"],
    "isPublic": true,
    "createdAt": "2025-03-22T18:30:45Z",
    "updatedAt": "2025-03-22T18:30:45Z",
    "likeCount": 10
  },
  // ... diğer PDF'ler
]
```

### PDF'e Yorum Ekleme

**Endpoint:** `POST /api/v1/pdfs/{id}/comments`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**İstek Gövdesi:**
```json
{
  "content": "Çok faydalı bir kaynak, teşekkürler!",
  "pageNumber": 5
}
```

**Başarılı Yanıt (201 Created):**
```json
{
  "id": 789,
  "pdfId": 456,
  "userId": 42,
  "content": "Çok faydalı bir kaynak, teşekkürler!",
  "pageNumber": 5,
  "createdAt": "2025-03-22T20:30:45Z"
}
```

### PDF Yorumlarını Getirme

**Endpoint:** `GET /api/v1/pdfs/{id}/comments`

**Kimlik Doğrulama:** Opsiyonel (Özel PDF'ler için gerekli)

**Sorgu Parametreleri:**
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 789,
    "contentId": 456,
    "userId": 42,
    "username": "johndoe",
    "fullName": "John Doe",
    "content": "Çok faydalı bir kaynak, teşekkürler!",
    "pageNumber": 5,
    "createdAt": "2025-03-22T20:30:45Z",
    "updatedAt": "2025-03-22T20:30:45Z"
  },
  // ... diğer yorumlar
]
```

### PDF'e İşaretleme Ekleme

**Endpoint:** `POST /api/v1/pdfs/{id}/annotations`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**İstek Gövdesi:**
```json
{
  "pageNumber": 5,
  "content": "Bu kısım önemli!",
  "x": 100.5,
  "y": 200.5,
  "width": 150.0,
  "height": 30.0,
  "type": "highlight",
  "color": "#FFFF00"
}
```

**Başarılı Yanıt (201 Created):**
```json
{
  "id": 101,
  "pdfId": 456,
  "userId": 42,
  "pageNumber": 5,
  "content": "Bu kısım önemli!",
  "x": 100.5,
  "y": 200.5,
  "width": 150.0,
  "height": 30.0,
  "type": "highlight",
  "color": "#FFFF00",
  "createdAt": "2025-03-22T21:15:30Z"
}
```

### PDF İşaretlemelerini Getirme

**Endpoint:** `GET /api/v1/pdfs/{id}/annotations`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 101,
    "pdfId": 456,
    "userId": 42,
    "pageNumber": 5,
    "content": "Bu kısım önemli!",
    "x": 100.5,
    "y": 200.5,
    "width": 150.0,
    "height": 30.0,
    "type": "highlight",
    "color": "#FFFF00",
    "createdAt": "2025-03-22T21:15:30Z"
  },
  // ... diğer işaretlemeler
]
```

### PDF Beğenme

**Endpoint:** `POST /api/v1/pdfs/{id}/like`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "PDF başarıyla beğenildi"
}
```

### PDF Beğenisini Kaldırma

**Endpoint:** `DELETE /api/v1/pdfs/{id}/like`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "PDF beğenisi başarıyla kaldırıldı"
}
```

### Beğenilen PDF'leri Getirme

**Endpoint:** `GET /api/v1/pdfs/liked`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Sorgu Parametreleri:**
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 456,
    "title": "Makine Öğrenmesi Ders Notları",
    "description": "2025 Bahar Dönemi Makine Öğrenmesi dersi notları",
    "userId": 56,
    "tags": ["yapay zeka", "veri bilimi"],
    "isPublic": true,
    "createdAt": "2025-03-22T18:30:45Z",
    "updatedAt": "2025-03-22T18:30:45Z",
    "likeCount": 10
  },
  // ... diğer PDF'ler
]
```

## Davet Bağlantısı (Invite) API

### Not için Davet Bağlantısı Oluşturma

**Endpoint:** `POST /api/v1/notes/{id}/invites`

**Açıklama:** Belirtilen not için bir davet bağlantısı oluşturur.

**Kimlik Doğrulama:** Gerekli (JWT Token)

**URL Parametreleri:**
- `id`: Not ID'si

**İstek Gövdesi:**
```json
{
  "expiresAt": "2025-04-24T00:00:00Z" // Opsiyonel, belirtilmezse 7 gün sonra sona erer
}
```

**Başarılı Yanıt (201 Created):**
```json
{
  "id": 1,
  "contentId": 123,
  "type": "note",
  "token": "abcdef123456",
  "expiresAt": "2025-04-24T00:00:00Z",
  "isActive": true,
  "createdAt": "2025-03-24T04:00:00Z"
}
```

### PDF için Davet Bağlantısı Oluşturma

**Endpoint:** `POST /api/v1/pdfs/{id}/invites`

**Açıklama:** Belirtilen PDF için bir davet bağlantısı oluşturur.

**Kimlik Doğrulama:** Gerekli (JWT Token)

**URL Parametreleri:**
- `id`: PDF ID'si

**İstek Gövdesi:**
```json
{
  "expiresAt": "2025-04-24T00:00:00Z" // Opsiyonel, belirtilmezse 7 gün sonra sona erer
}
```

**Başarılı Yanıt (201 Created):**
```json
{
  "id": 1,
  "contentId": 456,
  "type": "pdf",
  "token": "abcdef123456",
  "expiresAt": "2025-04-24T00:00:00Z",
  "isActive": true,
  "createdAt": "2025-03-24T04:00:00Z"
}
```

### Not için Davet Bağlantılarını Getirme

**Endpoint:** `GET /api/v1/notes/{id}/invites`

**Açıklama:** Belirtilen not için oluşturulmuş tüm davet bağlantılarını getirir.

**Kimlik Doğrulama:** Gerekli (JWT Token)

**URL Parametreleri:**
- `id`: Not ID'si

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 1,
    "contentId": 123,
    "type": "note",
    "token": "abcdef123456",
    "expiresAt": "2025-04-24T00:00:00Z",
    "isActive": true,
    "createdAt": "2025-03-24T04:00:00Z"
  },
  {
    "id": 2,
    "contentId": 123,
    "type": "note",
    "token": "ghijkl789012",
    "expiresAt": "2025-04-30T00:00:00Z",
    "isActive": true,
    "createdAt": "2025-03-25T04:00:00Z"
  }
]
```

### PDF için Davet Bağlantılarını Getirme

**Endpoint:** `GET /api/v1/pdfs/{id}/invites`

**Açıklama:** Belirtilen PDF için oluşturulmuş tüm davet bağlantılarını getirir.

**Kimlik Doğrulama:** Gerekli (JWT Token)

**URL Parametreleri:**
- `id`: PDF ID'si

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 3,
    "contentId": 456,
    "type": "pdf",
    "token": "mnopqr345678",
    "expiresAt": "2025-04-24T00:00:00Z",
    "isActive": true,
    "createdAt": "2025-03-24T04:00:00Z"
  },
  {
    "id": 4,
    "contentId": 456,
    "type": "pdf",
    "token": "stuvwx901234",
    "expiresAt": "2025-04-30T00:00:00Z",
    "isActive": true,
    "createdAt": "2025-03-25T04:00:00Z"
  }
]
```

### Davet Bağlantısını Devre Dışı Bırakma

**Endpoint:** `DELETE /api/v1/invites/{id}`

**Açıklama:** Belirtilen davet bağlantısını devre dışı bırakır.

**Kimlik Doğrulama:** Gerekli (JWT Token)

**URL Parametreleri:**
- `id`: Davet bağlantısı ID'si

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "Davet bağlantısı başarıyla devre dışı bırakıldı"
}
```

### Davet Bağlantısını Doğrulama

**Endpoint:** `GET /api/v1/invites/{token}`

**Açıklama:** Belirtilen davet bağlantısının geçerli olup olmadığını kontrol eder.

**Kimlik Doğrulama:** Gerekli değil

**URL Parametreleri:**
- `token`: Davet bağlantısı token'ı

**Başarılı Yanıt (200 OK):**
```json
{
  "valid": true,
  "contentId": 123,
  "type": "note",
  "expiresAt": "2025-04-24T00:00:00Z"
}
```

### Davet Bağlantısı ile Not Getirme

**Endpoint:** `GET /api/v1/notes/invite/{token}`

**Açıklama:** Belirtilen davet bağlantısı ile bir notu getirir. Bu endpoint, davet bağlantısı ile özel notlara erişim sağlar. Davet bağlantısı geçerli ise, notun `isPublic` değeri `false` olsa bile nota erişim sağlanabilir.

**Kimlik Doğrulama:** Gerekli değil

**URL Parametreleri:**
- `token`: Davet bağlantısı token'ı

**Başarılı Yanıt (200 OK):**
```json
{
  "id": 123,
  "title": "Not Başlığı",
  "content": "Not içeriği...",
  "userId": 1,
  "tags": ["etiket1", "etiket2"],
  "isPublic": false,
  "viewCount": 10,
  "likeCount": 5,
  "commentCount": 3,
  "createdAt": "2025-03-20T04:00:00Z",
  "updatedAt": "2025-03-22T04:00:00Z"
}
```

### Davet Bağlantısı ile PDF Getirme

**Endpoint:** `GET /api/v1/pdfs/invite/{token}`

**Açıklama:** Belirtilen davet bağlantısı ile bir PDF'i getirir. Bu endpoint, davet bağlantısı ile özel PDF'lere erişim sağlar. Davet bağlantısı geçerli ise, PDF'in `isPublic` değeri `false` olsa bile PDF'e erişim sağlanabilir.

**Kimlik Doğrulama:** Gerekli değil

**URL Parametreleri:**
- `token`: Davet bağlantısı token'ı

**Başarılı Yanıt (200 OK):**
```json
{
  "id": 456,
  "title": "PDF Başlığı",
  "description": "PDF açıklaması...",
  "filePath": "storage/pdfs/1_document.pdf",
  "fileSize": 1024,
  "userId": 1,
  "tags": ["etiket1", "etiket2"],
  "isPublic": false,
  "viewCount": 10,
  "likeCount": 5,
  "commentCount": 3,
  "createdAt": "2025-03-20T04:00:00Z",
  "updatedAt": "2025-03-22T04:00:00Z"
}
```

## Görüntüleme Takip (View) API

### Not Görüntüleme

**Endpoint:** `GET /api/v1/notes/{id}/view`

**Açıklama:** Bir notu görüntüler ve kullanıcı giriş yapmışsa görüntüleme kaydı oluşturur.

**Kimlik Doğrulama:** Opsiyonel (Kimlik doğrulama yapılmışsa görüntüleme kaydı oluşturulur)

**URL Parametreleri:**
- `id` (zorunlu): Görüntülenecek notun ID'si

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "Not görüntülendi"
}
```

### PDF Görüntüleme

**Endpoint:** `GET /api/v1/pdfs/{id}/view`

**Açıklama:** Bir PDF'i görüntüler ve kullanıcı giriş yapmışsa görüntüleme kaydı oluşturur.

**Kimlik Doğrulama:** Opsiyonel (Kimlik doğrulama yapılmışsa görüntüleme kaydı oluşturulur)

**URL Parametreleri:**
- `id` (zorunlu): Görüntülenecek PDF'in ID'si

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "PDF görüntülendi"
}
```

### İçerik Görüntüleme Kayıtlarını Getirme

**Endpoint:** `GET /api/v1/views/content/{type}/{id}`

**Açıklama:** Bir içeriğin (not veya PDF) görüntüleme kayıtlarını döndürür. Sadece içerik sahibi bu endpoint'e erişebilir.

**Kimlik Doğrulama:** Gerekli (JWT Token)

**URL Parametreleri:**
- `type` (zorunlu): İçerik türü (`note` veya `pdf`)
- `id` (zorunlu): İçerik ID'si

**Sorgu Parametreleri:**
- `limit` (opsiyonel): Sayfa başına kayıt sayısı (varsayılan: 10)
- `offset` (opsiyonel): Atlanacak kayıt sayısı (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
{
  "views": [
    {
      "id": 1,
      "userId": 2,
      "username": "johndoe",
      "firstName": "John",
      "lastName": "Doe",
      "contentId": 5,
      "type": "note",
      "viewedAt": "2025-03-24T15:30:45Z"
    },
    {
      "id": 2,
      "userId": 3,
      "username": "janedoe",
      "firstName": "Jane",
      "lastName": "Doe",
      "contentId": 5,
      "type": "note",
      "viewedAt": "2025-03-24T16:20:10Z"
    }
  ],
  "pagination": {
    "limit": 10,
    "offset": 0
  }
}
```

### Kullanıcı Görüntüleme Kayıtlarını Getirme

**Endpoint:** `GET /api/v1/views/user`

**Açıklama:** Kullanıcının görüntüleme kayıtlarını döndürür.

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Sorgu Parametreleri:**
- `limit` (opsiyonel): Sayfa başına kayıt sayısı (varsayılan: 10)
- `offset` (opsiyonel): Atlanacak kayıt sayısı (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
{
  "views": [
    {
      "id": 1,
      "userId": 1,
      "contentId": 5,
      "type": "note",
      "viewedAt": "2025-03-24T15:30:45Z"
    },
    {
      "id": 2,
      "userId": 1,
      "contentId": 8,
      "type": "pdf",
      "viewedAt": "2025-03-24T16:20:10Z"
    }
  ],
  "pagination": {
    "limit": 10,
    "offset": 0
  }
}
```

### Görüntüleme Durumu Kontrolü

**Endpoint:** `GET /api/v1/views/check`

**Açıklama:** Kullanıcının bir içeriği görüntüleyip görüntülemediğini kontrol eder.

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Sorgu Parametreleri:**
- `type` (zorunlu): İçerik türü (`note` veya `pdf`)
- `contentId` (zorunlu): İçerik ID'si

**Başarılı Yanıt (200 OK):**
```json
{
  "viewed": true
}
```
