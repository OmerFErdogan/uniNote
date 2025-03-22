#!/bin/bash

# Renk tanımlamaları
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test sonuçları
PASSED=0
FAILED=0
TOTAL=0

# Test fonksiyonu
test_endpoint() {
    local description=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected_status=$5
    local token=$6
    
    echo -e "\n${YELLOW}Test: $description${NC}"
    echo "Endpoint: $method $endpoint"
    
    # HTTP isteği gönder
    if [ "$method" = "GET" ]; then
        if [ -z "$token" ]; then
            response=$(curl -s -o response.txt -w "%{http_code}" -X GET "http://localhost:8888$endpoint")
        else
            response=$(curl -s -o response.txt -w "%{http_code}" -X GET "http://localhost:8888$endpoint" -H "Authorization: Bearer $token")
        fi
    elif [ "$method" = "POST" ]; then
        if [ -z "$token" ]; then
            response=$(curl -s -o response.txt -w "%{http_code}" -X POST "http://localhost:8888$endpoint" -H "Content-Type: application/json" -d "$data")
        else
            response=$(curl -s -o response.txt -w "%{http_code}" -X POST "http://localhost:8888$endpoint" -H "Content-Type: application/json" -H "Authorization: Bearer $token" -d "$data")
        fi
    elif [ "$method" = "PUT" ]; then
        if [ -z "$token" ]; then
            response=$(curl -s -o response.txt -w "%{http_code}" -X PUT "http://localhost:8888$endpoint" -H "Content-Type: application/json" -d "$data")
        else
            response=$(curl -s -o response.txt -w "%{http_code}" -X PUT "http://localhost:8888$endpoint" -H "Content-Type: application/json" -H "Authorization: Bearer $token" -d "$data")
        fi
    elif [ "$method" = "DELETE" ]; then
        if [ -z "$token" ]; then
            response=$(curl -s -o response.txt -w "%{http_code}" -X DELETE "http://localhost:8888$endpoint")
        else
            response=$(curl -s -o response.txt -w "%{http_code}" -X DELETE "http://localhost:8888$endpoint" -H "Authorization: Bearer $token")
        fi
    fi
    
    # Yanıtı kontrol et
    if [ "$response" = "$expected_status" ]; then
        echo -e "${GREEN}✓ Başarılı (HTTP $response)${NC}"
        cat response.txt | jq . 2>/dev/null || cat response.txt
        PASSED=$((PASSED+1))
    else
        echo -e "${RED}✗ Başarısız (Beklenen: HTTP $expected_status, Alınan: HTTP $response)${NC}"
        cat response.txt | jq . 2>/dev/null || cat response.txt
        FAILED=$((FAILED+1))
    fi
    
    TOTAL=$((TOTAL+1))
    
    # Yanıt içeriğini döndür
    cat response.txt
}

# Sunucunun çalışıp çalışmadığını kontrol et
echo -e "${YELLOW}Sunucu durumu kontrol ediliyor...${NC}"
if curl -s -o /dev/null -w "%{http_code}" http://localhost:8888/ | grep -q "200"; then
    echo -e "${GREEN}Sunucu çalışıyor!${NC}"
else
    echo -e "${RED}Sunucu çalışmıyor! Lütfen sunucuyu başlatın.${NC}"
    exit 1
fi

# Ana endpoint testi
test_endpoint "Ana endpoint" "GET" "/" "" 200

# Sağlık kontrolü testi
test_endpoint "Sağlık kontrolü" "GET" "/api/v1/health" "" 200

# Kullanıcı işlemleri testleri
echo -e "\n${YELLOW}Kullanıcı işlemleri testleri${NC}"

# Kullanıcı kaydı
register_response=$(test_endpoint "Kullanıcı kaydı" "POST" "/api/v1/register" '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "firstName": "Test",
    "lastName": "User",
    "university": "Test University",
    "department": "Computer Science",
    "class": "Senior"
}' 201)

