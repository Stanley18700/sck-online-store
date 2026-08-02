# BC2 — Spending Points (2 points = 1.00 THB) — Team Walkthrough

This document explains, in simple words, what the `BC2` branch does and why.
Every blue link opens the real code. On GitHub, the link also jumps to the exact line.

> สรุป: เอกสารนี้อธิบาย BC2 — ใช้แต้มเป็นส่วนลด (2 แต้ม = 1.00 บาท)
> point-service เป็นคนคำนวณส่วนลด, เปิดใช้ช่อง Use your points ที่หน้า checkout,
> แก้บั๊กเงิน 2 ตัว และเทสต์ผ่าน 100% ทุกชั้น (รัน 2 รอบซ้ำได้)

---

## 1. The business condition

> **"2 reward points are worth 1.00 THB to be used as a discount on future purchases."**

Extra rules we agreed:
- Points burn **in pairs**. If you have 3 points, only 2 are used (1.00 THB). 1 point stays.
- The discount can never be bigger than the product subtotal.
- **Shipping cost is never discounted.**
- The discount is always a whole baht — no strange rounding is possible.

> สรุป: 2 แต้ม = 1 บาท ใช้เป็นคู่เท่านั้น ส่วนลดไม่เกินราคาสินค้า และไม่ลดค่าส่ง

---

## 2. How it works — same shape as BC1

The calculator still lives in **one room only**: point-service.

```
 store-web                store-service                point-service
 (the website)            (the shop backend)           (the CALCULATOR)
─────────────            ──────────────────           ─────────────────────
 checkbox "Use    ──►     asks point-service   ──►     usable = even(points),
 your points"             GET /point/discount          capped at subtotal
 shows -฿80.00   ◄──      never calculates    ◄──      discount = usable ÷ 2
```

### The calculator (point-service)

