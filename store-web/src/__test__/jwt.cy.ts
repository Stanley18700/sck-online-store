import { isTokenExpired } from '@/utils/jwt'

const buildToken = (payload: object) => {
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))
  const body = btoa(JSON.stringify(payload))
  return `${header}.${body}.signature`
}

describe('Utils > jwt > isTokenExpired', () => {
  it('Should return false when the token exp is in the future', () => {
    const token = buildToken({ exp: Math.floor(Date.now() / 1000) + 3600 })

    expect(isTokenExpired(token)).to.equal(false)
  })

  it('Should return true when the token exp is in the past', () => {
    const token = buildToken({ exp: Math.floor(Date.now() / 1000) - 3600 })

    expect(isTokenExpired(token)).to.equal(true)
  })

  it('Should return true when the token has no exp claim', () => {
    const token = buildToken({ user_id: 1 })

    expect(isTokenExpired(token)).to.equal(true)
  })

  it('Should return true when the token is malformed', () => {
    expect(isTokenExpired('not-a-jwt')).to.equal(true)
  })
})
