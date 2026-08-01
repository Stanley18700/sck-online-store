# BC-1 Walkthrough — Point Calculation via Point-Service

> Speaking script for the team + instructor demo. Branch `BC-1` (7 commits on top of team
> main `440ebc2`). File references are `path:line` — clickable in VS Code, valid on this
> branch. To see any commit mentioned: `git show <hash>`.

---

## 1. The instruction, and where we started

Our task: **point calculation must come from point-service, not store-service** (board steps
①②), write API tests in Postman and UI tests in Robot (③④), execute until 100% pass (⑤–⑦),
commit (⑧). The business condition under test is **BC1: "For every 50.00 THB spent on
purchased products, 1 reward point is earned."**

## 2. How the calculation is linked via point-service (the architecture)

The rule lives in exactly **one place** — point-service — and everything else calls it over
HTTP. Nothing computes points locally anymore.

```
 store-web                 store-service                    point-service
 (display only)            (orchestrator, no rule)          (OWNS the rule)
──────────────            ─────────────────────            ──────────────────
 product page ──┐
 cart line    ──┼─► GET /api/v1/point/calculate ──► gateway ──► GET :8001/api/v1/point/calculate
 checkout     ──┘         (JWT-protected)                        floor(amount / POINT_RATE)
                          cart & order flows also
                          call the same gateway
```

**The owner — point-service:**
- `point-service/src/point/point.constant.ts:1` — `POINT_RATE = 50`. The single source of
  truth for the rate.
- `point-service/src/point/point.service.ts:20-21` — the entire business rule:
  `amount < 0 ? 0 : Math.floor(amount / POINT_RATE)`.
- `point-service/src/point/point.controller.ts:23-24` — exposed as `GET /point/calculate`
  (global prefix makes it `/api/v1/point/calculate`); validates `amount`, returns
  `{ "point": n }`. Invalid input → 400 `amount must be a number`.

**The bridge — store-service (no calculation, only delegation):**
- `store-service/internal/point/gateway.go:44-45` — the HTTP client:
  builds `http://point-service:8001/api/v1/point/calculate?amount=…` and parses the JSON
  reply. This is the *only* wire between the services.
- `store-service/internal/point/point.go:63-64` — thin service wrapper around the gateway,
  so domain code depends on an interface (`point.go:13`), which is what the unit tests mock.
- Three consumers, all delegating:
  - `store-service/internal/cart/cart.go:49` — cart summary's `receive_point`
  - `store-service/internal/order/order.go:110` — the order's persisted `earn_point`
  - `store-service/cmd/api/point.go:87` + `store-service/cmd/main.go:255` — a passthrough
    endpoint so the frontend can ask for a preview (JWT-protected, inside the `protected`
    group)

**The consumer — store-web (displays, never computes):**
- `store-web/src/services/calculate-point.ts:13-17` — calls store-service's passthrough.
- `store-web/src/app/product/[id]/components/product-content.tsx` and the cart's
  `product-item.tsx` — render whatever the backend returns.

**The proof that store-service no longer calculates:** the old calculator files are *gone* —
`store-service/internal/common/point.go` (deleted in commit `11ec093`) and the frontend mirror
`store-web/src/utils/point.ts` + `config.pointRate` (deleted in `3f0a9b6`). A repo-wide grep
for `/ 50` or `Math.floor(amount` hits only point-service. Changing the rate touches one
constant, one service, one deploy.

> Why this matters (say this sentence): *"Before, the rate lived in three places — Go, the
> frontend, and the tests — and they drifted. Now the rule has one owner; store-service is a
> client of it, like any other microservice consumer."*

## 3. What we found when we audited it (for the team)

**Finding 1 — the rate had silently reverted to 100.**
The refactor branch was cut *from* the `100 → 50` commit (`8b53620`) but reintroduced the
rate as `POINT_RATE = 100` in the new constant file, then deleted every rate-50 source. No
commit message mentioned it, and since all test fixtures also said 100, nothing failed — the
revert was invisible. Lesson: when a rule moves between services, its *value* travels by hand.

Fix: commit `0e6f1c8` — the semantic change is **one line**
(`point-service/src/point/point.constant.ts:1`); the other 15 files are expectations catching
up (Jest specs, Newman data, Robot values, Go mocks).

**Finding 2 — main's Docker build was broken.**
The merge resurrected `store-web/src/__test__/point.cy.ts`, which imports the deleted
`utils/point.ts`. `npm run dev` never compiles orphaned test files, so nobody saw it locally —
but `next build` (Docker) fails. Fix: commit `fd37329` (file deleted).

**Finding 3 — one hidden expectation lives in the collection, not the data.**
`TSS-AUTH-003` adds the bicycle twice; its second cart assertion gets its expected points from
a **hardcoded pre-request script** — `atdd/api/collections/001-Authentication.postman_collection.json:2617`
— not from the data file. Fix: commit `fe27d64` (86 → 172). Lesson: grep the collections too,
not just `data/`.

