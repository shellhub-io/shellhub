import store from '@/store';

describe('Notification', () => {
  it('returns notifications default', () => {
    const actual = store.getters['notifications/list'];

    expect(actual).toEqual([]);
  });
  it('returns number notifications default', () => {
    const actual = store.getters['notifications/getNumberNotifications'];

    expect(actual).toEqual(0);
  });
  it('complete test', () => {
    const notifications = [
      {
        uid: 'a582b47a42d',
        name: '39-5e-2b',
        identity: {
          mac: '00:00:00:00:00:00',
        },
        info: {
          id: 'debian',
          pretty_name: 'Debian GNU/Linux 10 (buster)',
          version: 'latest',
        },
        public_key: '----- PUBLIC KEY -----',
        tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
        last_seen: '2020-07-17T16:47:08.31Z',
        online: false,
        namespace: 'user',
        status: 'pending',
      },
      {
        uid: 'a582b47a42e',
        name: '39-5e-2d',
        identity:
        {
          mac: '00:00:00:00:00:00',
        },
        info: {
          id: 'debian',
          pretty_name: 'Debian GNU/Linux 10 (buster)',
          version: 'latest',
        },
        public_key: '----- PUBLIC KEY -----',
        tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
        last_seen: '2020-07-17T16:47:08.31Z',
        online: false,
        namespace: 'user',
        status: 'pending',
      },
    ];

    const numberNotifications = 2;

    store.commit('notifications/setNotifications', { data: notifications, headers: { 'x-total-count': numberNotifications } });
    expect(store.getters['notifications/list']).toEqual(notifications);
    expect(store.getters['notifications/getNumberNotifications']).toEqual(numberNotifications);
  });
});
