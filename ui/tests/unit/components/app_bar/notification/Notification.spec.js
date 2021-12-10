import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Notification from '@/components/app_bar/notification/Notification';
import { actions, authorizer } from '../../../../../src/authorizer';

describe('Notification', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const role = ['owner', 'operator'];

  const hasAuthorization = {
    owner: true,
    operator: false,
  };

  const statsWithoutNotification = {
    registered_devices: 0,
    online_devices: 0,
    active_sessions: 0,
    pending_devices: 0,
    rejected_devices: 0,
  };

  const statsWithNotification = { ...statsWithoutNotification, pending_devices: 2 };

  const notificationsGlobal = [
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

  const tests = [
    {
      description: 'Without notifications',
      variables: {
        listNotifications: [],
        numberNotifications: 0,
        stats: statsWithoutNotification,
      },
      data: {
        listNotifications: [],
        numberNotifications: 0,
        shown: false,
        inANamespace: false,
        defaultSize: 24,
        action: 'view',
      },
      computed: {
        getListNotifications: [],
        getNumberNotifications: 0,
        showNumberNotifications: 0,
        getStatusNotifications: true,
        hasNamespace: true,
      },
      template: {
        'notifications-badge': false,
        'hasNotifications-subheader': false,
        'show-btn': false,
        'noNotifications-subheader': true,
      },
    },
    {
      description: 'With notifications',
      variables: {
        listNotifications: notificationsGlobal,
        numberNotifications: 2,
        stats: statsWithNotification,
      },
      data: {
        listNotifications: [],
        numberNotifications: 0,
        shown: false,
        inANamespace: false,
        defaultSize: 24,
        action: 'view',
      },
      computed: {
        getListNotifications: notificationsGlobal,
        getNumberNotifications: 2,
        showNumberNotifications: 2,
        getStatusNotifications: false,
        hasNamespace: true,
      },
      template: {
        'notifications-badge': false,
        'hasNotifications-subheader': true,
        'show-btn': true,
        'noNotifications-subheader': false,
      },
    },
  ];

  const storeVuex = (
    notifications,
    numberNotifications,
    stats,
    currentrole,
  ) => new Vuex.Store({
    namespaced: true,
    state: {
      notifications,
      numberNotifications,
      stats,
      currentrole,
    },
    getters: {
      'notifications/list': (state) => state.notifications,
      'notifications/getNumberNotifications': (state) => state.numberNotifications,
      'stats/stats': (state) => state.stats,
      'auth/role': (state) => state.currentrole,
    },
    actions: {
      'notifications/fetch': () => {},
      'stats/get': () => {},
      'snackbar/showSnackbarErrorAssociation': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = shallowMount(Notification, {
            store: storeVuex(
              test.variables.listNotifications,
              test.variables.numberNotifications,
              test.variables.stats,
              currentrole,
            ),
            localVue,
            stubs: ['fragment', 'router-link'],
            mocks: {
              $authorizer: authorizer,
              $actions: actions,
            },
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
        // Data checking
        //////

        it('Compare data with default value', () => {
          Object.keys(test.data).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.data[item]);
          });
        });
        it('Process data in the computed', () => {
          Object.keys(test.computed).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.computed[item]);
          });
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentrole]);
        });

        //////
        // HTML validation
        //////

        it('Renders the template with data', () => {
          Object.keys(test.template).forEach((item) => {
            expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
          });
        });
      });
    });
  });
});
