import axiosShoppingMallApi from '@/utils/axios'
import { handleServiceError } from '@/utils/helper'

// ------------------------------------------------

export type CalculatePointServiceResponse = {
  data?: {
    point: number
  }
  message?: string
}

const calculatePointService = async (
  amount: number
): Promise<CalculatePointServiceResponse> => {
  try {
    const { data } = await axiosShoppingMallApi.get(
      `/api/v1/point/calculate?amount=${amount}`
    )
    return {
      data: data
    }
  } catch (error) {
    return handleServiceError(error)
  }
}

export default calculatePointService
