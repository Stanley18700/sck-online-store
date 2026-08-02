import { subTotal, totalPayment } from '@/utils/total-price'

describe('Utils > total price > Sub Total', () => {
  it('ต้องการเห็น ผลรวมสินค้า 620 บาท จากรายการสินค้าทั้งหมด', () => {
    const priceList = [
      {
        price: 100,
        quantity: 1
      },
      {
        price: 200,
        quantity: 2
      },
      {
        price: 120,
        quantity: 1
      }
    ]
    const actual = 620

    const total = subTotal(priceList)

    expect(total).to.equal(actual)
  })
})

// ส่วนลดคำนวณโดย point-service (2 แต้ม = 1.00 บาท) — ฟังก์ชันนี้แค่รวมยอด
describe('Utils > total price > Total Price', () => {
  it('ต้องการเห็น เงินทั้งหมด 550 บาท จากราคารวมสินค้าทั้งหมด 500 บาท และค่าขนส่ง 50 บาท โดยไม่ใช้ส่วนลดจากแต้ม', () => {
    const isUsePoint = false
    const discount = 0
    const subTotal = 500
    const shippingFee = 50
    const actual = 550

    const total = totalPayment(isUsePoint, discount, subTotal, shippingFee)

    expect(total).to.equal(actual)
  })

  it('ต้องการเห็น เงินทั้งหมด 500 บาท จากราคารวมสินค้าทั้งหมด 500 บาท และค่าขนส่ง 50 บาท โดยใช้ 100 แต้ม (ส่วนลด 50 บาท)', () => {
    const isUsePoint = true
    const discount = 50 // 100 points = 50.00 THB
    const subTotal = 500
    const shippingFee = 50
    const actual = 500

    const total = totalPayment(isUsePoint, discount, subTotal, shippingFee)

    expect(total).to.equal(actual)
  })

  it('ต้องการเห็น เงินทั้งหมด 100 บาท จากราคารวมสินค้าทั้งหมด 100 บาท และค่าขนส่ง 50 บาท โดยใช้ 100 แต้ม (ส่วนลด 50 บาท)', () => {
    const isUsePoint = true
    const discount = 50 // 100 points = 50.00 THB
    const subTotal = 100
    const shippingFee = 50
    const actual = 100

    const total = totalPayment(isUsePoint, discount, subTotal, shippingFee)

    expect(total).to.equal(actual)
  })

  it('ต้องการเห็น เงินทั้งหมด 75 บาท จากราคารวมสินค้าทั้งหมด 100 บาท และค่าขนส่ง 50 บาท โดยใช้ 150 แต้ม (ส่วนลด 75 บาท)', () => {
    const isUsePoint = true
    const discount = 75 // 150 points = 75.00 THB
    const subTotal = 100
    const shippingFee = 50
    const actual = 75

    const total = totalPayment(isUsePoint, discount, subTotal, shippingFee)

    expect(total).to.equal(actual)
  })

  it('ต้องการเห็น เงินทั้งหมดเท่ากับค่าขนส่ง เมื่อส่วนลดเท่ากับราคารวมสินค้า (กันพลาด — ปกติ point-service จะไม่ให้เกิดขึ้น)', () => {
    const isUsePoint = true
    const discount = 100
    const subTotal = 100
    const shippingFee = 50
    const actual = 50

    const total = totalPayment(isUsePoint, discount, subTotal, shippingFee)

    expect(total).to.equal(actual)
  })
})