## 4. The planted traps (for the instructor)

Three deliberately planted bugs, all committed 2025-05-23 by the workshop host — one with a
literally confessing message.

**Trap 1 — Product 3 shows a negative price and 0 points.**
Commit `ddfdf90`: *"[Added] bug for product id 3 that show minus sign(-) for price and
point."* Three client-side lines in `store-web/src/services/product-detail.ts` negated
`product_price_thb` after the API responded. The 0 Points was a second-order effect: the
negated value reached the calculator, whose `amount < 0` guard
(`point-service/src/point/point.service.ts:21`) correctly returned 0 — **the guard was
masking the bug, not causing it**. The tell: the product *list* page showed the correct
+฿897.45; only the detail page flipped. Fix: commit `6abf049`.

**Trap 2 — Product 7's detail page always 500s.**
`if ID == 7 { return error }` sat in the product *service*
(`store-service/internal/product/product.go`, now removed) — but order creation calls the
*repository* directly, so the item still ordered fine while its page failed. The
inconsistency was the tell. Fix: commit `85bc2ef`.

**Trap 3 — Product 8's cart line never matches the subtotal.**
`+0.01` was added to the displayed price only (`store-service/internal/cart/cart.go`, now
removed); the subtotal used the raw price — ฿717.61 vs ฿717.60. Designed to break exact-value
assertions. Fix: commit `6ea56f9`.

Also catalogued (data traps, deliberately **not** "fixed" — they are legitimate test inputs):
two zero-price products (ids 1044, 1339 in `tearup/store/init.sql`) and eight boundary
products at 1.39/2.78 USD whose THB value straddles the 50-baht point boundary between the
cart's 2-dp and the order's 6-dp rounding.

## 5. Red → green, per the board (steps ⑤–⑦)

The tests genuinely failed first, as the board intended (`FAILED = valid point == …`):
first run after rate 50 → TSS-AUTH-001 failed 27 assertions, TSS-AUTH-003 failed 2. We fixed
data, collection, and code until:

| Layer | Result |
|---|---|
| Newman (001: 6 folders · 002: 2 · 003: 2) | **10/10 folders, 423 assertions, 0 failures** |
| Robot (001 + 002) | **5 suites, 6 tests, 0 failures** |
| Go unit | 8 packages ok |
| Jest (point-service, rate-50 specs) | 17/17 |
| Cypress | 34/34 |

Verified in the browser: bicycle earns **86** (4,314.60 ÷ 50) on product page, cart, and PDF.

> Note for the instructor: the board's expected value was **80**, which assumes the
> 33.52 THB/USD example rate. The repo hardcodes 35.969964
> (`store-service/internal/common/currency.go:11`), giving 86. Question: should the class
> standardize on 33.52 — and if so, should the FX rate come from an API as discussed? That
> decision shifts every expected value.

## 6. Scope: what BC-1 does NOT include

Only the first business condition is implemented. Still open:

| Condition | Missing |
|---|---|
| 2 points = 1.00 THB (Spending) | burn math is still 1:1 (`store-web/src/utils/total-price.ts:18`) and the checkout Discount UI is commented out (`store-web/src/app/checkout/view.tsx:91`) |
| 180-day validity | no date columns on the `points` table (`point-service/src/point/point.entity.ts` — six columns only) |
| Approved on confirm receipt | no such endpoint or UI exists; earned points are never credited to the ledger (the only ledger write is the burn path, `store-service/internal/point/point.go:53`) |
| Status table (4 statuses) | no status concept anywhere in point-service |

Known debt (deferred deliberately): no HTTP timeout and no `Body.Close()` in
`store-service/internal/point/gateway.go`; a point-service outage currently 500s the whole
cart (`store-service/internal/cart/cart.go:49-53`). The cart uses 2-dp
(`cart.go:44`) and the order 6-dp (`order.go:86-88`) for the same amount — visible with the
2.78 USD products.

Test gap worth closing next: `atdd/api/data/003-Point-Calculate/TSS-PC-001.json` doesn't test
the requirement's own BVA triplet (50.00 → 1, 45.00 → 0, 389.00 → 7). Three data rows to add.

## 7. For the team: about branch `Pai`

Please don't merge `Pai` as-is: it conflicts with main on the Robot teardown keyword, predates
the `Close All Pdfs` fix, its rate-50 point values are already superseded by BC-1, and its
UUID download folder never cleans up. Suggest rebasing it on BC-1 and keeping only the
UUID-directory idea, with cleanup restored in both PDF suites.

---

*Branch `BC-1` is on the NyanSintZaw fork (no push rights to the team repo — Stanley can add
collaborators, or take this via PR).*
