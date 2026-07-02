export const decodeJWT = (token: string) => {
  const payloadB64 = token.split('.')[1]
  const payloadJson = JSON.parse(
    atob(payloadB64.replace(/-/g, '+').replace(/_/g, '/'))
  )
  return payloadJson
}

export const isTokenExpired = (token: string) => {
  try {
    const { exp } = decodeJWT(token)
    return !exp || Date.now() >= exp * 1000
  } catch {
    return true
  }
}
