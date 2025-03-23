# Beğeni API Dokümantasyonu

Bu doküman, UniNotes platformundaki beğeni sistemi API endpoint'lerini ve kullanımlarını açıklar.

## Genel Bakış

Beğeni sistemi, kullanıcıların notları ve PDF'leri beğenmesine olanak tanır. Sistem, tek bir model üzerinden hem not hem de PDF beğenilerini yönetir ve veri tutarlılığını sağlamak için transaction kullanır.

## Endpoint'ler

### 1. İçerik Beğenme

**Endpoint:** `POST /likes`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**İstek Gövdesi:**
```json
{
  "contentId": 123,
  "type": "note" // "note" veya "pdf"
}
```

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "İçerik başarıyla beğenildi"
}
```

**Hata Yanıtları:**
- `400 Bad Request`: Geçersiz istek formatı veya geçersiz içerik türü
- `404 Not Found`: İçerik bulunamadı
- `401 Unauthorized`: Kimlik doğrulama başarısız
- `500 Internal Server Error`: Sunucu hatası

**Örnek Kullanım:**
```bash
curl -X POST http://api.uninotes.com/likes \
  -H "Authorization: Bearer {TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"contentId": 123, "type": "note"}'
```

**Notlar:**
- Bir kullanıcı aynı içeriği birden fazla kez beğenmeye çalışırsa, ikinci beğeni isteği işlenmez ve başarılı yanıt döndürülür.
- Beğeni işlemi ve içerik beğeni sayacı güncellemesi transaction içinde gerçekleştirilir.

---

### 2. İçerik Beğenisini Kaldırma

**Endpoint:** `DELETE /likes`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**İstek Gövdesi:**
```json
{
  "contentId": 123,
  "type": "note" // "note" veya "pdf"
}
```

**Başarılı Yanıt (200 OK):**
```json
{
  "message": "İçerik beğenisi başarıyla kaldırıldı"
}
```

**Hata Yanıtları:**
- `400 Bad Request`: Geçersiz istek formatı veya geçersiz içerik türü
- `404 Not Found`: İçerik bulunamadı
- `401 Unauthorized`: Kimlik doğrulama başarısız
- `500 Internal Server Error`: Sunucu hatası

**Örnek Kullanım:**
```bash
curl -X DELETE http://api.uninotes.com/likes \
  -H "Authorization: Bearer {TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"contentId": 123, "type": "note"}'
```

**Notlar:**
- Kullanıcı beğenmediği bir içeriğin beğenisini kaldırmaya çalışırsa, işlem başarılı olarak kabul edilir.
- Beğeni silme işlemi ve içerik beğeni sayacı güncellemesi transaction içinde gerçekleştirilir.

---

### 3. Kullanıcının Beğenilerini Getirme

**Endpoint:** `GET /likes/my`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Sorgu Parametreleri:**
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 1,
    "userId": 42,
    "contentId": 123,
    "type": "note",
    "createdAt": "2025-03-22T15:30:45Z"
  },
  {
    "id": 2,
    "userId": 42,
    "contentId": 456,
    "type": "pdf",
    "createdAt": "2025-03-22T16:20:10Z"
  }
]
```

**Hata Yanıtları:**
- `401 Unauthorized`: Kimlik doğrulama başarısız
- `500 Internal Server Error`: Sunucu hatası

**Örnek Kullanım:**
```bash
curl -X GET "http://api.uninotes.com/likes/my?limit=20&offset=0" \
  -H "Authorization: Bearer {TOKEN}"
```

---

### 4. İçeriğin Beğenilerini Getirme

**Endpoint:** `GET /likes`

**Kimlik Doğrulama:** Gerekli değil

**Sorgu Parametreleri:**
- `contentId` (gerekli): İçerik ID'si
- `type` (gerekli): İçerik türü ("note" veya "pdf")
- `limit` (isteğe bağlı): Sayfalama için limit (varsayılan: 10)
- `offset` (isteğe bağlı): Sayfalama için offset (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 1,
    "userId": 42,
    "contentId": 123,
    "type": "note",
    "createdAt": "2025-03-22T15:30:45Z"
  },
  {
    "id": 3,
    "userId": 56,
    "contentId": 123,
    "type": "note",
    "createdAt": "2025-03-22T17:05:22Z"
  }
]
```

**Hata Yanıtları:**
- `400 Bad Request`: Geçersiz sorgu parametreleri veya geçersiz içerik türü
- `500 Internal Server Error`: Sunucu hatası

**Örnek Kullanım:**
```bash
curl -X GET "http://api.uninotes.com/likes?contentId=123&type=note&limit=20&offset=0"
```

---

### 5. Beğeni Durumu Kontrolü

**Endpoint:** `GET /likes/check`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**Sorgu Parametreleri:**
- `contentId` (gerekli): İçerik ID'si
- `type` (gerekli): İçerik türü ("note" veya "pdf")

**Başarılı Yanıt (200 OK):**
```json
{
  "isLiked": true
}
```

**Hata Yanıtları:**
- `400 Bad Request`: Geçersiz sorgu parametreleri veya geçersiz içerik türü
- `401 Unauthorized`: Kimlik doğrulama başarısız
- `500 Internal Server Error`: Sunucu hatası

**Örnek Kullanım:**
```bash
curl -X GET "http://api.uninotes.com/likes/check?contentId=123&type=note" \
  -H "Authorization: Bearer {TOKEN}"
