import { useUserStore } from '@/hooks/use-user-store'
import AuthLayout from '@/layouts/common/auth'
import { AppRouterContext } from 'next/dist/shared/lib/app-router-context.shared-runtime'

const buildToken = (payload: object) => {
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))
  const body = btoa(JSON.stringify(payload))
  return `${header}.${body}.signature`
}

const validUser = {
  user_id: 1,
  first_name: 'Nattapon',
  last_name: 'S',
  username: 'nattapon.s'
}

const mountAuthLayout = (push: (href: string) => void) => {
  const router = {
    push,
    replace: () => {},
    back: () => {},
    forward: () => {},
    refresh: () => {},
    prefetch: () => {}
  }

  cy.mount(
    <AppRouterContext.Provider value={router}>
      <AuthLayout>
        <div>login form</div>
      </AuthLayout>
    </AppRouterContext.Provider>
  )
}

describe('<AuthLayout />', () => {
  beforeEach(() => {
    localStorage.clear()
    useUserStore.getState().clearUser()
  })

  it('Should redirect to /product/list when the access token is still valid', () => {
    // Arrange
    const token = buildToken({
      ...validUser,
      exp: Math.floor(Date.now() / 1000) + 3600
    })
    localStorage.setItem('accessToken', token)
    const push = cy.stub().as('push')

    // Act
    mountAuthLayout(push)

    // Assert
    cy.get('@push').should('have.been.calledWith', '/product/list')
  })

  it('Should clear the stale token and stay on the login page when the access token is expired', () => {
    // Arrange
    const token = buildToken({
      ...validUser,
      exp: Math.floor(Date.now() / 1000) - 3600
    })
    localStorage.setItem('accessToken', token)
    const push = cy.stub().as('push')

    // Act
    mountAuthLayout(push)

    // Assert
    cy.get('@push').should('not.have.been.called')
    cy.contains('login form').should('be.visible')
    cy.wrap(null).should(() => {
      expect(localStorage.getItem('accessToken')).to.equal(null)
      expect(useUserStore.getState().user).to.equal(null)
    })
  })

  it('Should render the login page without redirecting when there is no access token', () => {
    // Arrange
    const push = cy.stub().as('push')

    // Act
    mountAuthLayout(push)

    // Assert
    cy.get('@push').should('not.have.been.called')
    cy.contains('login form').should('be.visible')
  })
})
