# BC-1 — Point Calculation via Point-Service (Team Walkthrough)

This document explains, in simple words, what the `BC-1` branch does and why.
Every blue link opens the real code. On GitHub, the link also jumps to the exact line.

> สรุป: เอกสารนี้อธิบายว่า BC-1 ทำอะไร — ย้ายการคำนวณแต้มไปที่ point-service,
> แก้บั๊กที่อาจารย์ซ่อนไว้ 3 ตัว, เปลี่ยนอัตราเป็น 50 บาท = 1 แต้ม, และทำให้เทสต์ผ่าน 100%

---

## 1. What was the task?

From the whiteboard, our instruction had 8 steps:

1. Change the point rate: **100 THB → 50.00 THB = 1 point**
2. The point calculation must be done by **point-service**, not store-service
3. Write API tests (Postman)
4. Write UI tests (Robot Framework)
5. Run the API tests → they **fail first** (this is normal and expected!)
6. Fix the code until tests pass **100%**
7. Run the UI tests until they pass **100%**
8. Commit to the repo

**All 8 steps are done.** This branch is the result.

> สรุป: งานคือเปลี่ยนอัตราแต้มเป็น 50 บาทต่อ 1 แต้ม และให้ point-service เป็นคนคำนวณ
> ตอนนี้ทำครบทั้ง 8 ขั้นตอนแล้ว

---

## 2. How does the calculation work now?

Think of it like this: **the calculator lives in one room only.**
That room is point-service. Everyone else must knock on its door and ask.

```
 store-web              store-service               point-service
 (the website)          (the shop backend)          (the CALCULATOR)
─────────────          ──────────────────          ─────────────────
 shows the number  ──►  asks point-service  ──►     amount ÷ 50
 never calculates       never calculates            answers: { "point": 86 }
```

### The calculator (point-service)

