# Görüntüleme Takip API'si

Bu API, not ve PDF'lere kimlerin baktığını takip etmek ve bu bilgileri not/PDF sahibinin görebilmesini sağlar.

## Genel Bakış

Görüntüleme takip sistemi, kullanıcıların içerikleri (not veya PDF) görüntülemelerini kaydeder ve içerik sahiplerinin bu görüntüleme kayıtlarını görmesine olanak tanır. Sistem, aşağıdaki özellikleri sağlar:

- İçerik görüntüleme kaydı oluşturma
- İçerik görüntüleme kayıtlarını listeleme (sadece içerik sahibi erişebilir)
- Kullanıcının görüntüleme kayıtlarını listeleme
- Kullanıcının bir içeriği görüntüleyip görüntülemediğini kontrol etme

## Endpoint'ler

### İçerik Görüntüleme

#### Not Görüntüleme

```
GET /api/v1/notes/{id}/view
```

**Açıklama:** Bir notu görüntüler ve kullanıcı giriş yapmışsa görüntüleme kaydı oluşturur.

**Yetkilendirme:** Opsiyonel (Kimlik doğrulama yapılmışsa görüntüleme kaydı oluşturulur)

**URL Parametreleri:**
- `id` (zorunlu): Görüntülenecek notun ID'si

**Yanıt:**
```json
{
  "message": "Not görüntülendi"
}
```

#### PDF Görüntüleme

```
GET /api/v1/pdfs/{id}/view
```

**Açıklama:** Bir PDF'i görüntüler ve kullanıcı giriş yapmışsa görüntüleme kaydı oluşturur.

**Yetkilendirme:** Opsiyonel (Kimlik doğrulama yapılmışsa görüntüleme kaydı oluşturulur)

**URL Parametreleri:**
- `id` (zorunlu): Görüntülenecek PDF'in ID'si

**Yanıt:**
```json
{
  "message": "PDF görüntülendi"
}
```

### İçerik Görüntüleme Kayıtları

```
GET /api/v1/views/content/{type}/{id}
```

**Açıklama:** Bir içeriğin (not veya PDF) görüntüleme kayıtlarını döndürür. Sadece içerik sahibi bu endpoint'e erişebilir.

**Yetkilendirme:** Zorunlu (Kimlik doğrulama gerekli)

**URL Parametreleri:**
- `type` (zorunlu): İçerik türü (`note` veya `pdf`)
- `id` (zorunlu): İçerik ID'si

**Sorgu Parametreleri:**
- `limit` (opsiyonel): Sayfa başına kayıt sayısı (varsayılan: 10)
- `offset` (opsiyonel): Atlanacak kayıt sayısı (varsayılan: 0)

**Yanıt:**
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

### Kullanıcı Görüntüleme Kayıtları

```
GET /api/v1/views/user
```

**Açıklama:** Kullanıcının görüntüleme kayıtlarını döndürür.

**Yetkilendirme:** Zorunlu (Kimlik doğrulama gerekli)

**Sorgu Parametreleri:**
- `limit` (opsiyonel): Sayfa başına kayıt sayısı (varsayılan: 10)
- `offset` (opsiyonel): Atlanacak kayıt sayısı (varsayılan: 0)

**Yanıt:**
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

```
GET /api/v1/views/check
```

**Açıklama:** Kullanıcının bir içeriği görüntüleyip görüntülemediğini kontrol eder.

**Yetkilendirme:** Zorunlu (Kimlik doğrulama gerekli)

**Sorgu Parametreleri:**
- `type` (zorunlu): İçerik türü (`note` veya `pdf`)
- `contentId` (zorunlu): İçerik ID'si

**Yanıt:**
```json
{
  "viewed": true
}
```

## Görüntüleme Takip Kuralları

1. Kullanıcılar kendi içeriklerini görüntülediklerinde görüntüleme kaydı oluşturulmaz.
2. Özel (public olmayan) içerikler için görüntüleme kaydı oluşturulmaz.
3. Görüntüleme kayıtları, içerik sahibi tarafından görüntülenebilir.
4. Kullanıcılar kendi görüntüleme kayıtlarını görebilirler.
5. Görüntüleme sayısı, içerik her görüntülendiğinde artırılır (aynı kullanıcı tarafından tekrar görüntülense bile).
6. Görüntüleme kayıtları, kullanıcı bilgileriyle zenginleştirilmiş olarak döndürülür (kullanıcı adı, ad, soyad).

## Hata Kodları

- `400 Bad Request`: Geçersiz istek parametreleri
- `401 Unauthorized`: Kimlik doğrulama gerekli
- `403 Forbidden`: Bu içeriğin görüntüleme kayıtlarına erişim izniniz yok
- `404 Not Found`: İçerik bulunamadı
- `500 Internal Server Error`: Sunucu hatası
