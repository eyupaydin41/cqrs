# 📮 Postman Collection Kullanım Kılavuzu

## 🚀 Hızlı Başlangıç

### 1. Collection'ı Import Etme

1. Postman'i açın
2. **Import** butonuna tıklayın
3. `CQRS-EventSourcing.postman_collection.json` dosyasını seçin
4. Import tamamlandı! ✅

### 2. Environment Variables

Collection otomatik olarak şu değişkenleri kullanır:

| Variable | Default Value | Açıklama |
|----------|--------------|----------|
| `auth_service_url` | `http://localhost:8088` | Auth Service URL |
| `query_service_url` | `http://localhost:8089` | Query Service URL |
| `event_store_url` | `http://localhost:8090` | Event Store URL |
| `user_id` | (auto-set) | Son oluşturulan user ID |
| `jwt_token` | (auto-set) | Login'den dönen JWT token |

**Not:** `user_id` ve `jwt_token` otomatik olarak kaydedilir!

---

## 📚 Collection Yapısı

### 🔐 Auth Service (COMMAND) - Port 8088
**Write operations (Command side of CQRS)**

- ✅ **Health Check** - Servis durumu
- ➕ **Register User** - Yeni kullanıcı kaydı
  - **gRPC:** Hayır (sadece Kafka publish)
  - **Event:** `user.created`
