# PowerShell script to test API endpoints

# Function to test an endpoint
function Test-Endpoint {
    param (
        [string]$Description,
        [string]$Method,
        [string]$Endpoint,
        [string]$Data,
        [int]$ExpectedStatus,
        [string]$Token
    )
    
    Write-Host "`nTest: $Description" -ForegroundColor Yellow
    Write-Host "Endpoint: $Method $Endpoint"
    
    $headers = @{
        "Content-Type" = "application/json"
    }
    
    if ($Token) {
        $headers["Authorization"] = "Bearer $Token"
    }
    
    $params = @{
        Method = $Method
        Uri = "http://localhost:8888$Endpoint"
        Headers = $headers
    }
    
    if ($Data -and ($Method -eq "POST" -or $Method -eq "PUT")) {
        $params["Body"] = $Data
    }
    
    try {
        $response = Invoke-RestMethod @params -ErrorAction SilentlyContinue
        Write-Host "✓ Başarılı (HTTP 200)" -ForegroundColor Green
        $response | ConvertTo-Json -Depth 10
        return $response
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq $ExpectedStatus) {
            Write-Host "✓ Başarılı (HTTP $statusCode)" -ForegroundColor Green
        }
        else {
            Write-Host "✗ Başarısız (Beklenen: HTTP $ExpectedStatus, Alınan: HTTP $statusCode)" -ForegroundColor Red
        }
        
        try {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $reader.BaseStream.Position = 0
            $reader.DiscardBufferedData()
            $responseBody = $reader.ReadToEnd()
            $responseBody
        }
        catch {
            Write-Host "Yanıt içeriği alınamadı: $_" -ForegroundColor Red
        }
    }
}

# Check if server is running
Write-Host "Sunucu durumu kontrol ediliyor..." -ForegroundColor Yellow
try {
    $null = Invoke-RestMethod -Uri "http://localhost:8888/" -Method GET -ErrorAction Stop
    Write-Host "Sunucu çalışıyor!" -ForegroundColor Green
}
catch {
    Write-Host "Sunucu çalışmıyor! Lütfen sunucuyu başlatın." -ForegroundColor Red
    exit 1
}

# Test main endpoint
$mainResponse = Test-Endpoint -Description "Ana endpoint" -Method "GET" -Endpoint "/" -ExpectedStatus 200

# Test health endpoint
$healthResponse = Test-Endpoint -Description "Sağlık kontrolü" -Method "GET" -Endpoint "/api/v1/health" -ExpectedStatus 200

# User operations tests
Write-Host "`nKullanıcı işlemleri testleri" -ForegroundColor Yellow

# Register user
$registerData = @{
    username = "testuser"
    email = "test@example.com"
    password = "password123"
    firstName = "Test"
    lastName = "User"
    university = "Test University"
    department = "Computer Science"
    class = "Senior"
} | ConvertTo-Json

$registerResponse = Test-Endpoint -Description "Kullanıcı kaydı" -Method "POST" -Endpoint "/api/v1/register" -Data $registerData -ExpectedStatus 201

# Login user
$loginData = @{
    email = "test@example.com"
    password = "password123"
} | ConvertTo-Json

$loginResponse = Test-Endpoint -Description "Kullanıcı girişi" -Method "POST" -Endpoint "/api/v1/login" -Data $loginData -ExpectedStatus 200

# Extract token
$token = $loginResponse.token

if (-not $token) {
    Write-Host "Token alınamadı. Kimlik doğrulama gerektiren testler atlanacak." -ForegroundColor Red
}
else {
    Write-Host "Token alındı: $token" -ForegroundColor Green
    
    # Get profile
    $profileResponse = Test-Endpoint -Description "Profil getirme" -Method "GET" -Endpoint "/api/v1/profile" -ExpectedStatus 200 -Token $token
    
    # Note operations tests
    Write-Host "`nNot işlemleri testleri" -ForegroundColor Yellow
    
    # Create note
    $noteData = @{
        title = "Test Not"
        content = "Bu bir test notudur."
        tags = @("test", "not")
        isPublic = $true
    } | ConvertTo-Json
    
    $createNoteResponse = Test-Endpoint -Description "Not oluşturma" -Method "POST" -Endpoint "/api/v1/notes" -Data $noteData -ExpectedStatus 201 -Token $token
    
    # Extract note ID
    $noteId = $createNoteResponse.id
    
    if ($noteId) {
        # Get note
        $getNoteResponse = Test-Endpoint -Description "Not getirme" -Method "GET" -Endpoint "/api/v1/notes/$noteId" -ExpectedStatus 200
        
        # Update note
        $updateNoteData = @{
            title = "Güncellenmiş Test Not"
            content = "Bu güncellenmiş bir test notudur."
            tags = @("test", "not", "güncelleme")
            isPublic = $true
        } | ConvertTo-Json
        
        $updateNoteResponse = Test-Endpoint -Description "Not güncelleme" -Method "PUT" -Endpoint "/api/v1/notes/$noteId" -Data $updateNoteData -ExpectedStatus 200 -Token $token
        
        # Add comment to note
        $commentData = @{
            content = "Bu bir test yorumudur."
        } | ConvertTo-Json
        
        $addCommentResponse = Test-Endpoint -Description "Nota yorum ekleme" -Method "POST" -Endpoint "/api/v1/notes/$noteId/comments" -Data $commentData -ExpectedStatus 201 -Token $token
        
        # Get note comments
        $getCommentsResponse = Test-Endpoint -Description "Not yorumlarını getirme" -Method "GET" -Endpoint "/api/v1/notes/$noteId/comments" -ExpectedStatus 200
        
        # Like note
        $likeNoteResponse = Test-Endpoint -Description "Notu beğenme" -Method "POST" -Endpoint "/api/v1/notes/$noteId/like" -ExpectedStatus 200 -Token $token
        
        # Unlike note
        $unlikeNoteResponse = Test-Endpoint -Description "Not beğenisini kaldırma" -Method "DELETE" -Endpoint "/api/v1/notes/$noteId/like" -ExpectedStatus 200 -Token $token
        
        # Delete note
        $deleteNoteResponse = Test-Endpoint -Description "Notu silme" -Method "DELETE" -Endpoint "/api/v1/notes/$noteId" -ExpectedStatus 200 -Token $token
    }
    else {
        Write-Host "Not ID'si alınamadı. Not işlemleri testleri tamamlanamadı." -ForegroundColor Red
    }
    
    # PDF operations tests
    Write-Host "`nPDF işlemleri testleri" -ForegroundColor Yellow
    Write-Host "Not: PDF yükleme testi için Invoke-RestMethod'un multipart/form-data desteği gereklidir." -ForegroundColor Yellow
    Write-Host "Bu test manuel olarak yapılmalıdır." -ForegroundColor Yellow
    
    # Get public PDFs
    $getPublicPDFsResponse = Test-Endpoint -Description "Herkese açık PDF'leri getirme" -Method "GET" -Endpoint "/api/v1/pdfs" -ExpectedStatus 200
    
    # Get user PDFs
    $getUserPDFsResponse = Test-Endpoint -Description "Kullanıcının PDF'lerini getirme" -Method "GET" -Endpoint "/api/v1/pdfs/my" -ExpectedStatus 200 -Token $token
    
    # Search PDFs
    $searchPDFsResponse = Test-Endpoint -Description "PDF arama" -Method "GET" -Endpoint "/api/v1/pdfs/search?q=test" -ExpectedStatus 200
}

Write-Host "`nTest tamamlandı!" -ForegroundColor Green