- The burn rate is one line:
  [point.constant.ts — line 2](../point-service/src/point/point.constant.ts#L2)
  → `BURN_RATE = 2` (2 points per 1.00 THB)
- The rule:
  [point.service.ts — line 38](../point-service/src/point/point.service.ts#L38)
  → round points down to a pair, cap at the subtotal, divide by 2.
- The door other services knock on:
  [point.controller.ts — line 59](../point-service/src/point/point.controller.ts#L59)
  → `GET /api/v1/point/discount?points=160&subtotal=500` answers
  `{ "burn_point": 160, "discount": 80 }`

### The messenger (store-service)

- The HTTP call: [gateway.go — line 71](../store-service/internal/point/gateway.go#L71)
- The website preview route (login required):
  [main.go — line 256](../store-service/cmd/main.go#L256)

### The display (store-web)

- Ticking the checkbox asks the backend for a quote — the website never does the math:
  [use-order-store.ts — line 177](../store-web/src/hooks/use-order-store.ts#L177)
- The red discount row at checkout:
  [order-summary.tsx — line 36](../store-web/src/app/checkout/components/order-summary.tsx#L36)

> สรุป: เหมือน BC1 — point-service คำนวณ (BURN_RATE=2), store-service ส่งต่อ, เว็บแค่แสดงผล

---

## 3. Two money bugs we fixed (the reasons behind the design)

### Bug 1: The server trusted the discount the customer sent ⚠️

Before, the order API accepted a `discount_price` field **from the browser** and subtracted
it. Anyone could edit the request and give themselves any discount. Worse: the server
multiplied that number by the USD exchange rate (35.97×) — sending `discount_price: 100`
would remove **฿3,597** from the order.

**Fix:** the server now ignores that field completely. It takes only `burn_point`, asks
point-service what the discount is, and rejects a burn the rule would not allow
([order.go — line 88](../store-service/internal/order/order.go#L88)).
Our test even sends a fake `discount_price: 999.99` and proves it is ignored.

### Bug 2: A failed point deduction still gave the discount for free ⚠️

Before, the points were deducted **after** the order was saved — and if the deduction failed,
the error was thrown away. The customer kept the points AND got the discount.

**Fix:** points are deducted **before** the order is written
([order.go — line 144](../store-service/internal/order/order.go#L144)).
If the deduction fails, the order fails. If saving the order fails afterwards, the points are
credited back ([order.go — line 155](../store-service/internal/order/order.go#L155)).

Also fixed on the way:
- The balance on the checkout page always showed 0, because the website called
  `PUT /api/v1/point` — a route that never existed. Now it is `GET`
  ([point.ts — line 15](../store-web/src/services/point.ts#L15)).
- The receipt PDF now shows a "Points Discount" line
  ([pdf_generator.go — line 120](../store-service/internal/order/pdf_generator.go#L120)).

> สรุป: แก้บั๊กเงิน 2 ตัว — เซิร์ฟเวอร์ไม่เชื่อส่วนลดจากเบราว์เซอร์อีกต่อไป (คำนวณเองจาก burn_point)
> และตัดแต้มก่อนบันทึกออเดอร์ ถ้าตัดไม่สำเร็จออเดอร์ไม่ผ่าน

---

## 4. One number that will surprise you: 84, not 86

The bicycle earns **86** points normally. But when you burn 160 points (฿80 discount),
the earning becomes **84**.

Why? Earning is calculated on the amount you actually pay for products:
`(4,314.60 − 80.00) ÷ 50 = 84.69 → 84 points.`

This is correct behavior — you earn on what you spend — and our end-to-end test asserts
exactly this number.

> สรุป: ถ้าใช้แต้มลดราคา แต้มที่ได้รับจะลดลงด้วย (คิดจากยอดหลังหักส่วนลด) — 84 ไม่ใช่ 86

---

## 5. The test scripts (board steps ③–⑦)

### Unit tests

| Layer | File | What it checks |
|---|---|---|
| Jest (point-service) | [point.service.spec.ts](../point-service/src/point/test/point.service.spec.ts) | the BVA set: 2 pts→฿1.00 (equal) · 1 pt→฿0.00 (less) · 160 pts→฿80.00 (more) · 3 pts→฿1.00 with 1 left · cap at subtotal · zero · negative |
| Go (store-service) | [order_test.go](../store-service/internal/order/order_test.go) | burn 8→฿4 with fake client discount ignored · odd burn rejected before anything is written · failed burn fails the order |
| Cypress (store-web) | [total-price.cy.ts](../store-web/src/__test__/total-price.cy.ts) | total composition at 2:1 — 100 pts on ฿500 → total ฿500 (was ฿450 under the old 1:1 mistake) |

### API tests (Newman) — [collection](../atdd/api/collections/004-Point-Discount.postman_collection.json)

| Folder | Type | What it does |
|---|---|---|
| TSS-PD-001 | Success | data-driven quote check, 7 rows ([data](../atdd/api/data/004-Point-Discount/TSS-PD-001.json)) |
| TSA-PD-001 | Alternative | `abc` / empty input → 400 "points must be a number" |
| TSS-PD-002 | Success, end-to-end | see below |

**TSS-PD-002 is self-cleaning** ([data](../atdd/api/data/004-Point-Discount/TSS-PD-002.json), user_6):
1. Mint exactly 160 points through the API (also proves the ledger was clean)
2. Order the bicycle with `burn_point: 160`
3. Assert: discount ฿80.00 · total ฿4,284.60 · receiving point **84**
4. Pay with OTP → Kerry tracking number
5. Assert the balance is **0** again

Because it burns exactly what it minted, you can run it **forever without resetting the
database**. We ran the whole suite twice in a row to prove it.

### UI test (Robot) — [TSS-PD-UI-001](../atdd/ui/004-Point-Discount/TSS-PD-UI-001-Order_with_point_discount_success.robot)

Same story through a real browser (user_7): the arrange step gives the user 160 points via
the API (python `requests`), then the test logs in, buys the bicycle, sees **"160 Points"**
next to the checkbox, ticks **Use your points**, sees the red **-฿80.00** row and the
**฿4,284.60** total, pays, gets a KR tracking number — and finally checks through the API
that the balance is back to 0.

### How to run

```bash
make run_newman_point_discount    # API suite
make run_robot_point_discount     # UI suite
```
Both are also inside `make run_newman` / `make run_robot`.

> สรุป: เทสต์ครบทุกชั้น — Jest 27, Go 8 แพ็กเกจ, Cypress 32, Newman 16 โฟลเดอร์ (423+ assertions),
> Robot 4 ชุด — และชุด 004 รันซ้ำ 2 รอบผ่านทั้งคู่เพราะเทสต์เก็บกวาดแต้มเอง

---

## 6. Results — everything passes ✅

| Test | Result |
|---|---|
| Jest (point-service) | 27 tests, 0 failed |
| Go unit (store-service) | 8 packages, all ok |
| Cypress (store-web) | 32 tests, 0 failed |
| Newman — 4 suites, 16 folders (004 run **twice**) | 0 failed |
| Robot — 001 + 002 + 004 (004 run **twice**) | 6 tests, 0 failed |

## 7. What is still NOT done (BC3 / BC4)

- Points are still never **earned into the ledger** — the balance only grows via the
  test/mint API. Earning-for-real arrives with BC4 (confirm receipt → Approved).
- No 180-day expiry, no point statuses yet.
- The mint behavior of `POST /point` is a known bug we deliberately kept for now — the BC2
  tests use it as their arrange step. It becomes a real credit flow in BC4.
- The ledger is still one global pot (user id is hardcoded) — planned for BC4.

> สรุป: ยังเหลือ BC3/BC4 — แต้มยังไม่ถูก "ได้รับ" จริงเข้ากระเป๋า, ยังไม่มีวันหมดอายุและสถานะแต้ม

---

*Branch `BC2` on the team repo. Built on top of BC1 (740f6a1).*
