# boltcli ⚡🍔

> Control your **Bolt Food** orders straight from the terminal.

```
$ boltcli history
289516420   New York Burger and Chicken KV   delivered   2026-02-20 20:04
287858187   New York Burger and Chicken KV   delivered   2026-02-12 21:20
247050348   KFC Karlovy Vary DT              delivered   2025-07-13 18:16
246490802   McDonald's Dolní Kamenná KV      delivered   2025-07-10 18:59
```

---

## ✨ Features

| Command | Description |
|---------|-------------|
| `login` | 📱 Sign in via phone number + SMS OTP |
| `history` | 📜 Browse your past orders |
| `orders` | 🔴 See active / in-progress orders |
| `order <id>` | 🧾 Full details of a specific order |
| `logout` | 👋 Clear stored credentials |

---

## 📦 Installation

Requires [Go 1.21+](https://go.dev/dl/)

```sh
git clone https://github.com/Lukynnnn/boltcli.git
cd boltcli
go build -o boltcli .
```

Optionally move to your PATH:

```sh
mv boltcli /usr/local/bin/
```

---

## 🚀 Quick Start

**1. Login**
```sh
boltcli login --phone +420XXXXXXXXX
# Enter the SMS code when prompted
```

**2. View order history**
```sh
boltcli history
boltcli history --limit 50
```

**3. Check order details**
```sh
boltcli order 289516420
```

**4. Active orders**
```sh
boltcli orders
```

---

## 🔧 How It Works

Bolt Food uses a private REST API at `https://deliveryuser.live.boltsvc.net`.
Auth is SMS OTP only — no OAuth dance, no client secrets needed.
After login, your `access_token` is saved locally and sent as a Bearer token on every request.

> Reverse-engineered via [mitmproxy](https://mitmproxy.org/) 🕵️

---

## 📁 Config

Credentials are stored at:

```
~/.config/boltcli/config.json   # Linux
~/Library/Application Support/boltcli/config.json   # macOS
```

---

## 📄 License

MIT © [Lukáš Molčan](https://github.com/Lukynnnn)
