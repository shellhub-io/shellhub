import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Dashboard from '@/views/Dashboard';

const localVue = createLocalVue();
localVue.use(Vuex);

const store = new Vuex.Store({
  namespaced: true,
  state: {
    stats: {
      registeredDevices: 2,
      onlineDevices: 1,
      activeSessions: 1,
    },
  },
  getters: {
    'stats/stats': (state) => state.stats,
  },
  actions: {
    'stats/get': () => {
    },
  },
});

describe('Dashboard', () => {
  const wrapper = shallowMount(Dashboard, {
    store,
    localVue,
    stubs: ['fragment'],
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('has restered devices', () => {
    expect(wrapper.vm.stats.registeredDevices).toBe(2);
  });
  it('has online devices', () => {
    expect(wrapper.vm.stats.onlineDevices).toBe(1);
  });
  it('has actived sessions', () => {
    expect(wrapper.vm.stats.activeSessions).toBe(1);
  });
  it('has a device add button', () => {
    expect(wrapper.find('[data-cy="addDevice-btn"]').exists()).toBe(true);
  });
  it('has a view all devices button', () => {
    expect(wrapper.find('[data-cy="viewDevices-btn"]').exists()).toBe(true);
  });
  it('has a view all sessions button', () => {
    expect(wrapper.find('[data-cy="viewSessions-btn"]').exists()).toBe(true);
  });
});
