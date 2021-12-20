import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Dashboard from '@/views/Dashboard';

describe('Dashboard', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const numberNamespaces1 = 3;

  const namespace1 = {
    name: 'namespace1',
    owner: 'user1',
    members: [{ name: 'user3' }, { name: 'user4' }, { name: 'user5' }],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484713',
  };

  const statsDev = {
    registered_devices: 2,
    pending_devices: 1,
    rejected_devices: 1,
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      stats: statsDev,
      namespace: namespace1,
      numberNamespaces: numberNamespaces1,
    },
    getters: {
      'stats/stats': (state) => state.stats,
      'namespaces/get': (state) => state.namespace,
    },
    actions: {
      'stats/get': () => {},
      'users/setStatusUpdateAccountDialog': () => {},
    },
  });

  beforeEach(() => {
    localStorage.setItem('namespacesWelcome', JSON.stringify({}));
    wrapper = shallowMount(Dashboard, {
      store,
      localVue,
      stubs: ['fragment'],
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

  it('Compare data with default value', () => {
    expect(wrapper.vm.stats.registered_devices).toBe(2);
    expect(wrapper.vm.stats.pending_devices).toBe(1);
    expect(wrapper.vm.stats.rejected_devices).toBe(1);
  });

  //////
  // HTML validation
  //////

  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="addDevice-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="viewDevices-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="viewSessions-btn"]').exists()).toBe(true);
  });
});
