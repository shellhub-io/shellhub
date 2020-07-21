import store from '@/store';

describe('Sessions', () => {
  it('returns sessions', () => {
    const actual = store.getters['sessions/list'];
    expect(actual).toEqual([]);
  });
  it('return session', () => {
    const actual = store.getters['sessions/get'];
    expect(actual).toEqual([]);
  });
  it('returns number session', () => {
    const actual = store.getters['sessions/getNumberSessions'];
    expect(actual).toEqual(0);
  });
  it('complete test', () => {
    const sessions = [
      {
        uid: '8c354a00f50',
        device_uid: 'a582b47a42d',
        device: {
          uid: 'a582b47a42d',
          name: '39-5e-2a',
          identity: {
            mac: '00:00:00:00:00:00',
          },
          info: {
            id: 'debian',
            pretty_name: 'Debian GNU/Linux 10 (buster)',
            version: 'v0.2.5',
          },
          public_key: '----- PUBLIC KEY -----',
          tenant_id: '00000000',
          last_seen: '2020-05-18T13:27:02.498Z',
          online: false,
          namespace: 'user',
        },
        tenant_id: '00000000',
        username: 'user',
        ip_address: '000.000.000.000',
        started_at: '2020-05-18T12:30:28.824Z',
        last_seen: '2020-05-18T12:30:30.205Z',
        active: false,
        authenticated: false,
      },
      {
        uid: '8c354a00f50',
        device_uid: 'a582b47a42d',
        device: {
          uid: 'a582b47a42d',
          name: 'b4-2e-99',
          identity: {
            mac: '00:00:00:00:00:00',
          },
          info: {
            id: 'debian',
            pretty_name: 'Debian GNU/Linux 10 (buster)',
            version: 'v0.2.5',
          },
          public_key: '----- PUBLIC KEY -----',
          tenant_id: '00000000',
          last_seen: '2020-05-18T13:27:02.498Z',
          online: false,
          namespace: 'user',
        },
        tenant_id: '00000000',
        username: 'user',
        ip_address: '000.000.000.000',
        started_at: '2020-05-18T12:30:28.824Z',
        last_seen: '2020-05-18T12:30:30.205Z',
        active: false,
        authenticated: false,
      },
    ];
    const session = {
      uid: '8c354a00f50',
      device_uid: 'a582b47a42d',
      device: {
        uid: 'a582b47a42d',
        name: '39-5e-2a',
        identity: {
          mac: '00:00:00:00:00:00',
        },
        info: {
          id: 'debian',
          pretty_name: 'Debian GNU/Linux 10 (buster)',
          version: 'v0.2.5',
        },
        public_key: '----- PUBLIC KEY -----',
        tenant_id: '00000000',
        last_seen: '2020-05-18T13:27:02.498Z',
        online: false,
        namespace: 'user',
      },
      tenant_id: '00000000',
      username: 'user',
      ip_address: '000.000.000.000',
      started_at: '2020-05-18T12:30:28.824Z',
      last_seen: '2020-05-18T12:30:30.205Z',
      active: false,
      authenticated: false,
    };
    const numberSessions = 2;

    store.commit('sessions/setSessions', { data: sessions, headers: { 'x-total-count': numberSessions } });
    store.commit('sessions/setSession', session);

    expect(store.getters['sessions/list']).toEqual(sessions);
    expect(store.getters['sessions/get']).toEqual(session);
    expect(store.getters['sessions/getNumberSessions']).toEqual(numberSessions);
  });
});
