# 🛵 boltcli — Bolt Food v terminálu

Go CLI nástroj pro správu Bolt Food objednávek z příkazové řádky.
API reverse-engineerováno pomocí mitmproxy.

## Funkce

- `login` — přihlášení přes telefon + SMS OTP
- `history` — seznam minulých objednávek
- `orders` — aktivní objednávky
- `order` — detail konkrétní objednávky
- `logout` — odhlášení

## Instalace

### Požadavky

- [Go 1.21+](https://go.dev/dl/)

### Build

```sh
git clone https://github.com/Lukynnnn/boltcli.git
cd boltcli
go build -o boltcli .
```

## Použití

### Přihlášení

```sh
./boltcli login --phone +420XXXXXXXXX
```

Zadej SMS kód a jsi přihlášen. Token se uloží do `~/.config/boltcli/config.json`.

### Historie objednávek

```sh
./boltcli history
./boltcli history --limit 50
```

### Detail objednávky

```sh
./boltcli order 289516420
```

### Aktivní objednávky

```sh
./boltcli orders
```

### Odhlášení

```sh
./boltcli logout
```

## Jak to funguje

Bolt Food používá REST API na `https://deliveryuser.live.boltsvc.net`.
Autentizace probíhá přes SMS OTP — žádný OAuth, žádný client secret.
Token se ukládá lokálně a posílá jako `Authorization: Bearer <token>`.

## Poznámky

- Testováno na Bolt Food CZ (Karlovy Vary)
- Nevyžaduje žádné API klíče ani registraci vývojáře
- Token z přihlášení vydrží dlouho (JWT s dlouhou expirací)
