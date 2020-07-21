import store from '@/store';

describe('Devices', () => {
  it('returns devices default', () => {
    const actual = store.getters['devices/list'];
    expect(actual).toEqual([]);
  });
  it('return device default', () => {
    const actual = store.getters['devices/get'];
    expect(actual).toEqual([]);
  });
  it('returns number devices default', () => {
    const actual = store.getters['devices/getNumberDevices'];
    expect(actual).toEqual(0);
  });
  it('complete test', () => {
    const devices = [
      {
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
      {
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
    ];
    const device = {
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
    };
    const numberDevices = 2;

    store.commit('devices/setDevices', { data: devices, headers: { 'x-total-count': numberDevices } });
    store.commit('devices/setDevice', device);

    expect(store.getters['devices/list']).toEqual(devices);
    expect(store.getters['devices/get']).toEqual(device);
    expect(store.getters['devices/getNumberDevices']).toEqual(numberDevices);
  });
});
