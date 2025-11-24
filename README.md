# Zalando Autobuyer

En automatiserad köpapplikation för Zalando.se skriven i Go.

## ⚠️ Disclaimer

**ENDAST FÖR UTBILDNINGSÄNDAMÅL**

Detta projekt är skapat för att demonstrera API-integrationer och automatisering. Det är **inte** avsett för kommersiell användning eller att kringgå Zalandos användarvillkor. Användning av denna programvara är på egen risk. Utvecklarna tar inget ansvar för eventuella konsekvenser av användning.

## 📋 Funktioner

- ✅ Automatisk inloggning med CSRF-tokenhantering
- ✅ Produktsökning och storleksval
- ✅ Lägg till i varukorg via GraphQL API
- ✅ Automatisk checkout-process
- ✅ Sökning och val av upphämtningsställe (Instabox/Budbee prioriteras)
- ✅ BNPL (Faktura) betalningsmetod
- ✅ Komplett köpflöde
- ✅ Retry-logik för rate limiting
- ✅ Färgkodad loggning
- ✅ Användarvänligt CLI-gränssnitt

## 🚀 Installation

### Förutsättningar

- Go 1.21 eller senare

### Steg

1. Klona repositoriet:
```bash
git clone https://github.com/walamagki446-sudo/zalando-autobuyer.git
cd zalando-autobuyer
```

2. Installera beroenden:
```bash
go mod download
```

## 📖 Användning

### Köra programmet

```bash
go run .
```

### Bygg körbar fil

För Linux/macOS:
```bash
go build -o zalando-autobuyer
./zalando-autobuyer
```

För Windows:
```bash
go build -o zalando-autobuyer.exe
zalando-autobuyer.exe
```

### Cross-compilation för Windows (från Linux/macOS):
```bash
GOOS=windows GOARCH=amd64 go build -o zalando-autobuyer.exe
```

## 🔄 Köpflöde

Programmet guidar dig genom följande steg:

1. **Inloggning**
   - Ange din e-postadress
   - Ange ditt lösenord
   - Automatisk CSRF-tokenhantering

2. **Produktval**
   - Klistra in Zalando produkt-URL
   - Välj önskad storlek från tillgängliga alternativ

3. **Leveransadress**
   - Ange gatuadress
   - Ange stad
   - Ange postnummer

4. **Checkout**
   - Automatisk checkout-session skapas

5. **Upphämtningsställe**
   - Lista över närliggande upphämtningsställen visas
   - Instabox och Budbee prioriteras
   - Välj manuellt eller låt programmet välja automatiskt

6. **Betalning**
   - BNPL (Faktura) väljs automatiskt

7. **Bekräftelse**
   - Bekräfta köpet
   - Få bekräftelse på slutfört köp

## 🔌 API-endpoints

Programmet använder följande Zalando API-endpoints:

### Autentisering
- `POST https://accounts.zalando.com/api/sso/authentications/credentials`
- `GET https://accounts.zalando.com/authenticate`

### Produkter
- `GET https://www.zalando.se/[produkt-url]`
- `POST https://www.zalando.se/api/graphql/add-to-cart/`

### Varukorg
- `GET https://www.zalando.se/api/cart-gateway/carts`

### Checkout
- `GET https://www.zalando.se/checkout/v3/fetch-or-create-checkout-trampoline`
- `POST https://www.zalando.se/api/checkout/search-pickup-points-by-address`
- `POST https://www.zalando.se/api/checkout/select-pickup-point`
- `POST https://www.zalando.se/api/checkout/next-step`

### Betalning
- `GET https://www.zalando.se/api/checkout/payment-methods`
- `POST https://www.zalando.se/api/checkout/update-payment`
- `POST https://www.zalando.se/checkout/payment-complete`

## 🛡️ Säkerhetsfunktioner

- CSRF-tokenhantering
- Cookie-baserad sessionshantering
- TLS 1.2+ kryptering
- User-Agent spoofing
- Rate limiting hantering med exponentiell backoff
- Retry-logik för "elevated-risk" scenarios

## 📝 Filstruktur

```
zalando-autobuyer/
├── main.go          # Huvudprogrammet med CLI-interface
├── auth.go          # Autentisering och session
├── product.go       # Produkthantering och varukorg
├── checkout.go      # Checkout och upphämtningsställen
├── payment.go       # Betalningshantering
├── utils.go         # Hjälpfunktioner och loggning
├── go.mod           # Go-modulberoenden
├── .gitignore       # Git ignore-regler
└── README.md        # Denna fil
```

## 🔧 Tekniska detaljer

### Beroenden
- `golang.org/x/net` - För cookie jar och publicsuffix

### HTTP-klient
- Cookie jar för sessionshantering
- 30 sekunders timeout
- TLS 1.2+ konfiguration
- Automatisk redirect-hantering

### Headers
- User-Agent: Chrome 131 på Windows 10
- Sec-Fetch-* headers för säkerhet
- CSRF-token i alla autentiserade requests
- Content-Type och Accept headers

## 🐛 Felsökning

### Problem: "CSRF token not found"
- Kontrollera din internetanslutning
- Zalando kan ha ändrat sin HTML-struktur

### Problem: "Inloggning misslyckades (401)"
- Kontrollera e-post och lösenord
- Vänta några minuter innan du försöker igen (rate limiting)

### Problem: "Rate limit nådd (429)"
- Vänta 30-60 sekunder innan du försöker igen
- Programmet försöker automatiskt med exponentiell backoff

### Problem: "Inga upphämtningsställen hittades"
- Kontrollera att adressen är korrekt
- Prova en annan adress i närheten

## 📄 Licens

Detta projekt är endast för utbildningsändamål. Använd på egen risk.

## 🤝 Bidrag

Detta är ett utbildningsprojekt. Bidrag är inte öppna för tillfället.

## 📧 Kontakt

För frågor eller problem, öppna en issue i GitHub-repositoriet.

---

**Kom ihåg:** Använd denna programvara ansvarsfullt och respektera Zalandos användarvillkor och rate limits.
