# BC-1 Walkthrough — Point Calculation via Point-Service

> Speaking script for the team + instructor demo. Branch:
> [`BC-1`](https://github.com/NyanSintZaw/sck-online-store/tree/BC-1) · 6 commits on top of
> team main `440ebc2`. Each section maps to the instruction board (steps ①–⑧).

---

## 1. The instruction, and where we started

Our task: **point calculation must come from point-service, not store-service** (steps ①②),
write API tests in Postman and UI tests in Robot (③④), execute until 100% pass (⑤–⑦), commit
(⑧). The business condition under test is **BC1: "For every 50.00 THB spent on purchased
products, 1 reward point is earned."**

The team had already landed the architecture on main: a new
[`GET /api/v1/point/calculate`](https://github.com/NyanSintZaw/sck-online-store/blob/BC-1/point-service/src/point/point.controller.ts)
endpoint in point-service, store-service delegating to it through
[`gateway.go`](https://github.com/NyanSintZaw/sck-online-store/blob/BC-1/store-service/internal/point/gateway.go),
the old Go calculator **deleted**, and the frontend showing backend values only. That part was
done and correct.

## 2. What we found when we audited it (say this to the team)

**Finding 1 — the rate had silently reverted to 100.**
The refactor branch was cut *from* the `100 → 50` commit (`8b53620`) but reintroduced the rate
as `POINT_RATE = 100` in the new constant file, then deleted every rate-50 source. No commit
message mentioned it, and since all test fixtures also said 100, nothing failed — the revert
was invisible. Lesson: when a rule moves between services, its *value* travels by hand.

Fix: [`0e6f1c8`](https://github.com/NyanSintZaw/sck-online-store/commit/0e6f1c8) — the
semantic change is **one line**,
[`point.constant.ts`](https://github.com/NyanSintZaw/sck-online-store/blob/BC-1/point-service/src/point/point.constant.ts):
`POINT_RATE = 50`. The other 15 files in that commit are expectations catching up
(Jest specs, Newman data, Robot values, Go mocks). That one-line blast radius is exactly what
the microservice ownership bought us.

**Finding 2 — main's Docker build was broken.**
The merge resurrected `point.cy.ts`, which imports the deleted `utils/point.ts`. `npm run dev`
never compiles orphaned test files, so nobody saw it locally — but `next build` (Docker) fails.
Fix: [`fd37329`](https://github.com/NyanSintZaw/sck-online-store/commit/fd37329).

**Finding 3 — one hidden expectation lives in the collection, not the data.**
`TSS-AUTH-003` adds the bicycle twice; its second cart assertion gets its expected points from
a **hardcoded pre-request script** (line 2617 of the collection), not from the data file.
Fix: [`fe27d64`](https://github.com/NyanSintZaw/sck-online-store/commit/fe27d64) (86 → 172).
Worth remembering: grep the collections too, not just `data/`.

## 3. The planted traps (say this to the instructor)

We found three deliberately planted bugs — all committed 2025-05-23 by the workshop host, one
with a literally confessing message.

**Trap 1 — Product 3 shows a negative price and 0 points.**
Commit `ddfdf90`: *"[Added] bug for product id 3 that show minus sign(-) for price and
point."* Three client-side lines negated `product_price_thb` after the API responded. The
0 Points was a second-order effect: the negated value reached the calculator, whose
`amount < 0` guard correctly returned 0 — **the guard was masking the bug, not causing it**.
The tell: the product *list* page showed the correct +฿897.45; only the detail page flipped.
Fix: [`6abf049`](https://github.com/NyanSintZaw/sck-online-store/commit/6abf049).

**Trap 2 — Product 7's detail page always 500s.**
`if ID == 7 { return error }` in the product *service* — but order creation calls the
*repository* directly, so the item still ordered fine while its page failed. The
inconsistency was the tell.
Fix: [`85bc2ef`](https://github.com/NyanSintZaw/sck-online-store/commit/85bc2ef).

**Trap 3 — Product 8's cart line never matches the subtotal.**
`+0.01` added to the displayed price only; the subtotal used the raw price (717.61 vs
717.60). Designed to break exact-value assertions.
Fix: [`6ea56f9`](https://github.com/NyanSintZaw/sck-online-store/commit/6ea56f9).

Also catalogued (data traps, deliberately **not** "fixed" — they are legitimate test inputs):
two zero-price products (ids 1044, 1339) and eight boundary products at 1.39/2.78 USD whose
THB value straddles the 50-baht point boundary between 2-dp and 6-dp rounding.

## 4. Red → green, per the board (steps ⑤–⑦)

The tests genuinely failed first, as the board intended (`FAILED = valid point == …`):
first run after rate 50 → TSS-AUTH-001 failed 27 assertions, TSS-AUTH-003 failed 2. We fixed
data, collection, and code until:

| Layer | Result |
|---|---|
| Newman (①①: 6 folders, ②: 2, ③: 2) | **10/10 folders, 423 assertions, 0 failures** |
| Robot (001 + 002) | **5 suites, 6 tests, 0 failures** |
| Go unit | 8 packages ok |
| Jest (point-service, rate-50 specs) | 17/17 |
| Cypress | 34/34 |

Verified in the browser: bicycle earns **86** (4,314.60 ÷ 50) on product page, cart, and PDF.

> Note for the instructor: the board's expected value was **80**, which assumes the
> 33.52 THB/USD example rate. The repo hardcodes 35.969964 (in
> [`currency.go`](https://github.com/NyanSintZaw/sck-online-store/blob/BC-1/store-service/internal/common/currency.go)),
> giving 86. Question: should the class standardize on 33.52 — and if so, should the FX rate
> come from an API as discussed? That decision shifts every expected value.

## 5. Scope: what BC-1 does NOT include

Only the first business condition is implemented. Still open:

| Condition | Missing |
|---|---|
| 2 points = 1.00 THB (Spending) | burn math is still 1:1 and the checkout Discount UI is commented out |
| 180-day validity | no date columns on the `points` table at all |
| Approved on confirm receipt | no such endpoint or UI exists; earned points are never credited to the ledger |
| Status table (4 statuses) | no status concept anywhere |

Known debt (deferred deliberately): no HTTP timeout / unclosed response bodies in the point
gateway, and a point-service outage currently 500s the whole cart. The cart uses 2-dp and the
order 6-dp for the same amount — visible with the 2.78 USD products.

Test gap worth closing next: `TSS-PC-001` doesn't test the requirement's own BVA triplet
(50.00 → 1, 45.00 → 0, 389.00 → 7). Three data rows to add.

## 6. For the team: about branch `Pai`

Please don't merge `Pai` as-is: it conflicts with main on the Robot teardown keyword, predates
the `Close All Pdfs` fix, its rate-50 point values are already superseded by BC-1, and its
UUID download folder never cleans up. Suggest rebasing it on BC-1 and keeping only the
UUID-directory idea, with cleanup restored in both PDF suites.

---

*Branch pushed as [`BC-1`](https://github.com/NyanSintZaw/sck-online-store/tree/BC-1) on the
NyanSintZaw fork (no push rights to the team repo — Stanley can add collaborators, or take
this via PR).*
