# Onboarding — Machine Setup

> ฉบับภาษาไทย: [`instruction.th.md`](instruction.th.md)

Everything you need to install to build, run, and test **sck-online-store** on Windows, macOS,
or Linux. Follow the sections in order; the whole thing takes about 30–45 minutes, most of it
downloads.

---

## 1. Required versions

These are the versions the project targets. Install exactly these unless a section below says
otherwise.

| Tool            | Version    | Used for                                             |
| --------------- | ---------- | ---------------------------------------------------- |
| Go              | `1.26.5`   | `store-service` (backend API)                        |
| Node.js         | `24.18.1`  | `store-web`, `point-service`, `thirdparty`, Newman   |
| Robot Framework | `7.4.2`    | ATDD UI tests (`atdd/ui`)                            |
| Docker          | `29.6.2`   | All services, DB, test fixtures                      |

You also need these, which have no pinned version — install whatever your OS provides:

| Tool                | Why                                                                    |
| ------------------- | ---------------------------------------------------------------------- |
| Git                 | Cloning, committing                                                    |
| GNU Make            | Every workflow in this repo is a `make` target                          |
| Python 3.9+ and pip | Robot Framework runs inside a `venv` created by the Makefile            |
| Chrome              | Cypress component tests and Robot UI tests drive a real browser         |

Optional, only if you deploy to EKS: `kubectl` and the AWS CLI. See the "Build & Deploy to
EKS" section of [`CLAUDE.md`](../CLAUDE.md).

---

## 2. Windows: pick your path

You have two options. Both work — pick based on how comfortable you are with Linux.

