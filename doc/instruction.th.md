# คู่มือติดตั้งเครื่องสำหรับสมาชิกใหม่

> English version: [`instruction.md`](instruction.md)

ทุกอย่างที่ต้องติดตั้งเพื่อ build, run และ test **sck-online-store** บน Windows, macOS หรือ Linux
ทำตามหัวข้อเรียงลำดับ ใช้เวลาประมาณ 30–45 นาที ซึ่งส่วนใหญ่คือเวลาดาวน์โหลด

---

## 1. เวอร์ชันที่ต้องใช้

นี่คือเวอร์ชันที่โปรเจกต์นี้ใช้ ให้ติดตั้งตามนี้เป๊ะ ๆ ยกเว้นมีหัวข้อด้านล่างระบุไว้เป็นอย่างอื่น

| เครื่องมือ        | เวอร์ชัน   | ใช้กับ                                               |
| --------------- | ---------- | ---------------------------------------------------- |
| Go              | `1.26.5`   | `store-service` (backend API)                        |
| Node.js         | `24.18.1`  | `store-web`, `point-service`, `thirdparty`, Newman   |
| Robot Framework | `7.4.2`    | ATDD UI tests (`atdd/ui`)                            |
| Docker          | `29.6.2`   | ทุก service, ฐานข้อมูล, test fixtures                 |

นอกจากนี้ยังต้องมีอีก 4 อย่าง ซึ่งไม่ได้ล็อกเวอร์ชันไว้ ใช้ตัวที่ OS ของคุณมีให้ได้เลย

| เครื่องมือ            | ใช้ทำอะไร                                                          |
| ------------------- | ------------------------------------------------------------------ |
| Git                 | clone และ commit โค้ด                                              |
| GNU Make            | ทุก workflow ใน repo นี้เป็น `make` target                          |
| Python 3.9+ กับ pip | Robot Framework รันอยู่ใน `venv` ที่ Makefile สร้างให้               |
| Chrome              | Cypress component tests และ Robot UI tests ต้องใช้เบราว์เซอร์จริง    |

ไม่บังคับ ติดตั้งเฉพาะกรณีที่ต้อง deploy ขึ้น EKS: `kubectl` และ AWS CLI ดูรายละเอียดได้ที่หัวข้อ
"Build & Deploy to EKS" ใน [`CLAUDE.md`](../CLAUDE.md)

---

## 2. Windows: เลือกเส้นทางของคุณ

มี 2 ทางเลือก ใช้ได้ทั้งคู่ เลือกตามความคุ้นเคยกับ Linux ของคุณ

