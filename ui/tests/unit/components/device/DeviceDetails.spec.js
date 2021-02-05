import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceDetails from '@/components/device/DeviceDetails';

describe('DeviceDetails', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  const owner = true;

  const device = {
    uid: 'a582b47a42d',
    name: '39-5e-2a',
    identity: {
      mac: '00:00:00:00:00:00',
    },
    info: {
      id: 'arch',
      pretty_name: 'Linux Mint 19.3',
      version: '',
    },
    public_key: '----- PUBLIC KEY -----',
    tenant_id: '00000000',
    last_seen: '2020-05-20T18:58:53.276Z',
    online: false,
    namespace: 'user',
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      device,
      owner,
    },
    getters: {
      'devices/get': (state) => state.device,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'devices/get': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(DeviceDetails, {
      store,
      localVue,
      stubs: ['fragment'],
      mocks: {
        $route: {
          params: {
            id: device.uid,
          },
        },
      },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  Object.keys(device).forEach((field) => {
    it(`Receives the field ${field} of device state from store`, () => {
      expect(wrapper.vm.device[field]).toEqual(device[field]);
    });
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="deviceUid-field"]').text()).toEqual(device.uid);
    expect(wrapper.find('[data-test="deviceMac-field"]').text()).toEqual(device.identity.mac);
    expect(wrapper.find('[data-test="devicePrettyName-field"]').text()).toEqual(device.info.pretty_name);
    expect(wrapper.find('[data-test="deviceConvertDate-field"]').text()).toEqual('Wednesday, May 20th 2020, 6:58:53 pm');
  });
});