| | **Path A — WSL2** | **Path B — plain Windows** |
| --- | --- | --- |
| You run commands in | Ubuntu terminal | Command Prompt (`cmd`) |
| `make` targets | All of them work | Not used — you run the commands directly |
| Learning curve | Some Linux basics | None beyond `cd` and `copy` |
| UI tests | Need Selenium Grid (`start_test_suite_grid`) | Your normal Chrome works |
| Follow | §2 below, then the **Linux** instructions everywhere | **[Appendix A](#appendix-a--plain-windows-command-prompt-only)** |

**Path A is recommended** if you're willing to learn a little WSL: you get the exact same
commands as everyone else on the team, so every instruction in this repo, in `CLAUDE.md`, and
in CI applies to you verbatim. Path B is fully supported and nothing is off-limits — you just
type the real commands instead of the `make` shortcut that wraps them.

**If you chose Path B, jump to [Appendix A](#appendix-a--plain-windows-command-prompt-only)
now** and skip §3–§7.

### Path A setup

**Do the whole setup inside WSL2 (Ubuntu), not in PowerShell or CMD.** The Makefile is written
for a POSIX shell — it uses `cp -f`, `python3 -m venv`, `sleep`, and `source`. None of that
exists in cmd, so `make` targets fail in confusing ways on native Windows.

```powershell
# In an Administrator PowerShell, once:
wsl --install -d Ubuntu
# Reboot, then open the "Ubuntu" terminal and do everything else there.
```

Then:

1. Install **Docker Desktop for Windows** on the Windows side and enable
   *Settings → Resources → WSL Integration* for your Ubuntu distro. Inside WSL, `docker` will
   then talk to Docker Desktop — do **not** `apt install docker.io` inside WSL as well.
2. Clone the repo into the **Linux** filesystem (`~/workspaces/…`), not `/mnt/c/…`. Building
   Go and Node across the Windows/Linux filesystem boundary is dramatically slower and breaks
   file watching.
3. From here on, follow the **Linux** instructions in every section below.

For the browser-based tests, Chrome installed on Windows is not reachable from WSL — use
`make start_test_suite_grid`, which runs Chrome inside a Selenium Grid container instead
(see §7).

---

## 3. Install the toolchain

> §3–§7 are for **macOS, Linux, and Windows Path A (WSL2)**. On Path B (plain Windows), use
> [Appendix A](#appendix-a--plain-windows-command-prompt-only) instead — but §8 (container
> versions) and §9 (troubleshooting) still apply to you.

### Go 1.26.5

<details>
<summary><b>macOS</b></summary>

```bash
brew install go            # then verify the version; if brew is behind, use the official pkg:
# https://go.dev/dl/ → go1.26.5.darwin-arm64.pkg (Apple Silicon) or -amd64.pkg (Intel)
```
</details>

<details>
<summary><b>Linux / WSL2</b></summary>

```bash
curl -LO https://go.dev/dl/go1.26.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc && source ~/.bashrc
```
</details>

Add `$HOME/go/bin` (macOS: `$(go env GOPATH)/bin`) to your `PATH` — the Go CLI tools in §4
install there.

### Node.js 24.18.1

Use a version manager; the exact version matters and you will hit other projects on other
versions.

```bash
# nvm — https://github.com/nvm-sh/nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
# restart your shell, then:
nvm install 24.18.1
nvm use 24.18.1
nvm alias default 24.18.1
```

### Docker 29.6.2

- **macOS / Windows** — Docker Desktop, from https://www.docker.com/products/docker-desktop/.
  Give it at least **8 GB RAM** in *Settings → Resources*; the full stack plus Selenium Grid
  is heavy.
- **Linux** — Docker Engine + the Compose plugin, per
  https://docs.docker.com/engine/install/. Then add yourself to the `docker` group so you
  don't need `sudo`:
  ```bash
  sudo usermod -aG docker $USER   # log out and back in
  ```

Compose **v2** is required — every target calls `docker compose` (a subcommand), never
`docker-compose` (the old standalone binary).

### Python + Robot Framework 7.4.2

The Robot targets build their own virtualenv at `atdd/ui/.venv` from
[`atdd/ui/requirements.txt`](../atdd/ui/requirements.txt) — which already pins
`robotframework==7.4.2` — so **you do not need a global Robot Framework install**. You only
need Python itself with the `venv` module:

```bash
# macOS
brew install python@3.12
# Ubuntu / WSL2
sudo apt update && sudo apt install -y python3 python3-pip python3-venv
```

If you want to run `robot` by hand outside the Makefile, install it into your own venv:

```bash
python3 -m venv ~/.venvs/sck && source ~/.venvs/sck/bin/activate
pip install -r atdd/ui/requirements.txt   # gives you robotframework 7.4.2 + SeleniumLibrary
```

---

## 4. Extra CLI tools

Two Makefile targets shell out to tools that are not installed by the steps above. Install
them now or those targets will fail with "command not found".

```bash
# Newman + the HTML reporter — needed by make run_newman (API tests)
npm install -g newman newman-reporter-htmlextra

# JUnit report converter — needed by make backend_unit_test
go install github.com/jstemmer/go-junit-report/v2@latest

# Swagger generator — only needed if you run make gen-swagger
go install github.com/swaggo/swag/cmd/swag@latest
```

---

## 5. Verify

```bash
go version                 # go version go1.26.5 …
node --version             # v24.18.1
npm --version
docker --version           # Docker version 29.6.2, …
docker compose version     # Docker Compose version v2.x
python3 --version          # 3.9 or newer
make --version
git --version
newman --version
go-junit-report -version
```

Every line must print a version. If `docker compose version` fails but `docker-compose
--version` works, you have Compose v1 — upgrade before continuing.

---

## 6. First run

```bash
git clone git@github.com:sck-shr-wlb/sck-online-store.git
cd sck-online-store

# Frontend env file — it is gitignored, so a fresh clone has none
cp store-web/.env_local store-web/.env

# Install workspace dependencies
make install_dependency_frontend      # store-web:      npm install
make install_dependency_backend       # store-service:  go mod tidy
cd point-service && npm install && cd ..   # no make target for this one; needed by make unit_test_all

# Build and start everything
make start_all
```

First `make start_all` pulls images and compiles both services — expect 5–15 minutes. Later
runs are much faster. When it settles:

| URL                             | What                                        |
| ------------------------------- | ------------------------------------------- |
| http://localhost                | The app, through nginx — **start here**     |
| http://localhost:3000           | store-web directly (Next.js)                |
| http://localhost:8000           | store-service API (Go)                      |
| http://localhost:8000/swagger/index.html | API docs                           |
| http://localhost:8001           | point-service (NestJS)                      |
| http://localhost:8080           | Adminer — DB browser (`user` / `password`)  |
| http://localhost:8882 / :8883   | Mock bank / shipping gateways               |

Stop everything with `make down`.

### Running the backend outside Docker

For a fast edit-compile loop on the Go service:

```bash
make store_db                 # MySQL only
make store_service_dev_mode   # runs store-service on your machine with dev env vars
```

---

## 7. Run the tests

```bash
# Unit tests — no Docker needed for the Go and Jest ones
make unit_test_all

# Integration tests — starts DB + thirdparty, runs the //go:build integration tests, tears down
make backend_integration_test

# Full ATDD suite
make start_test_suite     # local Chrome
make run_newman           # API tests
make run_robot            # UI tests
make stop_test_suite
```

On **WSL2**, or anywhere without a local Chrome, swap the first line for the Grid variant so
the browser runs in a container:

```bash
make start_test_suite_grid
REMOTE_HUB_URL=http://localhost:4444/wd/hub make run_robot
make stop_test_suite
```

> **Gotcha:** `start_test_suite` copies `store-web/.env_local` over `store-web/.env`, and
> `start_test_suite_grid` copies `.env_grid`. If you switch between local and Grid without
> re-running the matching target, the stale `.env` points the frontend at the wrong host and
> the UI tests fail for no visible reason.

Before pushing, `make test_all` runs the whole pipeline: lint → unit → ATDD.

---

## 8. Versions in the containers vs. on your machine

The containers now run the same versions §1 told you to install, so nothing here should look
like a bug. Every image tag is pinned to an exact patch — the table is the contract:

| Where                          | Version there              | You have     |
| ------------------------------ | -------------------------- | ------------ |
| `store-service/go.mod`         | `go 1.26.5` language level | Go 1.26.5    |
| `store-service/Dockerfile`     | `golang:1.26.5-alpine3.24` | Go 1.26.5    |
| `store-service` runtime stage  | `alpine:3.24`              | —            |
| `store-web/Dockerfile`         | `node:24.18.0-alpine3.24`  | Node 24.18.1 |
| `point-service/Dockerfile`     | `node:24.18.0-alpine3.24`  | Node 24.18.1 |

The one gap is the Node patch: **there is no `node:24.18.1` image on Docker Hub** — the newest
published patch on that line is `24.18.0` — so the containers sit one patch behind your local
`24.18.1`. That difference is invisible in practice. Don't "fix" it by floating the tag to
`node:24-alpine3.24`; a pinned patch is what keeps builds reproducible. When Docker Hub
publishes `24.18.1`, bump both Dockerfiles.

Go matches exactly, which means the two sides have to move together. Don't bump the `go`
directive in `go.mod` on its own when you install a newer Go — the Dockerfile pins the same
version, and a module that requires a newer toolchain than the builder image has will fail the
Docker build. Same the other way round: bumping the image without the `go` directive leaves the
docs and the build disagreeing about what the project targets.

The application frameworks are older than the runtimes underneath them — `store-web` is on
Next.js 14.0.4 and `point-service` on NestJS 9, both of which predate Node 24. They build and
run in practice, but if you hit an odd `next dev` / `next build` or Nest startup failure that
nobody else sees, drop to an older Node to confirm it's a runtime issue before chasing it:
`nvm install 20 && nvm use 20`.

---

## 9. Troubleshooting

**A port is already in use.** The stack claims 80, 3000, 3306, 8000, 8001, 8080, 8882, 8883,
and — with Grid — 4442/4443/4444. Port 80 and 3306 are the usual collisions (a local nginx,
Apache, or MySQL). Find the offender and stop it:

```bash
lsof -i :80          # macOS / Linux
sudo ss -lptn 'sport = :80'   # Linux alternative
```
```bat
netstat -ano | findstr :80    :: Windows — last column is the PID
tasklist /fi "pid eq 1234"    :: what that PID is
```

**Containers start then die.** Look at the logs before anything else:

```bash
docker compose logs -f store-service
docker compose ps            # which ones are unhealthy
```

**Database looks wrong or migrations didn't apply.** Wipe the volumes and rebuild. This
destroys local data, which is fine — everything is seeded:

```bash
make down
docker compose down -v
make start_test_suite
```

**Go tests pass locally but fail in CI, or vice versa.** The test cache is probably stale
after a fixture change: `make backend_clear_test_cache`.

**`make run_robot` fails installing Python packages.** Delete the venv and let it rebuild:
`rm -rf atdd/ui/.venv`.

**Apple Silicon: `exec format error` on an image.** Something was built for the wrong
architecture. For EKS images always use the `make eks_*` targets — they pass
`--platform linux/amd64` for you.

---

## Appendix A — Plain Windows, Command Prompt only

For **Path B**: no WSL, no Linux, no `make`. Everything below is typed into **Command Prompt**
(press `Win`, type `cmd`, hit Enter). There is nothing you can't do on this path — `make` is
only a shortcut, and this appendix gives you the real command behind each shortcut.

Two conventions in this appendix:

- Windows uses backslashes in paths: `store-web\.env`, not `store-web/.env`.
- Lines starting with `::` are comments. Don't type them.

### A.1 Install the tools

The fastest route is **winget**, which ships with Windows 10/11. Open Command Prompt **as
Administrator** (right-click → *Run as administrator*) and run:

```bat
winget install Git.Git
winget install GoLang.Go
winget install Python.Python.3.12
winget install CoreyButler.NVMforWindows
winget install Docker.DockerDesktop
```

If an ID isn't found, winget package names change occasionally — run `winget search go`,
`winget search docker`, and so on, or just download the installer from the tool's website.

**Close and reopen Command Prompt after installing.** Installers add to `PATH`, and an
already-open window keeps the old one. This is the single most common cause of "but I just
installed it" errors.

Then pin Node to the version this project uses:

```bat
nvm install 24.18.1
nvm use 24.18.1
```

> `nvm use` on Windows needs an Administrator prompt — it swaps a symlink. Run it once as
> Administrator and it sticks for normal windows too.

Then install the extra CLI tools from §4:

```bat
npm install -g newman newman-reporter-htmlextra
go install github.com/jstemmer/go-junit-report/v2@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

If `go-junit-report` isn't found afterwards, add Go's bin folder to your PATH:
*Start → "Edit the system environment variables" → Environment Variables → Path → New →*
`%USERPROFILE%\go\bin`. Reopen Command Prompt.

### A.2 About Docker Desktop and WSL

Docker Desktop uses WSL2 internally as its engine, and its installer sets that up for you.
**You never have to open or learn WSL** — you keep using Command Prompt, and `docker`
commands work there normally. If the installer asks to enable WSL2 or virtualization, say yes
and reboot.

Give Docker at least **8 GB RAM** in *Docker Desktop → Settings → Resources*, and make sure
the whale icon in the system tray says "Docker Desktop is running" before any `docker`
command.

### A.3 Verify

```bat
go version
node --version
npm --version
python --version
docker --version
docker compose version
git --version
newman --version
```

Note it's `python` on Windows, not `python3`. If Windows opens the Microsoft Store when you
type `python`, turn off the Store alias: *Settings → Apps → Advanced app settings → App
execution aliases* → switch off both `python.exe` entries.

### A.4 First run

```bat
git clone git@github.com:sck-shr-wlb/sck-online-store.git
cd sck-online-store

:: Frontend env file — gitignored, so a fresh clone has none
copy /Y store-web\.env_local store-web\.env

:: Install dependencies
cd store-web && npm install && cd ..
cd point-service && npm install && cd ..
cd store-service && go mod tidy && cd ..

:: Build and start everything
cp -f store-web/.env_local store-web/.env
docker compose up -d thirdparty point-service db store-service store-web nginx liquibase --build
```

That last line is exactly what `make start_test_suite` runs. Open http://localhost when it finishes —
see the URL table in §6 for the rest.

To stop everything:

```bat
docker compose down
```

### A.5 The command behind each `make` target

When a teammate, `CLAUDE.md`, or a README tells you to run `make <something>`, look it up
here. Run these from the repo root unless noted.

| Instead of | Run this in cmd |
| --- | --- |
| `make start_all` | `docker compose up -d db adminer liquibase thirdparty point-service store-service store-web nginx --build` |
| `make down` | `docker compose down` |
| `make store_db` | `docker compose up -d db` |
| `make install_dependency_frontend` | `cd store-web && npm install && cd ..` |
| `make install_dependency_backend` | `cd store-service && go mod tidy && cd ..` |
| `make code_analysis_frontend` | `cd store-web && npm run lint && cd ..` |
| `make code_analysis_backend` | `cd store-service && go vet ./... && cd ..` |
| `make backend_unit_test` | `cd store-service && go test -v ./... && cd ..` |
| `make code-coverage` | `cd store-service && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out && cd ..` |
| `make backend_clear_test_cache` | `cd store-service && go clean --testcache && cd ..` |
| `make build_backend` | `docker compose build store-service` |
| `make build_frontend` | `docker compose build store-web` |
| `make build_nginx` | `docker compose build nginx` |
| `make gen-swagger` | `cd store-service && swag init -g cmd/main.go -o cmd/docs && cd ..` |

The remaining targets need more than one line — they get their own sections below.

> **Do not use the `make eks_*` targets' commands by hand.** They generate a date-stamped
> image tag and must not be reproduced approximately. If you need to deploy to EKS, do it
> from a machine with `make`, or ask the team.

### A.6 Running the Go backend outside Docker

`make store_service_dev_mode` just sets environment variables and runs the service. In cmd:

```bat
docker compose up -d db

cd store-service\cmd
set DB_CONNECTION=user:password@tcp(localhost:3306)/store?parseTime=true
set POINT_GATEWAY=localhost:8001
set BANK_GATEWAY=localhost:8882
set SHIPPING_GATEWAY=localhost:8883
set JWT_SECRET=my-secret-key
go run main.go
```

> **Don't put quotes around the values.** In cmd, `set X="abc"` stores the quote characters as
> part of the value, and the service will fail to connect with a confusing error.

These variables live only in that Command Prompt window. Open a new window and you must set
them again — so keep this window open while you develop, and stop the service with `Ctrl+C`.

### A.7 Unit tests

```bat
:: Go
cd store-service && go test -v ./... && cd ..

:: point-service (Jest)
cd point-service && npm test && cd ..

:: store-web component tests (Cypress)
cd store-web && npm run test:component && cd ..
```

Together those three are `make unit_test_all`.

To run one Go package only, or one Cypress spec:

```bat
cd store-service && go test -v ./internal/order/... && cd ..
cd store-web && npx cypress run --component --spec "src/components/cart.cy.tsx" && cd ..
```

### A.8 Integration tests

The Go integration tests are hidden behind the `integration` build tag and need MySQL,
Liquibase, and the mock gateways running first. `make backend_integration_test` in cmd:

```bat
docker compose up -d db thirdparty
timeout /t 15
docker compose up liquibase

cd store-service && go test -tags=integration ./... && cd ..

docker compose down
```

(`timeout /t 15` is the Windows version of `sleep` — it waits for MySQL to accept
connections. If Liquibase fails, wait longer and rerun the `docker compose up liquibase`
line.)

### A.9 API tests (Newman)

```bat
:: Start the stack first
copy /Y store-web\.env_local store-web\.env
docker compose up -d thirdparty point-service db store-service store-web nginx liquibase --build

cd atdd\api

:: Authentication suite — all six cases
for %F in (TSS-AUTH-001 TSS-AUTH-002 TSS-AUTH-003 TSA-AUTH-001 TSA-AUTH-002 TSA-AUTH-003) do newman run collections\001-Authentication.postman_collection.json --folder "%F" -e sck-online-store.local.postman_environment.json -d data\001-Authentication\%F.json -r cli,junit,htmlextra

:: Order Summary PDF suite
for %F in (TSS-OSP-001 TSS-OSP-002) do newman run collections\002-Order-Summary-PDF.postman_collection.json --folder "%F" -e sck-online-store.local.postman_environment.json -d data\002-Order-Summary-PDF\%F.json -r cli,junit,htmlextra

cd ..\..
docker compose down
```

> If you save those `for` lines into a `.bat` file instead of typing them, double the percent
> signs: `%%F` instead of `%F`. That's a cmd quirk, not a typo.

Reports land in `atdd\api\newman\`.

### A.10 UI tests (Robot Framework)

This is where plain Windows is actually *easier* than WSL — your normal Chrome works, so you
can watch the tests run.

```bat
:: Start the stack (same as A.9)
copy /Y store-web\.env_local store-web\.env
docker compose up -d thirdparty point-service db store-service store-web nginx liquibase --build

cd atdd\ui
python -m venv .venv
.venv\Scripts\activate
pip install -r requirements.txt

robot -v URL:http://localhost/product/list -x reports\authen.xml 001-Authentication
robot -v URL:http://localhost/product/list -x reports\pdf.xml 002-Order-Summary-PDF

deactivate
cd ..\..
docker compose down
```

Notes:

- Activation on Windows is `.venv\Scripts\activate` — **not** `source .venv/bin/activate`,
  which is what the Makefile uses and what you'll see in Linux instructions.
- You only need the `python -m venv` and `pip install` lines the first time. After that,
  `.venv\Scripts\activate` then `robot …` is enough.
- The Makefile passes `-v REMOTE_HUB_URL:…`; you can omit it. It defaults to empty, which
  means "drive the local browser" — exactly what you want here.
- Results: `log.html` and `report.html` in `atdd\ui`. Open them in a browser.

### A.11 Do I need GNU Make?

You can install it (`winget install GnuWin32.Make`), and the single-line targets in the table
in §A.5 will then work. But these targets will still fail in cmd, because their recipes use
Linux-only commands:

| Target | Why it fails on native Windows |
| --- | --- |
| `setup_test_fixtures`, `backend_integration_test` | uses `sleep` |
| `start_test_suite`, `start_test_suite_grid` | uses `cp -f` |
| `run_robot*` | uses `python3` and `source .venv/bin/activate` |
| `store_service_dev_mode` | uses POSIX `VAR=value cmd` env syntax |
| `eks_*` | uses shell `$(date …)` substitution |

Since that's most of the interesting ones, installing Make buys you little. The commands in
this appendix are the complete substitute — that's the recommendation.

### A.12 Windows-specific gotchas

**Use `cmd`, not PowerShell**, or at least know the differences: PowerShell 5.1 (the one
bundled with Windows) doesn't support `&&` between commands, and environment variables are
`$env:JWT_SECRET = "value"` rather than `set JWT_SECRET=value`. Every example in this appendix
assumes Command Prompt.

**Git line endings.** Set this once, before your first commit, so you don't turn every file
in a diff into a change:

```bat
git config --global core.autocrlf true
```

**Antivirus and Windows Defender** scanning `node_modules` makes `npm install` and Next.js
builds several times slower. If builds crawl, add your repo folder and
`%USERPROFILE%\go` to the Defender exclusion list.

**Long paths.** `node_modules` can exceed the legacy 260-character limit. If you see
"filename too long" during `npm install` or `git clone`:

```bat
git config --global core.longpaths true
```

**"docker: command not found"** almost always means Docker Desktop isn't running (check the
tray) or you opened Command Prompt before installing it (reopen the window).

---

## Where to go next

- [`CLAUDE.md`](../CLAUDE.md) — architecture, every make target, naming conventions
- [`README.md`](../README.md) — coding conventions (Thai)
- [`atdd/CLAUDE.md`](../atdd/CLAUDE.md) — how the API and UI test suites are organized
- [`thirdparty/CLAUDE.md`](../thirdparty/CLAUDE.md) — the mock payment/shipping gateways
- [`deploy/README.md`](../deploy/README.md) — Kubernetes and Terraform
