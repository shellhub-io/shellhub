import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Welcome from '@/components/welcome/Welcome';

describe('Welcome', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const show = true;
  const tenant = 'a582b47a42e';

  const stats = {
    registered_devices: 0,
    online_devices: 0,
    active_sessions: 0,
    pending_devices: 1,
    rejected_devices: 0,
  };

  const devicePending = {
    uid: 'a582b47a',
    name: '39-5e-2b',
    identity: {
      mac: '00:00:00',
    },
    info: {
      id: 'linuxmint',
      pretty_name: 'Linux',
      version: '',
    },
    public_key: 'xxxxxxxx',
    tenant_id: 'xxxxxxxx',
    last_seen: '2020-05-20T19:58:53.276Z',
    online: true,
    namespace: 'user',
    status: 'accepted',
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      tenant,
      stats,
      devicePending,
    },
    getters: {
      'auth/tenant': (state) => state.tenant,
      'stats/stats': (state) => state.stats,
      'devices/getFirstPending': (state) => state.devicePending,
    },
    actions: {
      'stats/get': () => {},
      'devices/accept': () => {},
      'notifications/fetch': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(Welcome, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { show },
    });
  });

  ///////
  // Component Rendering
  //////

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  ///////
  // Data and Props checking
  //////

  it('Receive data in props', () => {
    expect(wrapper.vm.show).toEqual(show);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.e1).toEqual(1);
    expect(wrapper.vm.enable).toEqual(false);
    expect(wrapper.vm.polling).toEqual(null);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.showWelcome).toEqual(show);
  });
  it('Process data in the methods', () => {
    const command = `curl -sSf "http://localhost/install.sh?tenant_id=${tenant}" | sh`;

    expect(wrapper.vm.command()).toEqual(command);
  });

  //////
  // HTML validation
  //////

  it('Compare data with default value', async () => {
    expect(wrapper.vm.e1).toEqual(1);
    expect(wrapper.vm.enable).toEqual(false);
    expect(wrapper.vm.curl.hostname).toEqual('localhost');
    expect(wrapper.vm.curl.tenant).toEqual(tenant);

    //////
    // In this case is tested the click event.
    //////

    wrapper.find('[data-test="firstClick-btn"]').vm.$emit('click');
    expect(wrapper.vm.e1).toEqual(2);

    wrapper.setData({ enable: true });

    await localVue.nextTick();
    wrapper.find('[data-test="secondClick-btn"]').vm.$emit('click');
    expect(wrapper.vm.e1).toEqual(3);

    await localVue.nextTick();
    expect(wrapper.find('[data-test="thirdClick-btn"]').exists()).toBe(true);

    await wrapper.vm.acceptDevice();
    expect(wrapper.vm.e1).toEqual(4);
  });
});
