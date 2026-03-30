## r/golang — post draft

**Title:**
I reverse-engineered the Bolt Food API with mitmproxy and built a CLI in Go

**Body:**
Inspired by steipete's ordercli (for Foodora/Deliveroo), I wanted the same
thing for Bolt Food.

Captured the traffic with mitmproxy on my iPhone, figured out the auth flow
(SMS OTP → Bearer JWT, no OAuth, no client secrets needed), and built a small
CLI in Go.

What it does:

    $ boltcli login --phone +420XXXXXXXXX
    # enter SMS code

    $ boltcli history
    289516420   New York Burger and Chicken KV   delivered   2026-02-20 20:04
    287858187   New York Burger and Chicken KV   delivered   2026-02-12 21:20
    247050348   KFC Karlovy Vary DT              delivered   2025-07-13 18:16

    $ boltcli order 289516420
    Order:      BGL5M
    Restaurant: New York Burger and Chicken KV
    Status:     delivered
    Items:
      1x Small Crunchy Chicken Burger (79)
      1x Small Bacon Cheeseburger (85)
    Total:      228

Commands: login, history, orders, order, logout
Reorder coming soon (need to capture more flows).

https://github.com/Lukynnnn/boltcli