| | **ทาง A — WSL2** | **ทาง B — Windows ล้วน** |
| --- | --- | --- |
| พิมพ์คำสั่งที่ไหน | Ubuntu terminal | Command Prompt (`cmd`) |
| `make` target | ใช้ได้ทุกตัว | ไม่ได้ใช้ — พิมพ์คำสั่งจริงเอง |
| ต้องเรียนรู้เพิ่ม | พื้นฐาน Linux นิดหน่อย | แค่ `cd` กับ `copy` ก็พอ |
| UI tests | ต้องใช้ Selenium Grid (`start_test_suite_grid`) | ใช้ Chrome ปกติของเครื่องได้เลย |
| ให้อ่าน | §2 ด้านล่าง แล้วทำตามคำสั่งฝั่ง **Linux** ทั้งหมด | **[ภาคผนวก A](#ภาคผนวก-a)** |

**แนะนำทาง A** ถ้าคุณพอจะเรียนรู้ WSL ได้บ้าง เพราะคุณจะได้ใช้คำสั่งชุดเดียวกับคนอื่นในทีม
ทำให้ทุกคำสั่งใน repo นี้ ใน `CLAUDE.md` และใน CI ใช้ได้กับคุณตรง ๆ ส่วนทาง B ก็รองรับเต็มที่
ไม่มีอะไรที่ทำไม่ได้ เพียงแต่คุณต้องพิมพ์คำสั่งจริงแทนการเรียก `make` ที่เป็นแค่ตัวย่อเท่านั้น

**ถ้าเลือกทาง B ให้ข้ามไปที่ [ภาคผนวก A](#ภาคผนวก-a) เลย** และข้าม §3–§7 ไปได้

### การติดตั้งสำหรับทาง A

**ให้ติดตั้งทุกอย่างใน WSL2 (Ubuntu) ไม่ใช่ใน PowerShell หรือ CMD** เพราะ Makefile เขียนมาสำหรับ
POSIX shell โดยใช้ `cp -f`, `python3 -m venv`, `sleep` และ `source` ซึ่งไม่มีใน cmd
ทำให้ `make` target ต่าง ๆ พังแบบงง ๆ บน Windows ปกติ

```powershell
# เปิด PowerShell แบบ Administrator แล้วรันครั้งเดียว:
wsl --install -d Ubuntu
# รีสตาร์ตเครื่อง จากนั้นเปิด terminal "Ubuntu" แล้วทำทุกอย่างในนั้น
```

จากนั้น:

1. ติดตั้ง **Docker Desktop for Windows** ฝั่ง Windows แล้วเปิด
   *Settings → Resources → WSL Integration* ให้ distro Ubuntu ของคุณ เมื่อเข้าไปใน WSL คำสั่ง
   `docker` จะคุยกับ Docker Desktop ให้เอง — **อย่า** `apt install docker.io` ใน WSL ซ้ำอีก
2. Clone repo ลงใน filesystem ของ **Linux** (`~/workspaces/…`) ไม่ใช่ `/mnt/c/…` เพราะการ build
   Go และ Node ข้ามเส้นแบ่ง Windows/Linux ช้ากว่ามาก และทำให้ file watching พัง
3. จากจุดนี้ไป ให้ทำตามคำสั่งฝั่ง **Linux** ในทุกหัวข้อด้านล่าง

สำหรับเทสต์ที่ต้องใช้เบราว์เซอร์ Chrome ที่ติดตั้งบน Windows จะเรียกจาก WSL ไม่ได้ ให้ใช้
`make start_test_suite_grid` ซึ่งรัน Chrome ใน container ของ Selenium Grid แทน (ดู §7)

---

## 3. ติดตั้ง toolchain

> §3–§7 สำหรับ **macOS, Linux และ Windows ทาง A (WSL2)** ถ้าคุณใช้ทาง B (Windows ล้วน)
> ให้ไปที่ [ภาคผนวก A](#ภาคผนวก-a) แทน — แต่ §8 (เวอร์ชันใน container) และ §9 (แก้ปัญหา)
> ยังใช้กับคุณอยู่

### Go 1.26.5

<details>
<summary><b>macOS</b></summary>

```bash
brew install go            # แล้วเช็กเวอร์ชัน ถ้า brew ยังไม่อัปเดต ให้ใช้ตัวติดตั้งทางการ:
# https://go.dev/dl/ → go1.26.5.darwin-arm64.pkg (Apple Silicon) หรือ -amd64.pkg (Intel)
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

อย่าลืมเพิ่ม `$HOME/go/bin` (บน macOS คือ `$(go env GOPATH)/bin`) เข้า `PATH` ด้วย
เพราะเครื่องมือ Go CLI ใน §4 จะถูกติดตั้งไว้ที่นั่น

### Node.js 24.18.1

ใช้ version manager เถอะ เพราะเวอร์ชันสำคัญมาก และคุณจะต้องเจอโปรเจกต์อื่นที่ใช้เวอร์ชันอื่นแน่นอน

```bash
# nvm — https://github.com/nvm-sh/nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
# ปิดแล้วเปิด shell ใหม่ จากนั้น:
nvm install 24.18.1
nvm use 24.18.1
nvm alias default 24.18.1
```

### Docker 29.6.2

- **macOS / Windows** — ใช้ Docker Desktop จาก https://www.docker.com/products/docker-desktop/
  ตั้งค่า RAM ให้อย่างน้อย **8 GB** ที่ *Settings → Resources* เพราะรันครบทั้ง stack พร้อม
  Selenium Grid นั้นกินทรัพยากรพอสมควร
- **Linux** — ติดตั้ง Docker Engine พร้อม Compose plugin ตาม
  https://docs.docker.com/engine/install/ แล้วเพิ่มตัวเองเข้ากลุ่ม `docker` จะได้ไม่ต้องใช้ `sudo`:
  ```bash
  sudo usermod -aG docker $USER   # แล้ว logout แล้ว login ใหม่
  ```

ต้องเป็น Compose **v2** เท่านั้น เพราะทุก target เรียก `docker compose` (แบบ subcommand)
ไม่ใช่ `docker-compose` (ไบนารีตัวเก่า)

### Python + Robot Framework 7.4.2

Robot target จะสร้าง virtualenv ของตัวเองที่ `atdd/ui/.venv` จาก
[`atdd/ui/requirements.txt`](../atdd/ui/requirements.txt) ซึ่งล็อก `robotframework==7.4.2`
ไว้อยู่แล้ว **คุณจึงไม่จำเป็นต้องติดตั้ง Robot Framework แบบ global** แค่มี Python
พร้อมโมดูล `venv` ก็พอ

```bash
# macOS
brew install python@3.12
# Ubuntu / WSL2
sudo apt update && sudo apt install -y python3 python3-pip python3-venv
```

ถ้าอยากรัน `robot` เองโดยไม่ผ่าน Makefile ให้ติดตั้งลง venv ของคุณเอง:

```bash
python3 -m venv ~/.venvs/sck && source ~/.venvs/sck/bin/activate
pip install -r atdd/ui/requirements.txt   # ได้ robotframework 7.4.2 + SeleniumLibrary
```

---

## 4. เครื่องมือ CLI เพิ่มเติม

มี Makefile target บางตัวที่เรียกใช้เครื่องมือซึ่งขั้นตอนข้างบนไม่ได้ติดตั้งให้
ติดตั้งตอนนี้เลย ไม่งั้น target เหล่านั้นจะฟ้อง "command not found"

```bash
# Newman + HTML reporter — จำเป็นสำหรับ make run_newman (API tests)
npm install -g newman newman-reporter-htmlextra

# ตัวแปลงรายงานเป็น JUnit — จำเป็นสำหรับ make backend_unit_test
go install github.com/jstemmer/go-junit-report/v2@latest

# ตัวสร้าง Swagger — ใช้เฉพาะตอนรัน make gen-swagger
go install github.com/swaggo/swag/cmd/swag@latest
```

---

## 5. ตรวจสอบว่าติดตั้งครบ

```bash
go version                 # go version go1.26.5 …
node --version             # v24.18.1
npm --version
docker --version           # Docker version 29.6.2, …
docker compose version     # Docker Compose version v2.x
python3 --version          # 3.9 ขึ้นไป
make --version
git --version
newman --version
go-junit-report -version
```

ทุกบรรทัดต้องขึ้นเลขเวอร์ชัน ถ้า `docker compose version` ไม่ผ่าน แต่ `docker-compose --version`
ผ่าน แสดงว่าคุณยังใช้ Compose v1 อยู่ ให้อัปเกรดก่อนไปต่อ

---

## 6. รันครั้งแรก

```bash
git clone git@github.com:sck-shr-wlb/sck-online-store.git
cd sck-online-store

# ไฟล์ env ของ frontend — ถูก gitignore ไว้ เครื่องที่เพิ่ง clone มาจึงยังไม่มี
cp store-web/.env_local store-web/.env

# ติดตั้ง dependency ของแต่ละ service
make install_dependency_frontend      # store-web:      npm install
make install_dependency_backend       # store-service:  go mod tidy
cd point-service && npm install && cd ..   # ตัวนี้ไม่มี make target แต่ make unit_test_all ต้องใช้

# build แล้วสตาร์ตทุกอย่าง
make start_all
```

`make start_all` ครั้งแรกจะ pull image และ compile ทั้งสอง service ใช้เวลาประมาณ 5–15 นาที
ครั้งถัด ๆ ไปจะเร็วขึ้นมาก เมื่อทุกอย่างพร้อมแล้ว:

| URL                             | คืออะไร                                       |
| ------------------------------- | -------------------------------------------- |
| http://localhost                | ตัวแอป ผ่าน nginx — **เริ่มที่นี่**             |
| http://localhost:3000           | store-web ตรง ๆ (Next.js)                     |
| http://localhost:8000           | store-service API (Go)                        |
| http://localhost:8000/swagger/index.html | เอกสาร API                          |
| http://localhost:8001           | point-service (NestJS)                        |
| http://localhost:8080           | Adminer — เปิดดูฐานข้อมูล (`user` / `password`) |
| http://localhost:8882 / :8883   | Mock bank / shipping gateway                  |

ปิดทุกอย่างด้วย `make down`

### รัน backend นอก Docker

ถ้าอยากได้รอบ edit-compile ที่เร็วขึ้นตอนพัฒนา Go service:

```bash
make store_db                 # เปิดเฉพาะ MySQL
make store_service_dev_mode   # รัน store-service บนเครื่องคุณ พร้อม env สำหรับ dev
```

---

## 7. รันเทสต์

```bash
# Unit tests — ฝั่ง Go กับ Jest ไม่ต้องใช้ Docker
make unit_test_all

# Integration tests — เปิด DB + thirdparty, รันเทสต์ที่ติด //go:build integration แล้วปิดให้
make backend_integration_test

# ชุด ATDD เต็ม
make start_test_suite     # ใช้ Chrome บนเครื่อง
make run_newman           # API tests
make run_robot            # UI tests
make stop_test_suite
```

ถ้าอยู่บน **WSL2** หรือเครื่องที่ไม่มี Chrome ในตัว ให้เปลี่ยนบรรทัดแรกเป็นแบบ Grid
เพื่อให้เบราว์เซอร์รันใน container แทน:

```bash
make start_test_suite_grid
REMOTE_HUB_URL=http://localhost:4444/wd/hub make run_robot
make stop_test_suite
```

> **จุดที่พลาดกันบ่อย:** `start_test_suite` จะ copy `store-web/.env_local` ทับ `store-web/.env`
> ส่วน `start_test_suite_grid` จะ copy `.env_grid` ทับ ถ้าคุณสลับไปมาระหว่างแบบ local กับ Grid
> โดยไม่ได้รัน target ที่ตรงกันใหม่ ไฟล์ `.env` ที่ค้างอยู่จะชี้ frontend ไปผิด host
> แล้ว UI test จะ fail โดยไม่มีสาเหตุที่มองเห็นได้

ก่อน push ให้รัน `make test_all` ซึ่งจะไล่ทั้ง pipeline: lint → unit → ATDD

---

## 8. เวอร์ชันใน container เทียบกับบนเครื่องคุณ

ตอนนี้ container ใช้เวอร์ชันเดียวกับที่ §1 บอกให้ติดตั้งแล้ว ดังนั้นไม่ควรมีอะไรในนี้ที่ดูเหมือนบั๊ก
ทุก image tag ถูก pin ไว้ที่ระดับ patch แบบเจาะจง — ตารางนี้คือข้อตกลง:

| ที่ไหน                          | เวอร์ชันตรงนั้น              | ของคุณ       |
| ------------------------------ | -------------------------- | ------------ |
| `store-service/go.mod`         | ระดับภาษา `go 1.26.5`       | Go 1.26.5    |
| `store-service/Dockerfile`     | `golang:1.26.5-alpine3.24` | Go 1.26.5    |
| stage runtime ของ `store-service` | `alpine:3.24`           | —            |
| `store-web/Dockerfile`         | `node:24.18.0-alpine3.24`  | Node 24.18.1 |
| `point-service/Dockerfile`     | `node:24.18.0-alpine3.24`  | Node 24.18.1 |

จุดเดียวที่ไม่ตรงกันคือเลข patch ของ Node: **บน Docker Hub ไม่มี image `node:24.18.1`**
patch ล่าสุดที่เผยแพร่ในสายนี้คือ `24.18.0` ดังนั้น container จึงตามหลังเครื่องคุณ (`24.18.1`)
อยู่หนึ่ง patch ซึ่งในทางปฏิบัติไม่มีผลอะไร และ**อย่า**แก้ให้ตรงด้วยการเปลี่ยนไปใช้ tag แบบลอย
อย่าง `node:24-alpine3.24` เพราะการ pin patch ไว้คือสิ่งที่ทำให้ build ทำซ้ำได้เหมือนเดิมทุกครั้ง
เมื่อไหร่ที่ Docker Hub ปล่อย `24.18.1` ให้อัปเดต Dockerfile ทั้งสองไฟล์

ฝั่ง Go ตรงกันเป๊ะ ซึ่งหมายความว่าสองฝั่งต้องขยับไปพร้อมกัน **อย่าแก้ค่า `go` ใน `go.mod`
เพียงฝ่ายเดียว** เวลาคุณติดตั้ง Go เวอร์ชันใหม่ เพราะ Dockerfile pin เวอร์ชันเดียวกันไว้
ถ้าโมดูลต้องการ toolchain ที่ใหม่กว่า image ที่ใช้ build อยู่ Docker build จะพัง และในทางกลับกัน
ถ้าเปลี่ยน image โดยไม่แก้ค่า `go` เอกสารกับตัว build ก็จะไม่ตรงกันว่าโปรเจกต์เล็งเวอร์ชันไหน

framework ของแอปเก่ากว่า runtime ที่รองรับมันอยู่ — `store-web` ใช้ Next.js 14.0.4 และ
`point-service` ใช้ NestJS 9 ซึ่งทั้งคู่ออกมาก่อน Node 24 ในทางปฏิบัติมัน build และรันได้
แต่ถ้าคุณเจอ error แปลก ๆ ตอน `next dev` / `next build` หรือตอน Nest เริ่มทำงานที่ไม่มีใครเจอ
ให้ลองลดไปใช้ Node เวอร์ชันเก่ากว่าเพื่อยืนยันว่าเป็นปัญหาของ runtime ก่อนไล่หาสาเหตุ:
`nvm install 20 && nvm use 20`

---

## 9. แก้ปัญหาที่พบบ่อย

**พอร์ตถูกใช้อยู่แล้ว** stack นี้ใช้พอร์ต 80, 3000, 3306, 8000, 8001, 8080, 8882, 8883
และถ้าใช้ Grid ก็เพิ่ม 4442/4443/4444 ตัวที่ชนบ่อยที่สุดคือ 80 กับ 3306 (จาก nginx, Apache
หรือ MySQL ที่ติดตั้งไว้บนเครื่อง) หาตัวที่ยึดพอร์ตอยู่แล้วปิดมัน:

```bash
lsof -i :80          # macOS / Linux
sudo ss -lptn 'sport = :80'   # อีกวิธีสำหรับ Linux
```
```bat
netstat -ano | findstr :80    :: Windows — คอลัมน์สุดท้ายคือ PID
tasklist /fi "pid eq 1234"    :: ดูว่า PID นั้นคือโปรแกรมอะไร
```

**Container สตาร์ตแล้วดับ** ให้ดู log ก่อนเป็นอันดับแรก:

```bash
docker compose logs -f store-service
docker compose ps            # ดูว่าตัวไหน unhealthy
```

**ฐานข้อมูลผิดเพี้ยน หรือ migration ไม่ทำงาน** ให้ลบ volume แล้ว build ใหม่
ข้อมูลบนเครื่องจะหายซึ่งไม่เป็นไร เพราะมี seed data ให้อยู่แล้ว:

```bash
make down
docker compose down -v
make start_all
```

**เทสต์ Go ผ่านบนเครื่องแต่ fail ใน CI หรือกลับกัน** มักเกิดจาก test cache ค้างหลังแก้ fixture:
`make backend_clear_test_cache`

**`make run_robot` ติดตั้ง Python package ไม่ผ่าน** ลบ venv ทิ้งแล้วให้มันสร้างใหม่:
`rm -rf atdd/ui/.venv`

**Apple Silicon: เจอ `exec format error` ตอนรัน image** แปลว่า image ถูก build มาผิด
สถาปัตยกรรม สำหรับ image ที่จะขึ้น EKS ให้ใช้ `make eks_*` เสมอ เพราะมันใส่
`--platform linux/amd64` ให้อยู่แล้ว

---

<a id="ภาคผนวก-a"></a>

## ภาคผนวก A — Windows ล้วน ใช้แค่ Command Prompt

สำหรับ **ทาง B**: ไม่ใช้ WSL ไม่ต้องรู้ Linux ไม่ต้องมี `make` ทุกอย่างด้านล่างพิมพ์ใน
**Command Prompt** (กด `Win` พิมพ์ `cmd` แล้ว Enter) ไม่มีอะไรที่ทางนี้ทำไม่ได้ เพราะ `make`
เป็นแค่ตัวย่อ และภาคผนวกนี้ให้คำสั่งจริงที่อยู่เบื้องหลังตัวย่อแต่ละตัว

ข้อตกลง 2 ข้อในภาคผนวกนี้:

- Windows ใช้ backslash ในพาธ: `store-web\.env` ไม่ใช่ `store-web/.env`
- บรรทัดที่ขึ้นต้นด้วย `::` คือคอมเมนต์ ไม่ต้องพิมพ์ตาม

### A.1 ติดตั้งเครื่องมือ

วิธีที่เร็วที่สุดคือใช้ **winget** ซึ่งมีมาให้อยู่แล้วใน Windows 10/11 เปิด Command Prompt
**แบบ Administrator** (คลิกขวา → *Run as administrator*) แล้วรัน:

```bat
winget install Git.Git
winget install GoLang.Go
winget install Python.Python.3.12
winget install CoreyButler.NVMforWindows
winget install Docker.DockerDesktop
```

ถ้าหา ID ไหนไม่เจอ เพราะชื่อ package ของ winget เปลี่ยนเป็นครั้งคราว ให้ลอง `winget search go`,
`winget search docker` หรือดาวน์โหลดตัวติดตั้งจากเว็บของเครื่องมือนั้นโดยตรงก็ได้

**ปิดแล้วเปิด Command Prompt ใหม่หลังติดตั้งเสร็จ** เพราะตัวติดตั้งจะเพิ่มค่าใน `PATH`
แต่หน้าต่างที่เปิดค้างไว้ยังใช้ค่าเดิม นี่คือสาเหตุอันดับหนึ่งของอาการ "ก็เพิ่งลงไปนี่นา"

จากนั้นล็อกเวอร์ชัน Node ให้ตรงกับที่โปรเจกต์ใช้:

```bat
nvm install 24.18.1
nvm use 24.18.1
```

> `nvm use` บน Windows ต้องรันจากหน้าต่างแบบ Administrator เพราะมันสลับ symlink
> รันแบบ Administrator ครั้งเดียว แล้วหน้าต่างธรรมดาก็จะใช้เวอร์ชันนั้นตามไปด้วย

แล้วติดตั้งเครื่องมือ CLI เพิ่มเติมจาก §4:

```bat
npm install -g newman newman-reporter-htmlextra
go install github.com/jstemmer/go-junit-report/v2@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

ถ้าหลังจากนั้นเรียก `go-junit-report` ไม่เจอ ให้เพิ่มโฟลเดอร์ bin ของ Go เข้า PATH:
*Start → "Edit the system environment variables" → Environment Variables → Path → New →*
`%USERPROFILE%\go\bin` แล้วเปิด Command Prompt ใหม่

### A.2 เรื่อง Docker Desktop กับ WSL

Docker Desktop ใช้ WSL2 เป็น engine อยู่ภายใน และตัวติดตั้งจะจัดการให้เองทั้งหมด
**คุณไม่จำเป็นต้องเปิดหรือเรียนรู้ WSL เลย** — คุณยังใช้ Command Prompt ตามปกติ และคำสั่ง
`docker` ก็ทำงานได้ตรงนั้น ถ้าตัวติดตั้งถามว่าจะเปิดใช้ WSL2 หรือ virtualization ให้ตอบตกลง
แล้วรีสตาร์ตเครื่อง

ตั้ง RAM ให้ Docker อย่างน้อย **8 GB** ที่ *Docker Desktop → Settings → Resources*
และก่อนรันคำสั่ง `docker` ทุกครั้ง ให้ดูว่าไอคอนวาฬใน system tray ขึ้นว่า
"Docker Desktop is running" แล้ว

### A.3 ตรวจสอบว่าติดตั้งครบ

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

สังเกตว่าบน Windows ใช้ `python` ไม่ใช่ `python3` ถ้าพิมพ์ `python` แล้ว Windows เด้ง
Microsoft Store ขึ้นมา ให้ปิด alias ทิ้ง: *Settings → Apps → Advanced app settings →
App execution aliases* แล้วปิดสวิตช์ `python.exe` ทั้งสองอัน

### A.4 รันครั้งแรก

```bat
git clone git@github.com:sck-shr-wlb/sck-online-store.git
cd sck-online-store

:: ไฟล์ env ของ frontend — ถูก gitignore ไว้ เครื่องที่เพิ่ง clone จึงยังไม่มี
copy /Y store-web\.env_local store-web\.env

:: ติดตั้ง dependency
cd store-web && npm install && cd ..
cd point-service && npm install && cd ..
cd store-service && go mod tidy && cd ..

:: build แล้วสตาร์ตทุกอย่าง
cp -f store-web/.env_local store-web/.env
docker compose up -d thirdparty point-service db store-service store-web nginx liquibase --build
```

บรรทัดสุดท้ายคือสิ่งที่ `make start_all` รันจริง ๆ เมื่อเสร็จแล้วให้เปิด http://localhost
ส่วน URL อื่น ๆ ดูได้จากตารางใน §6

ปิดทุกอย่างด้วย:

```bat
docker compose down
```

### A.5 คำสั่งจริงที่อยู่เบื้องหลัง `make` แต่ละตัว

เวลาเพื่อนร่วมทีม, `CLAUDE.md` หรือ README บอกให้รัน `make <อะไรสักอย่าง>` ให้เปิดตารางนี้ดู
รันจาก root ของ repo ยกเว้นระบุไว้เป็นอย่างอื่น

| แทนที่จะรัน | ให้พิมพ์ใน cmd |
| --- | --- |
| `make start_all` | `docker compose up -d db adminer seed liquibase thirdparty point-service store-service store-web nginx --build` |
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

ส่วน target ที่เหลือใช้มากกว่าหนึ่งบรรทัด แยกไปอยู่ในหัวข้อย่อยด้านล่าง

> **อย่าพิมพ์คำสั่งของ `make eks_*` เองมือเปล่า** เพราะมันสร้าง image tag ที่ฝังวันเวลาไว้
> และห้ามทำเลียนแบบแบบใกล้เคียง ถ้าต้อง deploy ขึ้น EKS ให้ทำจากเครื่องที่มี `make`
> หรือถามทีมก่อน

### A.6 รัน Go backend นอก Docker

`make store_service_dev_mode` ก็แค่เซ็ต environment variable แล้วรัน service ใน cmd ทำแบบนี้:

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

> **อย่าใส่เครื่องหมายคำพูดครอบค่า** เพราะใน cmd คำสั่ง `set X="abc"` จะเก็บเครื่องหมายคำพูด
> เป็นส่วนหนึ่งของค่าไปด้วย แล้ว service จะต่อฐานข้อมูลไม่ได้พร้อม error ที่ชวนงง

ตัวแปรเหล่านี้มีผลเฉพาะในหน้าต่าง Command Prompt นั้น ถ้าเปิดหน้าต่างใหม่ต้องเซ็ตใหม่ทั้งหมด
ดังนั้นให้เปิดหน้าต่างนี้ค้างไว้ตอนพัฒนา และหยุด service ด้วย `Ctrl+C`

### A.7 Unit tests

```bat
:: Go
cd store-service && go test -v ./... && cd ..

:: point-service (Jest)
cd point-service && npm test && cd ..

:: store-web component tests (Cypress)
cd store-web && npm run test:component && cd ..
```

สามคำสั่งนี้รวมกันคือ `make unit_test_all`

ถ้าอยากรันเฉพาะ package เดียวของ Go หรือเฉพาะ Cypress spec เดียว:

```bat
cd store-service && go test -v ./internal/order/... && cd ..
cd store-web && npx cypress run --component --spec "src/components/cart.cy.tsx" && cd ..
```

### A.8 Integration tests

Integration test ของ Go ซ่อนอยู่หลัง build tag ชื่อ `integration` และต้องมี MySQL,
Liquibase และ mock gateway รันอยู่ก่อน เทียบเท่า `make backend_integration_test` ใน cmd คือ:

```bat
docker compose up -d db thirdparty
timeout /t 15
docker compose up liquibase

cd store-service && go test -tags=integration ./... && cd ..

docker compose down
```

(`timeout /t 15` คือ `sleep` เวอร์ชัน Windows ใช้รอให้ MySQL พร้อมรับ connection
ถ้า Liquibase fail ให้รอนานกว่านี้แล้วรันบรรทัด `docker compose up liquibase` ซ้ำ)

### A.9 API tests (Newman)

```bat
:: สตาร์ต stack ก่อน
copy /Y store-web\.env_local store-web\.env
docker compose up -d thirdparty point-service db store-service store-web nginx liquibase --build

cd atdd\api

:: ชุด Authentication — ครบทั้ง 6 เคส
for %F in (TSS-AUTH-001 TSS-AUTH-002 TSS-AUTH-003 TSA-AUTH-001 TSA-AUTH-002 TSA-AUTH-003) do newman run collections\001-Authentication.postman_collection.json --folder "%F" -e sck-online-store.local.postman_environment.json -d data\001-Authentication\%F.json -r cli,junit,htmlextra

:: ชุด Order Summary PDF
for %F in (TSS-OSP-001 TSS-OSP-002) do newman run collections\002-Order-Summary-PDF.postman_collection.json --folder "%F" -e sck-online-store.local.postman_environment.json -d data\002-Order-Summary-PDF\%F.json -r cli,junit,htmlextra

cd ..\..
docker compose down
```

> ถ้าเอาบรรทัด `for` พวกนี้ไปเซฟเป็นไฟล์ `.bat` แทนการพิมพ์เอง ต้องใส่เครื่องหมายเปอร์เซ็นต์
> เป็นสองตัว: `%%F` แทน `%F` นี่เป็นข้อกำหนดของ cmd ไม่ใช่พิมพ์ผิด

รายงานจะอยู่ที่ `atdd\api\newman\`

### A.10 UI tests (Robot Framework)

ตรงนี้แหละที่ Windows ล้วนสบายกว่า WSL เพราะใช้ Chrome ปกติของเครื่องได้ คุณจึงนั่งดูเทสต์
รันไปทีละขั้นได้ด้วย

```bat
:: สตาร์ต stack (เหมือน A.9)
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

หมายเหตุ:

- การ activate บน Windows คือ `.venv\Scripts\activate` — **ไม่ใช่** `source .venv/bin/activate`
  ซึ่งเป็นแบบที่ Makefile ใช้และเป็นแบบที่คุณจะเห็นในคำสั่งฝั่ง Linux
- บรรทัด `python -m venv` กับ `pip install` ทำแค่ครั้งแรกครั้งเดียว หลังจากนั้นแค่
  `.venv\Scripts\activate` แล้วตามด้วย `robot …` ก็พอ
- Makefile ส่ง `-v REMOTE_HUB_URL:…` มาด้วย แต่คุณไม่ต้องใส่ก็ได้ เพราะค่า default คือค่าว่าง
  ซึ่งแปลว่า "ใช้เบราว์เซอร์บนเครื่อง" ตรงกับที่คุณต้องการพอดี
- ผลลัพธ์: ไฟล์ `log.html` และ `report.html` ในโฟลเดอร์ `atdd\ui` เปิดดูด้วยเบราว์เซอร์ได้เลย

### A.11 จำเป็นต้องลง GNU Make ไหม

ลงก็ได้ (`winget install GnuWin32.Make`) แล้ว target แบบบรรทัดเดียวในตาราง §A.5 จะใช้ได้
แต่ target เหล่านี้จะยังพังใน cmd อยู่ดี เพราะข้างในใช้คำสั่งที่มีแต่บน Linux:

| Target | ทำไมถึงพังบน Windows ปกติ |
| --- | --- |
| `setup_test_fixtures`, `backend_integration_test` | ใช้ `sleep` |
| `start_test_suite`, `start_test_suite_grid` | ใช้ `cp -f` |
| `run_robot*` | ใช้ `python3` และ `source .venv/bin/activate` |
| `store_service_dev_mode` | ใช้ syntax env แบบ POSIX `VAR=value cmd` |
| `eks_*` | ใช้ shell substitution `$(date …)` |

ในเมื่อตัวที่พังคือตัวที่ใช้งานจริงเสียส่วนใหญ่ การลง Make จึงแทบไม่ช่วยอะไร คำสั่งในภาคผนวกนี้
ใช้แทนได้ครบทุกอย่างอยู่แล้ว และนั่นคือสิ่งที่แนะนำ

### A.12 ข้อควรระวังเฉพาะ Windows

**ใช้ `cmd` ไม่ใช่ PowerShell** หรืออย่างน้อยต้องรู้ความต่าง: PowerShell 5.1 (ตัวที่ติดมากับ
Windows) ไม่รองรับ `&&` ระหว่างคำสั่ง และการเซ็ต environment variable ต้องเขียนเป็น
`$env:JWT_SECRET = "value"` แทน `set JWT_SECRET=value` ทุกตัวอย่างในภาคผนวกนี้อ้างอิง
Command Prompt

**เรื่อง line ending ของ Git** ตั้งค่านี้ครั้งเดียวก่อน commit แรก จะได้ไม่ทำให้ทุกไฟล์
กลายเป็นการแก้ไขใน diff:

```bat
git config --global core.autocrlf true
```

**Antivirus และ Windows Defender** ที่สแกน `node_modules` ทำให้ `npm install` และการ build
Next.js ช้าลงหลายเท่า ถ้ารู้สึกว่า build ช้าผิดปกติ ให้เพิ่มโฟลเดอร์ repo และ
`%USERPROFILE%\go` เข้าไปในรายการยกเว้นของ Defender

**พาธยาวเกิน** โฟลเดอร์ `node_modules` อาจยาวเกินขีดจำกัดเดิมที่ 260 ตัวอักษร ถ้าเจอ error
"filename too long" ตอน `npm install` หรือ `git clone`:

```bat
git config --global core.longpaths true
```

**เจอ "docker: command not found"** เกือบทุกครั้งแปลว่า Docker Desktop ยังไม่ได้เปิด
(ดูที่ system tray) หรือคุณเปิด Command Prompt ไว้ก่อนติดตั้ง Docker (ให้เปิดหน้าต่างใหม่)

---

## อ่านต่อที่ไหน

- [`CLAUDE.md`](../CLAUDE.md) — สถาปัตยกรรม, make target ทั้งหมด, convention การตั้งชื่อ
- [`README.md`](../README.md) — convention การเขียนโค้ด (ภาษาไทย)
- [`atdd/CLAUDE.md`](../atdd/CLAUDE.md) — โครงสร้างชุดเทสต์ API และ UI
- [`thirdparty/CLAUDE.md`](../thirdparty/CLAUDE.md) — mock payment/shipping gateway
- [`deploy/README.md`](../deploy/README.md) — Kubernetes และ Terraform
