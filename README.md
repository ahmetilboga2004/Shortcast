# Shortcast - 60 Saniyelik Podcast Platformu

Shortcast, kullanıcıların 60 saniyelik podcastler paylaşabildiği, keşfedebildiği ve dinleyebildiği bir podcast platformunun backend servisidir.

## 🚀 Özellikler

- 🎙️ 60 saniyelik podcast yükleme
- 🔍 Podcast keşfetme ve akış
- 📁 Kategori bazlı podcast arama
- ❤️ Beğeni sistemi
- 💬 Yorum sistemi
- 👤 Kullanıcı yönetimi
- 🔒 JWT tabanlı kimlik doğrulama
- 🚀 Yüksek performanslı önbellekleme

## 🛠️ Kullanılan Teknolojiler

- **Web Framework**: [Fiber](https://github.com/gofiber/fiber) - Yüksek performanslı Go web framework'ü
- **Veritabanı**: [PostgreSQL](https://www.postgresql.org/) - İlişkisel veritabanı
- **ORM**: [GORM](https://gorm.io/) - Go için ORM kütüphanesi
- **Önbellekleme**: [Redis](https://redis.io/) - İn-memory veri yapısı deposu
- **Object Storage**: [Cloudflare R2](https://www.cloudflare.com/products/r2/) - S3 uyumlu object storage
- **Kimlik Doğrulama**: [JWT](https://jwt.io/) - JSON Web Token
- **API Dokümantasyonu**: [Swagger](https://swagger.io/) - API dokümantasyonu

## 🏗️ Mimari

Proje, Clean Architecture prensiplerine uygun olarak tasarlanmıştır:

```
shortcast/
├── internal/
│   ├── config/      # Yapılandırma yönetimi
│   ├── container/   # Dependency Injection
│   ├── dto/         # Data Transfer Objects
│   ├── handler/     # HTTP istek işleyicileri
│   ├── middleware/  # HTTP middleware'ler
│   ├── model/       # Veritabanı modelleri
│   ├── repository/  # Veritabanı işlemleri
│   ├── router/      # Rota tanımlamaları
│   ├── service/     # İş mantığı
│   └── utils/       # Yardımcı fonksiyonlar
├── docs/            # API dokümantasyonu
└── main.go          # Uygulama giriş noktası
```

### Mimari Katmanlar

1. **Handler**: HTTP isteklerini alır ve yanıtları döndürür
2. **Service**: İş mantığını içerir
3. **Repository**: Veritabanı işlemlerini yönetir
4. **Model**: Veritabanı modellerini tanımlar
5. **DTO**: Veri transfer nesnelerini tanımlar

## 🚀 Kurulum

### Gereksinimler

- Go 1.21 veya üzeri
- PostgreSQL
- Redis
- Cloudflare R2 hesabı

### Adımlar

1. Projeyi klonlayın:
```bash
git clone https://github.com/yourusername/shortcast.git
cd shortcast
```

2. Bağımlılıkları yükleyin:
```bash
go mod download
```

3. `.env.example` dosyasını `.env` olarak kopyalayın ve gerekli değerleri doldurun:
```bash
cp .env.example .env
```

4. Veritabanını oluşturun:
```bash
createdb shortcast
```

5. Uygulamayı başlatın:
```bash
go run main.go
```

## ⚙️ Yapılandırma

`.env` dosyasında aşağıdaki değişkenleri ayarlayın:

```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=shortcast
SECRET_KEY=your_secret_key
JWT_EXPIRATION=3600
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0
R2_ACCOUNT_ID=your_r2_account_id
R2_ACCESS_KEY_ID=your_r2_access_key_id
R2_ACCESS_KEY_SECRET=your_r2_access_key_secret
R2_BUCKET_NAME=your_bucket_name
R2_ENDPOINT=https://your_account_id.r2.cloudflarestorage.com
```

## 📚 API Dokümantasyonu

API dokümantasyonuna erişmek için:
1. Uygulamayı başlatın
2. Tarayıcınızda `http://localhost:8080/docs` adresine gidin

## 🔄 Önbellekleme Stratejisi

- Podcast URL'leri Redis'te 24 saat boyunca önbelleğe alınır
- Her istekte yeni signed URL oluşturulmaz
- Redis bağlantısı koparsa sistem çalışmaya devam eder (graceful degradation)

## 🔒 Güvenlik

- JWT tabanlı kimlik doğrulama
- Şifreler hash'lenerek saklanır
- CORS yapılandırması
- Rate limiting
- Input validation

## 🤝 Katkıda Bulunma

1. Fork'layın
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Değişikliklerinizi commit edin (`git commit -m 'feat: add amazing feature'`)
4. Branch'inizi push edin (`git push origin feature/amazing-feature`)
5. Pull Request açın

## 📝 Lisans

Bu proje MIT lisansı altında lisanslanmıştır. Detaylar için [LICENSE](LICENSE) dosyasına bakın. 