- The rate is **one line of code**:
  [point.constant.ts — line 1](../point-service/src/point/point.constant.ts#L1)
  → `POINT_RATE = 50`
- The whole rule is **two lines**:
  [point.service.ts — line 20](../point-service/src/point/point.service.ts#L20)
  → take the amount, divide by 50, round down. Negative amount → 0 points.
- Other services call it through this door:
  [point.controller.ts — line 23](../point-service/src/point/point.controller.ts#L23)
  → `GET /api/v1/point/calculate?amount=4314.6` answers `{ "point": 86 }`

### The messenger (store-service)

store-service does **not** calculate anymore. It only sends the amount and waits for the answer:

- The HTTP call to point-service:
  [gateway.go — line 44](../store-service/internal/point/gateway.go#L44)
- Used by the cart page: [cart.go — line 49](../store-service/internal/cart/cart.go#L49)
- Used when creating an order: [order.go — line 110](../store-service/internal/order/order.go#L110)
- The website can also ask through this route:
  [main.go — line 255](../store-service/cmd/main.go#L255) (login required)

### The display (store-web)

The website only **shows** the number it receives:
[calculate-point.ts — line 13](../store-web/src/services/calculate-point.ts#L13)

### How do we know store-service does not calculate?

Because the old calculator files are **deleted**:
- `store-service/internal/common/point.go` — gone
- `store-web/src/utils/point.ts` — gone

If you search the whole project for "divide by 50", you find it **only** in point-service.

**Why is this good?** Before, the rate 100 was written in 3 different places (Go code,
website code, tests). They could disagree — and they did (see next section). Now, changing
the rate = changing **one line, in one file, in one service**.

> สรุป: point-service เป็นคนคำนวณคนเดียว (หาร 50 ปัดเศษลง) — store-service แค่ส่งจำนวนเงินไปถาม
> แล้วเว็บก็แค่แสดงผล ไฟล์คำนวณเก่าถูกลบหมดแล้ว

---

## 3. Problems we found and fixed

### Problem 1: The rate secretly went back to 100 ⚠️ (most important)

The team changed 100 → 50 (commit `8b53620`). Good.
But the new point-service code was written with **100** again, and the old 50 code was deleted.
Result: the rate went back to 100 — and **no test failed**, because the tests also said 100.

**Fix (commit `0e6f1c8`):** change
[point.constant.ts — line 1](../point-service/src/point/point.constant.ts#L1) to 50,
and update all the test numbers to match (43 → 86, 9 → 18, 52 → 104, and so on).

**Lesson:** when code moves to a new service, the *number inside it* does not move
automatically. A person must carry it.

### Problem 2: The Docker build of the website was broken

A merge brought back a deleted test file (`point.cy.ts`) that imports a file which no longer
exists. On our laptops (`npm run dev`) everything looked fine. But Docker build failed.
**Fix (commit `fd37329`):** delete the file.

### Problem 3: One test number was hiding inside the Postman collection

Almost all expected numbers live in the data files (`atdd/api/data/...`).
But one number was hardcoded inside the collection itself:
[001-Authentication.postman_collection.json — line 2617](../atdd/api/collections/001-Authentication.postman_collection.json#L2617)
**Fix (commit `fe27d64`):** 86 → 172 (a cart with two bicycles).
**Lesson:** when updating test numbers, search the collection files too, not only the data files.

> สรุป: เจอ 3 ปัญหา — อัตราแอบกลับไปเป็น 100 (แก้เหลือบรรทัดเดียว), Docker build เว็บพัง,
> และมีเลขเทสต์ซ่อนอยู่ใน collection ไม่ใช่ data file

---

## 4. The instructor's hidden bugs (the traps) 🪤

The instructor planted 3 bugs on purpose. One commit message even says it directly:
*"[Added] bug for product id 3 that show minus sign(-) for price and point"* (commit `ddfdf90`).

### Trap 1: Product 3 shows a minus price and 0 points

The website code **flipped the price to negative** after receiving it from the API
(only on the detail page). The negative price then went to the calculator, and the
calculator correctly answers 0 for negative amounts. So "0 Points" was not the real bug —
it was a *symptom*.

**How we caught it:** the product **list** page showed the correct price (+฿897.45).
Only the **detail** page showed minus. Same product, two different prices → the bug must be
in the detail page code. **Fix: commit `6abf049`.**

### Trap 2: Product 7's page always shows an error

Hidden code: `if ID == 7 → return error`. But you could still **buy** product 7,
because ordering uses a different code path.
**How we caught it:** a product you can buy but cannot look at is very strange.
**Fix: commit `85bc2ef`.**

### Trap 3: Product 8's price in the cart never matches the total

Hidden code added **+0.01** to the displayed price only. Line shows ฿717.61,
total shows ฿717.60. Made to break exact-value test assertions.
**Fix: commit `6ea56f9`.**

To see any fix: `git show <commit>` in the terminal.

Also good to know (we did **not** change these — they are valid test data):
- 2 products with price 0 (ids 1044, 1339)
- 8 products at 1.39/2.78 USD that sit exactly on the 50-baht point boundary

> สรุป: อาจารย์ซ่อนบั๊กไว้ 3 ตัว — สินค้า 3 ราคาติดลบ (เว็บกลับเครื่องหมายเอง),
> สินค้า 7 เปิดหน้าไม่ได้แต่ซื้อได้, สินค้า 8 ราคาในตะกร้าไม่ตรงกับยอดรวม แก้หมดแล้ว

---

## 5. Test results — everything passes ✅

The tests **failed first** (step 5 of the board — this is the ATDD way), then we fixed
until green:

| Test | Result |
|---|---|
| Postman/Newman — 3 suites, 10 folders | **423 assertions, 0 failed** |
| Robot UI — 2 suites | **6 tests, 0 failed** |
| Go unit tests | 8 packages, all pass |
| Jest (point-service) | 17 tests, all pass |
| Cypress (store-web) | 34 tests, all pass |

Check in the browser: the Balance Training Bicycle now shows **86 points**
(4,314.60 ÷ 50 = 86).

**Question for the instructor:** the whiteboard example said **80 points**, using the
exchange rate 33.52. Our code uses the old hardcoded rate 35.969964
([currency.go — line 11](../store-service/internal/common/currency.go#L11)), which gives 86.
Which rate should the class use? This decision changes every expected number.

> สรุป: เทสต์ผ่านหมดทุกชั้น รถจักรยานได้ 86 แต้ม — แต่ต้องถามอาจารย์เรื่องเรต 33.52 (ได้ 80)
> กับ 35.97 (ได้ 86) ว่าจะใช้ตัวไหน

---

## 6. What is NOT done yet

BC-1 covers only the **first** business condition. Still to do:

| Business condition | What is missing |
|---|---|
| 2 points = 1.00 THB discount | spend code still thinks 1 point = 1 THB ([total-price.ts — line 18](../store-web/src/utils/total-price.ts#L18)); the discount box at checkout is switched off ([view.tsx — line 91](../store-web/src/app/checkout/view.tsx#L91)) |
| Points valid 180 days | the points table has **no date columns** at all ([point.entity.ts](../point-service/src/point/point.entity.ts)) |
| Approve on confirm receipt | this button/endpoint does not exist yet; earned points are **never saved** to the point ledger |
| Point status (Pending / Approved / Redeemed / Expired) | no status exists anywhere yet |

Also known (not fixed on purpose, team decision): if point-service is down, the cart page
shows an error; the HTTP call has no timeout.

Small test gap: the requirement's own three examples (50.00 → 1, 45.00 → 0, 389.00 → 7)
are not in the test data yet — three rows to add in
[TSS-PC-001.json](../atdd/api/data/003-Point-Calculate/TSS-PC-001.json).

> สรุป: เสร็จแค่เงื่อนไขแรก ยังเหลือ ฝั่งใช้แต้ม (2 แต้ม = 1 บาท), วันหมดอายุ 180 วัน,
> ปุ่มยืนยันรับสินค้า และสถานะแต้ม 4 แบบ

---

## 7. About the `Pai` branch

Please **do not merge `Pai` yet**. Reasons:
- It conflicts with the newest code (the Robot cleanup keyword)
- Its point numbers are already included in BC-1
- Its download folder is never cleaned → files pile up forever

Better plan: rebase `Pai` on top of BC-1, keep only the unique-folder idea, and add a
cleanup step.

> สรุป: อย่าเพิ่ง merge branch Pai — ให้ rebase ทับ BC-1 ก่อน

---

*Branch `BC-1` is on the NyanSintZaw fork. PR #3 targets Stanley's `BC1` branch.*
