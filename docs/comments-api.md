# Yorum API Dokümantasyonu

Bu dokümantasyon, UniNotes uygulamasının yorum API'sini açıklamaktadır. Yorum API'si, kullanıcıların notlara ve PDF'lere yorum yapabilmelerini sağlar.

## Genel Bakış

Yorum API'si, aşağıdaki özellikleri sağlar:

- Notlara yorum ekleme
- PDF'lere yorum ekleme
- Not yorumlarını getirme
- PDF yorumlarını getirme
- Kullanıcı bilgileriyle zenginleştirilmiş yorum yanıtları

## Endpoint'ler

### Not Yorumları

#### Yorum Ekleme

```
POST /api/v1/notes/{id}/comments
```

**Açıklama:** Bir nota yorum ekler.

**Yetkilendirme:** Gerekli (JWT token)

**URL Parametreleri:**
- `id`: Not ID'si

**İstek Gövdesi:**
```json
{
  "content": "Bu not çok faydalı oldu, teşekkürler!"
}
```

**Başarılı Yanıt (201 Created):**
```json
{
  "id": 1,
  "noteId": 123,
  "userId": 456,
  "content": "Bu not çok faydalı oldu, teşekkürler!",
  "createdAt": "2025-03-23T10:30:00Z",
  "updatedAt": "2025-03-23T10:30:00Z"
}
```

**Hata Yanıtları:**
- `400 Bad Request`: Geçersiz istek formatı
- `401 Unauthorized`: Kimlik doğrulama hatası
- `404 Not Found`: Not bulunamadı
- `500 Internal Server Error`: Sunucu hatası

#### Yorumları Getirme

```
GET /api/v1/notes/{id}/comments
```

**Açıklama:** Bir notun yorumlarını getirir.

**Yetkilendirme:** Opsiyonel (Özel notlar için gerekli)

**URL Parametreleri:**
- `id`: Not ID'si