```

**Notlar:**
- Bu endpoint, istemci tarafında önbelleğe alma için "Cache-Control: private, max-age=300" header'ı ile yanıt verir.
- Yanıt 5 dakika boyunca istemci tarafında önbelleğe alınabilir.

---

### 6. Toplu Beğeni Durumu Kontrolü

**Endpoint:** `POST /likes/check-bulk`

**Kimlik Doğrulama:** Gerekli (JWT Token)

**İstek Gövdesi:**
```json
{
  "items": [
    {
      "contentId": 123,
      "type": "note"
    },
    {
      "contentId": 456,
      "type": "pdf"
    },
    {
      "contentId": 789,
      "type": "note"
    }
  ]
}
```

**Başarılı Yanıt (200 OK):**
```json
{
  "results": {
    "123_note": true,
    "456_pdf": false,
    "789_note": true
  }
}
```

**Hata Yanıtları:**
- `400 Bad Request`: Geçersiz istek formatı veya boş items dizisi
- `401 Unauthorized`: Kimlik doğrulama başarısız
- `500 Internal Server Error`: Sunucu hatası

**Örnek Kullanım:**
```bash
curl -X POST http://api.uninotes.com/likes/check-bulk \
  -H "Authorization: Bearer {TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"items":[{"contentId":123,"type":"note"},{"contentId":456,"type":"pdf"}]}'
```

**Notlar:**
- Bu endpoint, istemci tarafında önbelleğe alma için "Cache-Control: private, max-age=300" header'ı ile yanıt verir.
- Yanıt 5 dakika boyunca istemci tarafında önbelleğe alınabilir.
- Geçersiz içerik türleri veya hata durumunda ilgili içerikler sonuç listesinden atlanır.
- Sonuç anahtarları "{contentId}_{type}" formatındadır.

---

### 7. Kullanıcının Beğendiği Notları Getirme

**Endpoint:** `GET /notes/liked`

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
    "content": "...",
    "userId": 56,
    "createdAt": "2025-03-20T10:15:30Z",
    "updatedAt": "2025-03-21T14:25:10Z",
    "likeCount": 42,
    "tags": ["bilgisayar", "algoritma"]
  },
  // ... diğer notlar
]
```

**Hata Yanıtları:**
- `401 Unauthorized`: Kimlik doğrulama başarısız
- `500 Internal Server Error`: Sunucu hatası

**Örnek Kullanım:**
```bash
curl -X GET "http://api.uninotes.com/notes/liked?limit=20&offset=0" \
  -H "Authorization: Bearer {TOKEN}"
```

---

### 8. Kullanıcının Beğendiği PDF'leri Getirme

**Endpoint:** `GET /pdfs/liked`

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
    "filePath": "/storage/pdfs/456_machine_learning.pdf",
    "userId": 78,
    "createdAt": "2025-03-19T09:45:20Z",
    "likeCount": 35,
    "tags": ["yapay zeka", "veri bilimi"]
  },
  // ... diğer PDF'ler
]
```

**Hata Yanıtları:**
- `401 Unauthorized`: Kimlik doğrulama başarısız
- `500 Internal Server Error`: Sunucu hatası

**Örnek Kullanım:**
```bash
curl -X GET "http://api.uninotes.com/pdfs/liked?limit=20&offset=0" \
  -H "Authorization: Bearer {TOKEN}"
```

## Performans İyileştirmeleri

### Transaction Kullanımı

Beğeni ekleme ve silme işlemleri sırasında, beğeni kaydı ve ilgili içeriğin beğeni sayacı güncellemesi tek bir transaction içinde gerçekleştirilir. Bu, veri tutarlılığını sağlar ve bir işlem sırasında hata oluşursa tüm değişikliklerin geri alınmasını garanti eder.

### İstemci Tarafı Önbelleğe Alma

Beğeni durumu kontrol endpoint'leri (`/likes/check` ve `/likes/check-bulk`), "Cache-Control: private, max-age=300" header'ı ile yanıt verir. Bu, istemci tarafında 5 dakika boyunca önbelleğe almayı sağlayarak gereksiz API çağrılarını azaltır.

### Toplu Beğeni Kontrolü

`/likes/check-bulk` endpoint'i, birden fazla içeriğin beğeni durumunu tek bir API çağrısıyla kontrol etmeyi sağlar. Bu, istemcinin her içerik için ayrı bir istek göndermek yerine tek bir istek göndermesine olanak tanıyarak ağ trafiğini ve sunucu yükünü azaltır.
