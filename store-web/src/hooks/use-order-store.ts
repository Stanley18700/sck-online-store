import SHIPPING_METHOD from '@/assets/data/shipping_method.json'
import GetProductInCartService, {
  ProductDetailInCart
} from '@/services/cart/get-product-list'
import getDiscountQuoteService from '@/services/discount'
import * as priceCalculate from '@/utils/total-price'
import type {} from '@redux-devtools/extension' // required for devtools typing
import { produce } from 'immer'
import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

// ---------------------------------------------------

type PaymentCreditInformationType = {
  name: string
  number: string
  expiry: string
  cvv: string
  issuer: string
  focused: string
}

type ShippingInformationType = {
  firstName: string
  lastName: string
  address: string
  mobileNumber: string
  provinceId: number
  districtId: number
  subDistrictId: number
  provinceName: string
  districtName: string
  subDistrictName: string
  zipCode: number
  focused: string
}

type ShippingType = {
  shippingMethod: number
  shippingFee: number
  shippingInformation: ShippingInformationType
}

type PointType = {
  point: number // Current Point
  isUsePoint: boolean
  burnPoint: number
  discountThb: number // Server-quoted discount for burnPoint (2 points = 1.00 THB)
}

type PaymentType = {
  paymentMethod: number
  paymentCreditInformation: PaymentCreditInformationType
}

type OrderSummaryType = {
  total_price: number
  total_price_thb: number
  total_price_full_thb: number
  receive_point: number
}

type OrderStoreType = {
  cart: ProductDetailInCart[]
  summary: OrderSummaryType
  totalProduct: number
  // subTotal: number
  totalPayment: number
  receivePoint: number
  shipping: ShippingType
  point: PointType
  payment: PaymentType
  getProductListInCart: () => void
  setPoint: (point: number) => void
  setIsUsePoint: (isUsePoint: boolean) => void
  setPaymentMethod: (paymentMethod: string) => void
  setPaymentInformation: (
    paymentCreditInformation: PaymentCreditInformationType
  ) => void
  setShippingMethod: (shippingMethod: number, shippingFee: number) => void
  setShippingInformation: (shippingInformation: ShippingInformationType) => void
  updateSummary: () => void
}

const useOrderStore = create<OrderStoreType>()(
  devtools((set, get) => ({
    cart: [],
    summary: {
      total_price: 0,
      total_price_thb: 0,
      total_price_full_thb: 0,
      receive_point: 0
    },
    totalProduct: 0,
    // subTotal: 0,
    totalPayment: 0,
    receivePoint: 0,
    shipping: {
      shippingMethod: SHIPPING_METHOD[0].id,
      shippingFee: SHIPPING_METHOD[0].price,
      shippingInformation: {
        firstName: '',
        lastName: '',
        address: '',
        mobileNumber: '',
        provinceId: 0,
        districtId: 0,
        subDistrictId: 0,
        provinceName: '',
        districtName: '',
        subDistrictName: '',
        zipCode: 0,
        focused: ''
      }
    },
    point: {
      point: 0,
      burnPoint: 0,
      isUsePoint: false,
      discountThb: 0
    },
    payment: {
      paymentMethod: 1,
      paymentCreditInformation: {
        number: '',
        name: '',
        expiry: '',
        cvv: '',
        issuer: '',
        focused: ''
      }
    },
    getProductListInCart: async () => {
      // Mock userId
      const productInCart = await GetProductInCartService()

      if (productInCart.data) {
        // const price = productInCart.data.map((item) => {
        //   return {
        //     price: item.product_price_thb,
        //     quantity: item.quantity
        //   }
        // })

        // const total = priceCalculate.subTotal(price)

        set(
          produce((state) => {
            state.totalProduct = productInCart.data?.carts.length
            state.cart = productInCart.data?.carts
            state.summary = productInCart.data?.summary

            state.subTotal = productInCart.data?.summary.total_price_thb
            state.totalPayment = productInCart.data?.summary.total_price_thb
          })
        )
      }

      // Reset Discount Point
      set(
        produce((state) => {
          state.point.burnPoint = 0
          state.point.isUsePoint = false
          state.point.discountThb = 0
        })
      )

      get().updateSummary()
    },
    setPoint(point: number) {
      set(
        produce((state) => {
          state.point.point = point
        })
      )
    },
    setIsUsePoint: async (isUsePoint: boolean) => {
      if (!isUsePoint) {
        set(
          produce((state) => {
            state.point.isUsePoint = false
            state.point.burnPoint = 0
            state.point.discountThb = 0
          })
        )

        get().updateSummary()
        return
      }

      // The burn and discount amounts come from point-service - never computed here
      const balance = get().point.point
      const subTotal = get().summary.total_price_thb
      const result = await getDiscountQuoteService(balance, subTotal)

      set(
        produce((state) => {
          state.point.isUsePoint = true
          state.point.burnPoint = result.data ? result.data.burn_point : 0
          state.point.discountThb = result.data ? result.data.discount : 0
        })
      )

      get().updateSummary()
    },
    setPaymentMethod: (paymentMethod: string) => {
      set(
        produce((state) => {
          state.payment.paymentMethod = paymentMethod
        })
      )
    },
    setPaymentInformation: (
      paymentCreditInformation: PaymentCreditInformationType
    ) => {
      set(
        produce((state) => {
          state.payment.paymentCreditInformation = paymentCreditInformation
        })
      )
    },
    setShippingMethod: (shippingMethod: number, shippingFee: number) => {
      const newTotalPayment = get().summary.total_price_thb + shippingFee

      set(
        produce((state) => {
          state.totalPayment = newTotalPayment
          state.shipping.shippingMethod = shippingMethod
          state.shipping.shippingFee = shippingFee
        })
      )

      get().updateSummary()
    },
    setShippingInformation: (shippingInformation: ShippingInformationType) => {
      set(
        produce((state) => {
          state.shipping.shippingInformation = shippingInformation
        })
      )
    },
    updateSummary: async () => {
      const isUsePoint = get().point.isUsePoint
      const discountThb = get().point.discountThb

      const subTotal = get().summary.total_price_thb
      const shippingFee = get().shipping.shippingFee

      // Total Payment (the discount itself was quoted by point-service)
      const totalPayment = priceCalculate.totalPayment(
        isUsePoint,
        discountThb,
        subTotal,
        shippingFee
      )

      // Point Receive (from backend cart summary via point-service)
      const receivePoint = get().summary.receive_point

      set(
        produce((state) => {
          state.totalPayment = totalPayment
          state.receivePoint = receivePoint
        })
      )
    }
  }))
)

export default useOrderStore
