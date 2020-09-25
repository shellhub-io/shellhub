import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Dashboard from '@/views/Dashboard';

describe('Dashboard', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const stats = {
    registeredDevices: 2,
    onlineDevices: 1,
    activeSessions: 1,
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      stats,
    },
    getters: {
      'stats/stats': (state) => state.stats,
    },
    actions: {
      'stats/get': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(Dashboard, {
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
  it('Compare data with default value', () => {
    expect(wrapper.vm.stats.registeredDevices).toBe(2);
    expect(wrapper.vm.stats.onlineDevices).toBe(1);
    expect(wrapper.vm.stats.activeSessions).toBe(1);
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-cy="addDevice-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-cy="viewDevices-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-cy="viewSessions-btn"]').exists()).toBe(true);
  });
});
