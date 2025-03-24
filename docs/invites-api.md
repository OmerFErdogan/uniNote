# Davet Bağlantıları API

Bu dokümantasyon, UniNotes platformunda notlar ve PDF'ler için davet bağlantıları oluşturma, yönetme ve kullanma ile ilgili API endpoint'lerini açıklar.

## Genel Bakış

Davet bağlantıları, kullanıcıların özel notlarını veya PDF'lerini başkalarıyla paylaşmalarına olanak tanır. Bir davet bağlantısı oluşturulduğunda, bu bağlantıya sahip herkes, içeriğin sahibi olmasa bile içeriğe erişebilir.

## Endpoint'ler

### 1. Not için Davet Bağlantısı Oluşturma

```
POST /api/v1/notes/{id}/invites
```

**Açıklama:** Belirtilen not için bir davet bağlantısı oluşturur.

**Yetkilendirme:** Gerekli (JWT Token)

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

**Hata Yanıtları:**
- `400 Bad Request`: Geçersiz istek formatı veya geçersiz içerik türü
- `401 Unauthorized`: Yetkilendirme hatası
- `403 Forbidden`: Bu işlem için yetkiniz yok
- `404 Not Found`: Not bulunamadı
- `500 Internal Server Error`: Sunucu hatası

### 2. PDF için Davet Bağlantısı Oluşturma

```
POST /api/v1/pdfs/{id}/invites
```

**Açıklama:** Belirtilen PDF için bir davet bağlantısı oluşturur.

**Yetkilendirme:** Gerekli (JWT Token)

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

**Hata Yanıtları:**
- `400 Bad Request`: Geçersiz istek formatı veya geçersiz içerik türü
- `401 Unauthorized`: Yetkilendirme hatası
- `403 Forbidden`: Bu işlem için yetkiniz yok
- `404 Not Found`: PDF bulunamadı
- `500 Internal Server Error`: Sunucu hatası

### 3. Not için Davet Bağlantılarını Getirme

```
GET /api/v1/notes/{id}/invites
```

**Açıklama:** Belirtilen not için oluşturulmuş tüm davet bağlantılarını getirir.

**Yetkilendirme:** Gerekli (JWT Token)

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

**Hata Yanıtları:**
- `401 Unauthorized`: Yetkilendirme hatası
- `403 Forbidden`: Bu işlem için yetkiniz yok
- `404 Not Found`: Not bulunamadı
- `500 Internal Server Error`: Sunucu hatası

### 4. PDF için Davet Bağlantılarını Getirme

```
GET /api/v1/pdfs/{id}/invites
```

**Açıklama:** Belirtilen PDF için oluşturulmuş tüm davet bağlantılarını getirir.

**Yetkilendirme:** Gerekli (JWT Token)

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

**Hata Yanıtları:**
- `401 Unauthorized`: Yetkilendirme hatası
- `403 Forbidden`: Bu işlem için yetkiniz yok
- `404 Not Found`: PDF bulunamadı
- `500 Internal Server Error`: Sunucu hatası

### 5. Davet Bağlantısını Devre Dışı Bırakma

```
DELETE /api/v1/invites/{id}
```

**Açıklama:** Belirtilen davet bağlantısını devre dışı bırakır.

**Yetkilendirme:** Gerekli (JWT Token)

**URL Parametreleri:**
- `id`: Davet bağlantısı ID'si

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "Davet bağlantısı başarıyla devre dışı bırakıldı"
}
```

**Hata Yanıtları:**
- `401 Unauthorized`: Yetkilendirme hatası
- `403 Forbidden`: Bu işlem için yetkiniz yok
- `404 Not Found`: Davet bağlantısı bulunamadı
- `500 Internal Server Error`: Sunucu hatası

### 6. Davet Bağlantısını Doğrulama

```
GET /api/v1/invites/{token}
```

**Açıklama:** Belirtilen davet bağlantısının geçerli olup olmadığını kontrol eder.

**Yetkilendirme:** Gerekli değil

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

**Hata Yanıtları:**
- `400 Bad Request`: Geçersiz token
- `403 Forbidden`: Davet bağlantısı aktif değil veya süresi dolmuş
- `404 Not Found`: Davet bağlantısı bulunamadı veya içerik bulunamadı
- `500 Internal Server Error`: Sunucu hatası

### 7. Davet Bağlantısı ile Not Getirme

```
GET /api/v1/notes/invite/{token}
```

**Açıklama:** Belirtilen davet bağlantısı ile bir notu getirir.

**Yetkilendirme:** Gerekli değil

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

**Hata Yanıtları:**
- `400 Bad Request`: Geçersiz token veya bu davet bağlantısı bir not için değil
- `403 Forbidden`: Davet bağlantısı aktif değil veya süresi dolmuş
- `404 Not Found`: Davet bağlantısı bulunamadı veya not bulunamadı
- `500 Internal Server Error`: Sunucu hatası

### 8. Davet Bağlantısı ile PDF Getirme

```
GET /api/v1/pdfs/invite/{token}
```

**Açıklama:** Belirtilen davet bağlantısı ile bir PDF'i getirir.

**Yetkilendirme:** Gerekli değil

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

**Hata Yanıtları:**
- `400 Bad Request`: Geçersiz token veya bu davet bağlantısı bir PDF için değil
- `403 Forbidden`: Davet bağlantısı aktif değil veya süresi dolmuş
- `404 Not Found`: Davet bağlantısı bulunamadı veya PDF bulunamadı
- `500 Internal Server Error`: Sunucu hatası

## Kullanım Örnekleri

### Örnek 1: Not için Davet Bağlantısı Oluşturma

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"expiresAt": "2025-04-24T00:00:00Z"}' \
  http://localhost:8080/api/v1/notes/123/invites
```

### Örnek 2: Davet Bağlantısı ile Not Getirme

```bash
curl -X GET \
  http://localhost:8080/api/v1/notes/invite/abcdef123456
```

## Notlar

- Davet bağlantıları varsayılan olarak oluşturulduktan 7 gün sonra sona erer, ancak bu süre isteğe bağlı olarak değiştirilebilir.
- Davet bağlantıları, içerik sahibi tarafından devre dışı bırakılabilir.
- Davet bağlantıları, içerik sahibi olmayan kullanıcıların özel içeriklere erişmesine olanak tanır.
