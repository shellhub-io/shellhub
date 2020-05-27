import store from '../../src/store.js';

describe('Auth', () => {
  it('returns status', () => {
    const actual = store.getters['auth/status'];
    expect(actual).toEqual();
  });
  it('returns token', () => {
    const actual = store.getters['auth/token'];
    expect(actual).toEqual();
  });
  it('returns user', () => {
    const actual = store.getters['auth/user'];
    expect(actual).toEqual();
  });
  it('returns tenant', () => {
    const actual = store.getters['auth/tenant'];
    expect(actual).toEqual('');
  });
  it('complete test', () => {
    const statusLoading = 'loading';
    const statusError = 'error';
    const statusSuccess = 'success';
    const token = 'eyJhbGciOiJSUzI1NiIsInR5c';
    const user = 'user';
    const tenant = '00000000';

    store.commit('auth/authRequest');
    expect(store.getters['auth/authStatus']).toEqual(statusLoading);

    store.commit('auth/authError');
    expect(store.getters['auth/authStatus']).toEqual(statusError);

    store.commit('auth/authSuccess', {token: token, user: user, tenant: tenant});
    expect(store.getters['auth/authStatus']).toEqual(statusSuccess);
    expect(store.getters['auth/isLoggedIn']).toEqual(true);
    expect(store.getters['auth/currentUser']).toEqual(user);
    expect(store.getters['auth/tenant']).toEqual(tenant);

    store.commit('auth/logout');
    expect(store.getters['auth/authStatus']).toEqual('');
    expect(store.getters['auth/isLoggedIn']).toEqual(false);
    expect(store.getters['auth/currentUser']).toEqual('');
    expect(store.getters['auth/tenant']).toEqual('');    
  });
});