# Kullanıcı girişi
login_response=$(test_endpoint "Kullanıcı girişi" "POST" "/api/v1/login" '{
    "email": "test@example.com",
    "password": "password123"
}' 200)

# Token'ı çıkar
token=$(echo "$login_response" | jq -r '.token' 2>/dev/null)

if [ -z "$token" ] || [ "$token" = "null" ]; then
    echo -e "${RED}Token alınamadı. Kimlik doğrulama gerektiren testler atlanacak.${NC}"
else
    echo -e "${GREEN}Token alındı: $token${NC}"
    
    # Profil getirme
    test_endpoint "Profil getirme" "GET" "/api/v1/profile" "" 200 "$token"
    
    # Not işlemleri testleri
    echo -e "\n${YELLOW}Not işlemleri testleri${NC}"
    
    # Not oluşturma
    create_note_response=$(test_endpoint "Not oluşturma" "POST" "/api/v1/notes" '{
        "title": "Test Not",
        "content": "Bu bir test notudur.",
        "tags": ["test", "not"],
        "isPublic": true
    }' 201 "$token")
    
    # Not ID'sini çıkar
    note_id=$(echo "$create_note_response" | jq -r '.id' 2>/dev/null)
    
    if [ -n "$note_id" ] && [ "$note_id" != "null" ]; then
        # Not getirme
        test_endpoint "Not getirme" "GET" "/api/v1/notes/$note_id" "" 200
        
        # Not güncelleme
        test_endpoint "Not güncelleme" "PUT" "/api/v1/notes/$note_id" '{
            "title": "Güncellenmiş Test Not",
            "content": "Bu güncellenmiş bir test notudur.",
            "tags": ["test", "not", "güncelleme"],
            "isPublic": true
        }' 200 "$token"
        
        # Nota yorum ekleme
        test_endpoint "Nota yorum ekleme" "POST" "/api/v1/notes/$note_id/comments" '{
            "content": "Bu bir test yorumudur."
        }' 201 "$token"
        
        # Not yorumlarını getirme
        test_endpoint "Not yorumlarını getirme" "GET" "/api/v1/notes/$note_id/comments" "" 200
        
        # Notu beğenme
        test_endpoint "Notu beğenme" "POST" "/api/v1/notes/$note_id/like" "" 200 "$token"
        
        # Not beğenisini kaldırma
        test_endpoint "Not beğenisini kaldırma" "DELETE" "/api/v1/notes/$note_id/like" "" 200 "$token"
        
        # Notu silme
        test_endpoint "Notu silme" "DELETE" "/api/v1/notes/$note_id" "" 200 "$token"
    else
        echo -e "${RED}Not ID'si alınamadı. Not işlemleri testleri tamamlanamadı.${NC}"
    fi
    
    # PDF işlemleri testleri
    echo -e "\n${YELLOW}PDF işlemleri testleri${NC}"
    echo -e "${YELLOW}Not: PDF yükleme testi için curl'ün multipart/form-data desteği gereklidir.${NC}"
    echo -e "${YELLOW}Bu test manuel olarak yapılmalıdır.${NC}"
    
    # Herkese açık PDF'leri getirme
    test_endpoint "Herkese açık PDF'leri getirme" "GET" "/api/v1/pdfs" "" 200
    
    # Kullanıcının PDF'lerini getirme
    test_endpoint "Kullanıcının PDF'lerini getirme" "GET" "/api/v1/pdfs/my" "" 200 "$token"
    
    # PDF arama
    test_endpoint "PDF arama" "GET" "/api/v1/pdfs/search?q=test" "" 200
fi

# Sonuçları göster
echo -e "\n${YELLOW}Test Sonuçları:${NC}"
echo -e "${GREEN}Başarılı: $PASSED${NC}"
echo -e "${RED}Başarısız: $FAILED${NC}"
echo -e "Toplam: $TOTAL"

# Geçici dosyaları temizle
rm -f response.txt

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}Tüm testler başarıyla tamamlandı!${NC}"
    exit 0
else
    echo -e "\n${RED}Bazı testler başarısız oldu.${NC}"
    exit 1
fi
