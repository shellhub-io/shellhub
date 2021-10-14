import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Notification from '@/components/app_bar/notification/Notification';

describe('Notification', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const inANamespace = false;
  const owner = true;
  const numberNotifications = 2;
  const noNotifications = Array(0);

  const statsWithNotification = {
    registered_devices: 0,
    online_devices: 0,
    active_sessions: 0,
    pending_devices: 2,
    rejected_devices: 0,
  };

  const statsWithoutNotification = { ...statsWithNotification, pending_devices: 0 };

  const notifications = [
    {
      uid: 'a582b47a42d',
      name: '39-5e-2a',
      identity: {
        mac: '00:00:00:00:00:00',
      },
      info: {
        id: 'linuxmint',
        pretty_name: 'Linux Mint 19.3',
        version: '',
      },
      public_key: '----- PUBLIC KEY -----',
      tenant_id: '00000000',
      last_seen: '2020-05-20T18:58:53.276Z',
      online: false,
      namespace: 'user',
      status: 'pending',
    },
    {
      uid: 'a582b47a42e',
      name: '39-5e-2b',
      identity: {
        mac: '00:00:00:00:00:00',
      },
      info: {
        id: 'linuxmint',
        pretty_name: 'Linux Mint 19.3',
        version: '',
      },
      public_key: '----- PUBLIC KEY -----',
      tenant_id: '00000001',
      last_seen: '2020-05-20T19:58:53.276Z',
      online: true,
      namespace: 'user',
      status: 'pending',
    },
  ];

  const storeNotOwner = new Vuex.Store({
    namespaced: true,
    state: {
      notifications,
      numberNotifications,
      owner: !owner,
      stats: statsWithNotification,
    },
    getters: {
      'notifications/list': (state) => state.notifications,
      'notifications/getNumberNotifications': (state) => state.numberNotifications,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'notifications/fetch': () => {},
      'stats/get': () => {},
    },
  });

  const storeOwner = new Vuex.Store({
    namespaced: true,
    state: {
      notifications,
      numberNotifications,
      owner,
      stats: statsWithNotification,
    },
    getters: {
      'notifications/list': (state) => state.notifications,
      'notifications/getNumberNotifications': (state) => state.numberNotifications,
      'namespaces/owner': (state) => state.owner,
      'stats/stats': (state) => state.stats,
    },
    actions: {
      'notifications/fetch': () => {},
      'stats/get': () => {},
    },
  });

  const storeNoNotifications = new Vuex.Store({
    namespaced: true,
    state: {
      notifications: noNotifications,
      numberNotifications: 0,
      owner,
      stats: statsWithoutNotification,
    },
    getters: {
      'notifications/list': (state) => state.notifications,
      'notifications/getNumberNotifications': (state) => state.numberNotifications,
      'namespaces/owner': (state) => state.owner,
      'stats/stats': (state) => state.stats,
    },
    actions: {
      'notifications/fetch': () => {},
      'stats/get': () => {},
    },
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is notifications rendering. And the button available
  // for the user to accept the pending device.
  ///////

  describe('Owner with notifications', () => {
    beforeEach(() => {
      wrapper = shallowMount(Notification, {
        store: storeOwner,
        localVue,
        stubs: ['fragment', 'router-link'],
        propsData: { inANamespace },
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
      expect(wrapper.vm.listNotifications).toEqual([]);
      expect(wrapper.vm.numberNotifications).toEqual(0);
      expect(wrapper.vm.shown).toEqual(false);
      expect(wrapper.vm.inANamespace).toEqual(false);
      expect(wrapper.vm.defaultSize).toEqual(24);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.getListNotifications).toEqual(notifications);
      expect(wrapper.vm.getNumberNotifications).toEqual(numberNotifications);
      expect(wrapper.vm.getStatusNotifications).toEqual(false);
      expect(wrapper.vm.isOwner).toEqual(owner);
      expect(wrapper.vm.hasNamespace).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      Object.keys(notifications).forEach((field) => {
        expect(wrapper.find(`[data-test="${notifications[field].uid}-field"]`).text()).toEqual(notifications[field].name);
        expect(wrapper.find(`[data-test="${notifications[field].uid}-btn"]`).exists()).toEqual(true);
      });
    });
  });

  ///////
  // In this case, when the user doesn't own the namespace and the
  // focus of the test is notifications rendering. And the button
  // unavailable for the user to accept the pending device.
  ///////

  describe('Doesn\'t own the namespace with notifications', () => {
    beforeEach(() => {
      wrapper = shallowMount(Notification, {
        store: storeNotOwner,
        localVue,
        stubs: ['fragment', 'router-link'],
        propsData: { inANamespace },
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
      expect(wrapper.vm.listNotifications).toEqual([]);
      expect(wrapper.vm.numberNotifications).toEqual(0);
      expect(wrapper.vm.shown).toEqual(false);
      expect(wrapper.vm.inANamespace).toEqual(false);
      expect(wrapper.vm.defaultSize).toEqual(24);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.getListNotifications).toEqual(notifications);
      expect(wrapper.vm.getNumberNotifications).toEqual(numberNotifications);
      expect(wrapper.vm.getStatusNotifications).toEqual(false);
      expect(wrapper.vm.isOwner).toEqual(!owner);
      expect(wrapper.vm.hasNamespace).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      Object.keys(notifications).forEach((field) => {
        expect(wrapper.find(`[data-test="${notifications[field].uid}-field"]`).text()).toEqual(notifications[field].name);
        expect(wrapper.find(`[data-test="${notifications[field].uid}-btn"]`).exists()).toEqual(false);
      });
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test rendering is the message that the user has no
  // notification.
  ///////

  describe('Owner without notifications', () => {
    beforeEach(() => {
      wrapper = shallowMount(Notification, {
        store: storeNoNotifications,
        localVue,
        stubs: ['fragment', 'router-link'],
        propsData: { inANamespace },
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
      expect(wrapper.vm.listNotifications).toEqual([]);
      expect(wrapper.vm.numberNotifications).toEqual(0);
      expect(wrapper.vm.shown).toEqual(false);
      expect(wrapper.vm.inANamespace).toEqual(false);
      expect(wrapper.vm.defaultSize).toEqual(24);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.getListNotifications).toEqual([]);
      expect(wrapper.vm.getNumberNotifications).toEqual(0);
      expect(wrapper.vm.getStatusNotifications).toEqual(true);
      expect(wrapper.vm.isOwner).toEqual(owner);
      expect(wrapper.vm.hasNamespace).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="noNotifications"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="noNotifications"]').text()).toEqual('You don\'t have notifications');
    });
  });
});
