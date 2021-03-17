import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Dashboard from '@/views/Dashboard';
import 'mock-local-storage';

describe('Dashboard', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  let wrapper2;

  const numberNamespaces1 = 3;
  const numberNamespaces2 = 1;

  const namespace1 = {
    name: 'namespace1',
    owner: 'user1',
    members: [{ name: 'user3' }, { name: 'user4' }, { name: 'user5' }],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484713',
  };

  const namespace2 = {
    name: 'namespace2',
    owner: 'user4',
    members: [{ name: 'user4' }, { name: 'user5' }],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484712',
  };

  const statsDev = {
    registered_devices: 2,
    pending_devices: 1,
    rejected_devices: 1,
  };

  const statsNoDevices = {
    registered_devices: 0,
    pending_devices: 0,
    rejected_devices: 0,
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
      'namespaces/getNumberNamespaces': (state) => state.numberNamespaces,
      'namespaces/get': (state) => state.namespace,
    },
    actions: {
      'stats/get': () => {
      },
      'namespaces/get': () => {
      },
      'auth/setShowWelcomeScreen': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
    },
  });

  const storeNoDevices = new Vuex.Store({
    namespaced: true,
    state: {
      stats: statsNoDevices,
      namespace: namespace2,
      numberNamespaces: numberNamespaces2,
    },
    getters: {
      'stats/stats': (state) => state.stats,
      'namespaces/getNumberNamespaces': (state) => state.numberNamespaces,
      'namespaces/get': (state) => state.namespace,
    },
    actions: {
      'stats/get': () => {
      },
      'namespaces/get': () => {
      },
      'auth/setShowWelcomeScreen': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
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

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.stats.registered_devices).toBe(2);
    expect(wrapper.vm.stats.pending_devices).toBe(1);
    expect(wrapper.vm.stats.rejected_devices).toBe(1);
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-cy="addDevice-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-cy="viewDevices-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-cy="viewSessions-btn"]').exists()).toBe(true);
  });
  it('The welcome screen loads with the expected behavior', async () => {
    wrapper2 = shallowMount(Dashboard, { // wrapper without devices
      store: storeNoDevices,
      localVue,
      stubs: ['fragment'],
    });

    expect(wrapper2.vm.hasDevices()).toBe(false);
    await wrapper2.vm.showScreenWelcome().then(() => {
      expect(wrapper2.vm.show).toBe(true);
    });
    localStorage.setItem('namespacesWelcome', JSON.stringify({ ...{ [namespace2.tenant_id]: true } }));

    await wrapper2.vm.showScreenWelcome().then(() => {
      expect(wrapper2.vm.show).toBe(false);
    });

    expect(wrapper.vm.namespaceHasBeenShown(namespace1.tenant_id)).toBe(false);

    localStorage.setItem('namespacesWelcome', JSON.stringify({
      ...JSON.parse(localStorage.getItem('namespacesWelcome')),
      ...{ [namespace1.tenant_id]: true },
    }));

    expect(wrapper.vm.namespaceHasBeenShown(namespace1.tenant_id)).toBe(true);

    await wrapper.vm.showScreenWelcome(); // wrapper having devices
    expect(wrapper.vm.hasDevices()).toBe(true);
    expect(wrapper.vm.show).toBe(false);
    expect(Object.keys(JSON.parse(localStorage.getItem('namespacesWelcome')))).toHaveLength(2);
  });
});
