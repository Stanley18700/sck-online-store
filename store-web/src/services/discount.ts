import axiosShoppingMallApi from '@/utils/axios'
import { handleServiceError } from '@/utils/helper'

// ------------------------------------------------

export type DiscountQuoteServiceResponse = {
  data?: {
    burn_point: number
    discount: number
  }
  message?: string
}

const getDiscountQuoteService = async (
  points: number,
  subtotal: number
): Promise<DiscountQuoteServiceResponse> => {
  try {
    const { data } = await axiosShoppingMallApi.get(
      `/api/v1/point/discount?points=${points}&subtotal=${subtotal}`
    )
    return {
      data: data
    }
  } catch (error) {
    return handleServiceError(error)
  }
}

export default getDiscountQuoteService
