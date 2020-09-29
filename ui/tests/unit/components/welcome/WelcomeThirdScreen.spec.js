import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import WelcomeThirdScreen from '@/components/welcome/WelcomeThirdScreen';

describe('WelcomeThirdScreen', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

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
    online: true,
    namespace: 'user',
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      device,
    },
    getters: {
      'devices/getFirstPending': (state) => state.device,
    },
    actions: {
      'devices/setFirstPending': () => {
      },
      'modals/showSnackbarErrorLoading': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(WelcomeThirdScreen, {
      store,
      localVue,
      stubs: ['fragment'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.getPendingDevice).toEqual(device);
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="deviceName-field"]').text()).toEqual(device.name);
    expect(wrapper.find('[data-test="devicePrettyName-field"]').text()).toEqual(device.info.pretty_name);
  });
});