- 🔑 **Change Password** - Şifre değiştirme
  - **gRPC:** ✅ Evet! (Event-Store'dan aggregate history çeker)
  - **Event:** `user.password.changed`
- 📧 **Change Email** - Email değiştirme
  - **gRPC:** ✅ Evet!
  - **Event:** `user.email.changed`

### 🔍 Query Service (QUERY) - Port 8089
**Read operations (Query side of CQRS)**

- ✅ **Health Check**
- 👥 **Get All Users** - Tüm kullanıcıları listele (Read Model'den)
- 🔓 **Login** - Kullanıcı girişi (JWT döner)

### 📦 Event Store - Port 8090
**Event storage and retrieval**

- ✅ **Health Check**
- 📋 **Get All Events** - Tüm event'leri listele (filtering ile)
- 🎯 **Get Events by Aggregate** - Belirli bir aggregate'in tüm event'leri
- 🔢 **Get Event Count** - Toplam event sayısı
- ⏪ **Replay Events Since** - Belirli tarihten sonraki event'ler

### ⏰ Time Travel & Replay - Port 8090
**Event Sourcing'in süper güçleri!**

- 📸 **Get Current User State** - Şu anki kullanıcı durumu (event'lerden reconstruct)
- ⏰ **Get User State at Point in Time** - Belirli bir tarihteki kullanıcı durumu
- 📜 **Get User History** - Tüm değişiklik geçmişi
- 🔄 **Compare States** - İki tarih arasındaki farkları göster

### 📸 Snapshots - Port 8090
**Performance optimization**

- ➕ **Create Snapshot** - Snapshot oluştur
- 📥 **Get Latest Snapshot** - Son snapshot'ı getir
- 🎯 **Get Aggregate State** - Snapshot + son event'lerle state'i getir

### 🎯 Complete User Journey
**Tüm akışı adım adım test et!**

1. Register User
2. Wait 2 seconds (Kafka processing)
3. Login
4. Change Password (gRPC demo!)
5. View Event History
6. Time Travel
7. Create Snapshot

---

## 🎮 Kullanım Örnekleri

### Senaryo 1: Yeni Kullanıcı Kaydı ve Şifre Değişikliği

1. **Register User**
   ```json
   POST /register
   {
     "email": "test@example.com",
     "password": "pass123"
   }
   ```
   - Response'dan `user_id` otomatik kaydedilir!

2. **2-3 saniye bekleyin** (Kafka consumer'ların event'i işlemesi için)

3. **Change Password** (gRPC çalışacak!)
   ```json
   PUT /users/{{user_id}}/password
   {
     "old_password": "pass123",
     "new_password": "newpass456"
   }
   ```
   - Auth-Service → gRPC call → Event-Store
   - Event history yüklenir
   - Aggregate reconstruct edilir
   - Şifre değiştirilir

### Senaryo 2: Time Travel

1. **Register User** → User ID: `abc-123`

2. **Change Password** (19:58:00)

3. **Change Email** (19:59:00)

4. **Time Travel - See state at 19:58:30** (password değişmiş ama email henüz değişmemiş)
   ```
   GET /replay/user/abc-123/state-at?timestamp=2025-10-25T19:58:30Z
   ```
   Sonuç: Yeni şifre, eski email! ⏰

5. **Compare States**
   ```
   GET /replay/user/abc-123/compare?time1=2025-10-25T19:57:00Z&time2=2025-10-25T20:00:00Z
   ```
   Sonuç: Hem password hem email değişmiş!

### Senaryo 3: Event History İncelemesi

1. **Get Events by Aggregate**
   ```
   GET /events/aggregate/{{user_id}}
   ```
   Sonuç:
   ```json
   {
     "events": [
       {
         "event_type": "user.created",
         "version": 1,
         "timestamp": "2025-10-25T19:57:00Z"
       },
       {
         "event_type": "user.password.changed",
         "version": 2,
         "timestamp": "2025-10-25T19:58:00Z"
       },
       {
         "event_type": "user.email.changed",
         "version": 3,
         "timestamp": "2025-10-25T19:59:00Z"
       }
     ]
   }
   ```

### Senaryo 4: Snapshot Performance Test

1. **Register User**

2. **Change Password 5 kez** (5 event)

3. **Get User State** (6 event replay edilecek - yavaş)
   ```
   GET /replay/user/{{user_id}}/state
   ```

4. **Create Snapshot** (Şu anki state'i snapshot olarak kaydet)
   ```
   POST /snapshots/{{user_id}}
   ```

5. **Change Password 2 kez daha** (2 yeni event)

6. **Get Aggregate State with Snapshot** (snapshot + 2 event - çok hızlı!)
   ```
   GET /snapshots/{{user_id}}/state
   ```

---

## 🔍 gRPC İletişimini Görüntüleme

gRPC call'ları görmek için Docker loglarını kontrol edin:

```bash
# Auth-Service logs (gRPC client)
docker logs cqrs-auth-service-1 --follow

# Event-Store logs (gRPC server)
docker logs cqrs-event-store-1 --follow
```

**gRPC call yapıldığında göreceğiniz loglar:**

**Auth-Service:**
```
🔄 Loading aggregate abc-123 from event-store via gRPC...
gRPC Call: GetAggregateEvents for aggregate_id=abc-123
gRPC Response: Received 2 events
📦 Reconstructing aggregate from 2 events
✅ Aggregate loaded: Status=active, Email=test@example.com
```

**Event-Store:**
```
gRPC: GetAggregateEvents called for aggregate_id: abc-123
retrieved 2 events for aggregate: abc-123
```

---

## 🎯 Önerilen Test Sırası

### 1. Basic Flow (5 dakika)
1. ✅ Health Check (tüm servisler)
2. ➕ Register User
3. 👥 Get All Users
4. 🔓 Login
5. 📋 Get Events by Aggregate

### 2. CQRS + Event Sourcing (10 dakika)
1. ➕ Register User
2. 🔑 Change Password (gRPC!)
3. 📧 Change Email (gRPC!)
4. 📋 Get Events by Aggregate
5. 📸 Get Current User State

### 3. Time Travel (10 dakika)
1. ➕ Register User
2. ⏰ Get User State at Point in Time (register anı)
3. 🔑 Change Password
4. ⏰ Get User State at Point in Time (password değişmeden önce)
5. 🔄 Compare States (öncesi vs sonrası)
6. 📜 Get User History

### 4. Performance (Snapshots) (10 dakika)
1. ➕ Register User
2. 🔑 Change Password (3-4 kez)
3. 📸 Get Current User State (tüm event'leri replay eder)
4. ➕ Create Snapshot
5. 🔑 Change Password (1 kez daha)
6. 🎯 Get Aggregate State (snapshot + son event - hızlı!)

### 5. Complete Journey (15 dakika)
"🎯 Complete User Journey" klasöründeki tüm request'leri sırasıyla çalıştır!

---

## 🛠️ Troubleshooting

### Problem: `user_id` değişkeni boş
**Çözüm:** "Register User" request'ini tekrar çalıştırın. Post-response script otomatik olarak kaydedecek.

### Problem: Kafka event'i işlenmedi
**Çözüm:** 2-3 saniye bekleyin. Kafka consumer asynchronous çalışır.

```bash
# Kafka consumer loglarını kontrol et
docker logs cqrs-event-store-1 --tail 20
docker logs cqrs-query-service-1 --tail 20
```

### Problem: gRPC connection failed
**Çözüm:** Event-Store'un gRPC server'ının çalıştığını kontrol edin:

```bash
docker logs cqrs-event-store-1 | grep "gRPC"
# Görmeli: "🚀 gRPC server starting on port 9090"
```

### Problem: Time Travel timestamp hatası
**Çözüm:** RFC3339 formatı kullanın:
- ✅ Doğru: `2025-10-25T19:58:00Z`
- ❌ Yanlış: `2025-10-25 19:58:00`

### Problem: Password change "invalid password" hatası
**Çözüm:**
1. Önce "Get Events by Aggregate" ile event history'yi kontrol edin
2. Doğru `old_password` kullanın
3. User'ın `status` field'ı `active` olmalı

---

## 📊 Beklenen Response'lar

### Başarılı Register
```json
{
  "id": "abc-def-123",
  "message": "User registered successfully. Please query from query-service."
}
```

### Başarılı Password Change
```json
{
  "message": "Password changed successfully"
}
```

### Başarılı Login
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": "abc-def-123",
  "email": "test@example.com"
}
```

### Event History
```json
{
  "aggregate_id": "abc-def-123",
  "events": [
    {
      "id": "evt-001",
      "event_type": "user.created",
      "version": 1,
      "timestamp": "2025-10-25T19:57:00Z",
      "data": "{\"email\":\"test@example.com\"}"
    },
    {
      "id": "evt-002",
      "event_type": "user.password.changed",
      "version": 2,
      "timestamp": "2025-10-25T19:58:00Z"
    }
  ],
  "count": 2
}
```

### Time Travel State
```json
{
  "user_id": "abc-def-123",
  "point_in_time": "2025-10-25T19:57:30Z",
  "state": {
    "id": "abc-def-123",
    "email": "test@example.com",
    "status": "active",
    "version": 1
  },
  "message": "State reconstructed at specified time"
}
```

---

## 🎓 Collection Features

### ✨ Auto-Save Variables
- **user_id** otomatik kaydedilir (Register response'dan)
- **jwt_token** otomatik kaydedilir (Login response'dan)
- Diğer request'lerde `{{user_id}}` olarak kullanılır

### 📝 Detailed Descriptions
Her endpoint için:
- Ne yaptığı açıklaması
- gRPC kullanıp kullanmadığı
- Hangi event'i publish ettiği
- Use case'ler
- Data flow diyagramı

### 🎯 Pre/Post Scripts
Bazı request'lerde:
- **Pre-request:** Hazırlık (örn: wait)
- **Post-response:** Variables kaydetme (user_id, jwt_token)

---

## 📚 Ek Kaynaklar

### Docker Commands
```bash
# Servisleri başlat
docker-compose up -d

# Logları izle
docker-compose logs -f

# Belirli servisin logları
docker logs cqrs-auth-service-1 -f

# gRPC loglarını filtrele
docker logs cqrs-auth-service-1 | grep "gRPC"
```

### ClickHouse Event'leri Görüntüleme
```bash
docker exec cqrs-clickhouse-1 clickhouse-client --query \
  "SELECT event_type, version, timestamp FROM events.events WHERE aggregate_id = 'USER_ID_BURAYA' ORDER BY version"
```

### Kafka Topic İçeriği
```bash
docker exec cqrs-kafka-1 kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic user-events \
  --from-beginning
```

---

## 🎉 İyi Testler!

Collection'daki tüm özellikler çalışır durumda. Herhangi bir sorun yaşarsanız:

1. Docker container'ların çalıştığını kontrol edin
2. Logları inceleyin
3. Kafka'nın event'leri işlediğini doğrulayın

**Pro Tip:** "🎯 Complete User Journey" klasöründeki request'leri sırasıyla çalıştırarak tüm mimariyi test edebilirsiniz!
