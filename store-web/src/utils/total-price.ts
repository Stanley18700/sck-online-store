// ----------------------------------------------------------------------------

type SubTotalType = {
  price: number
  quantity: number
}

export const subTotal = (priceList: SubTotalType[]): number => {
  let total = 0

  for (let i = 0; i < priceList.length; i++) {
    total += priceList[i].price * priceList[i].quantity
  }

  return total
}

// The discount amount is quoted by point-service (2 points = 1.00 THB,
// capped at the subtotal) - this function only composes the final total.
// Shipping is never discounted.
export const totalPayment = (
  isUsePoint: boolean,
  discount: number,
  subTotal: number,
  shippingFee: number
) => {
  let totalPayment = 0

  if (isUsePoint) {
    if (subTotal <= discount) {
      totalPayment = shippingFee
    } else {
      totalPayment = subTotal - discount + shippingFee
    }
  } else {
    totalPayment = subTotal + shippingFee
  }

  return totalPayment
}