**Sorgu Parametreleri:**
- `limit`: Sayfa başına yorum sayısı (varsayılan: 10)
- `offset`: Atlanacak yorum sayısı (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 1,
    "contentId": 123,
    "userId": 456,
    "username": "ahmet_yilmaz",
    "fullName": "Ahmet Yılmaz",
    "content": "Bu not çok faydalı oldu, teşekkürler!",
    "createdAt": "2025-03-23T10:30:00Z",
    "updatedAt": "2025-03-23T10:30:00Z"
  },
  {
    "id": 2,
    "contentId": 123,
    "userId": 789,
    "username": "ayse_demir",
    "fullName": "Ayşe Demir",
    "content": "Benim de çok işime yaradı.",
    "createdAt": "2025-03-23T11:15:00Z",
    "updatedAt": "2025-03-23T11:15:00Z"
  }
]
```

**Hata Yanıtları:**
- `403 Forbidden`: Erişim izni yok (özel not için)
- `404 Not Found`: Not bulunamadı
- `500 Internal Server Error`: Sunucu hatası

### PDF Yorumları

#### Yorum Ekleme

```
POST /api/v1/pdfs/{id}/comments
```

**Açıklama:** Bir PDF'e yorum ekler.

**Yetkilendirme:** Gerekli (JWT token)

**URL Parametreleri:**
- `id`: PDF ID'si

**İstek Gövdesi:**
```json
{
  "content": "Bu PDF çok faydalı oldu, teşekkürler!",
  "pageNumber": 5
}
```

**Başarılı Yanıt (201 Created):**
```json
{
  "id": 1,
  "pdfId": 123,
  "userId": 456,
  "content": "Bu PDF çok faydalı oldu, teşekkürler!",
  "pageNumber": 5,
  "createdAt": "2025-03-23T10:30:00Z",
  "updatedAt": "2025-03-23T10:30:00Z"
}
```

**Hata Yanıtları:**
- `400 Bad Request`: Geçersiz istek formatı
- `401 Unauthorized`: Kimlik doğrulama hatası
- `404 Not Found`: PDF bulunamadı
- `500 Internal Server Error`: Sunucu hatası

#### Yorumları Getirme

```
GET /api/v1/pdfs/{id}/comments
```

**Açıklama:** Bir PDF'in yorumlarını getirir.

**Yetkilendirme:** Opsiyonel (Özel PDF'ler için gerekli)

**URL Parametreleri:**
- `id`: PDF ID'si

**Sorgu Parametreleri:**
- `limit`: Sayfa başına yorum sayısı (varsayılan: 10)
- `offset`: Atlanacak yorum sayısı (varsayılan: 0)

**Başarılı Yanıt (200 OK):**
```json
[
  {
    "id": 1,
    "contentId": 123,
    "userId": 456,
    "username": "ahmet_yilmaz",
    "fullName": "Ahmet Yılmaz",
    "content": "Bu PDF çok faydalı oldu, teşekkürler!",
    "pageNumber": 5,
    "createdAt": "2025-03-23T10:30:00Z",
    "updatedAt": "2025-03-23T10:30:00Z"
  },
  {
    "id": 2,
    "contentId": 123,
    "userId": 789,
    "username": "ayse_demir",
    "fullName": "Ayşe Demir",
    "content": "Benim de çok işime yaradı.",
    "pageNumber": 7,
    "createdAt": "2025-03-23T11:15:00Z",
    "updatedAt": "2025-03-23T11:15:00Z"
  }
]
```

**Hata Yanıtları:**
- `403 Forbidden`: Erişim izni yok (özel PDF için)
- `404 Not Found`: PDF bulunamadı
- `500 Internal Server Error`: Sunucu hatası

## Veri Modelleri

### Not Yorumu

```json
{
  "id": 1,
  "noteId": 123,
  "userId": 456,
  "content": "Yorum içeriği",
  "createdAt": "2025-03-23T10:30:00Z",
  "updatedAt": "2025-03-23T10:30:00Z"
}
```

### PDF Yorumu

```json
{
  "id": 1,
  "pdfId": 123,
  "userId": 456,
  "content": "Yorum içeriği",
  "pageNumber": 5,
  "createdAt": "2025-03-23T10:30:00Z",
  "updatedAt": "2025-03-23T10:30:00Z"
}
```

### Zenginleştirilmiş Yorum Yanıtı

```json
{
  "id": 1,
  "contentId": 123,
  "userId": 456,
  "username": "kullanici_adi",
  "fullName": "Ad Soyad",
  "content": "Yorum içeriği",
  "pageNumber": 5, // Sadece PDF yorumları için
  "createdAt": "2025-03-23T10:30:00Z",
  "updatedAt": "2025-03-23T10:30:00Z"
}
```

## Erişim Kontrolü

- Yorum eklemek için kimlik doğrulama gereklidir.
- Yorumları görüntülemek için:
  - Herkese açık içerikler için kimlik doğrulama gerekmez.
  - Özel içerikler için, kullanıcı içeriğin sahibi olmalı veya içeriğe erişim izni olmalıdır.

## Kullanım Örnekleri

### Not Yorumu Ekleme

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{"content": "Bu not çok faydalı oldu, teşekkürler!"}' \
  https://api.uninotes.com/api/v1/notes/123/comments
```

### PDF Yorumu Ekleme

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{"content": "Bu PDF çok faydalı oldu, teşekkürler!", "pageNumber": 5}' \
  https://api.uninotes.com/api/v1/pdfs/123/comments
```

### Not Yorumlarını Getirme

```bash
curl -X GET \
  https://api.uninotes.com/api/v1/notes/123/comments?limit=10&offset=0
```

### PDF Yorumlarını Getirme

```bash
curl -X GET \
  https://api.uninotes.com/api/v1/pdfs/123/comments?limit=10&offset=0
