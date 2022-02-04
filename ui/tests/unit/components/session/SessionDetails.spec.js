import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import timezoneMock from 'timezone-mock';
import SessionDetails from '@/components/session/SessionDetails';

describe('SessionDetails', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const session = {
    uid: '8c354a00',
    device_uid: 'a582b47a',
    device: {
      uid: 'a582b47a',
      name: '39-5e-2a',
      identity: {
        mac: '00:00:00',
      },
      info: {
        id: 'debian',
        pretty_name: 'Debian',
        version: 'v0.2.5',
      },
      public_key: 'xxxxxxxx',
      tenant_id: '00000000',
      last_seen: '2020-05-18T13:27:02.498Z',
      online: true,
      namespace: 'user',
    },
    tenant_id: '00000000',
    username: 'user',
    ip_address: '00.00.00',
    started_at: '2020-05-18T12:30:28.824Z',
    last_seen: '2020-05-18T12:30:30.205Z',
    active: true,
    authenticated: true,
    recorded: true,
  };

  const storeRecordedTrue = new Vuex.Store({
    namespaced: true,
    state: {
      session,
    },
    getters: {
      'sessions/get': (state) => state.session,
    },
    actions: {
      'sessions/get': () => {},
      'sessions/close': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  const storeRecordedFalse = new Vuex.Store({
    namespaced: true,
    state: {
      session: { ...session, recorded: false },
    },
    getters: {
      'sessions/get': (state) => state.session,
    },
    actions: {
      'sessions/get': () => {},
      'sessions/close': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  const storeRecordedFalseAndOffline = new Vuex.Store({
    namespaced: true,
    state: {
      session: {
        ...session, device: { online: false }, active: false, recorded: false,
      },
    },
    getters: {
      'sessions/get': (state) => state.session,
    },
    actions: {
      'sessions/get': () => {},
      'sessions/close': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  ///////
  // In this case, it is checking the rendering of components when
  // the session has been recorded and the device is online.
  ///////

  describe('Recorded is true and device is online', () => {
    beforeEach(() => {
      timezoneMock.register('UTC');

      wrapper = shallowMount(SessionDetails, {
        store: storeRecordedTrue,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $route: {
            params: {
              id: session.uid,
            },
          },
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
    // Data and Props checking
    //////

    it('Compare data with default value', () => {
      expect(wrapper.vm.uid).toEqual(session.uid);
      expect(wrapper.vm.session).toEqual(session);
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.sessionPlayDialog).toEqual(false);
      expect(wrapper.vm.sessionCloseDialog).toEqual(false);
      expect(wrapper.vm.hide).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="sessionPlay-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="sessionClose-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="sessionDeleteRecord-component"]').exists()).toBe(true);
    });

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="sessionUid-field"]').text()).toEqual(session.uid);
      expect(wrapper.find('[data-test="sessionUser-field"]').text()).toEqual(session.username);
      expect(wrapper.find('[data-test="sessionIpAddress-field"]').text()).toEqual(session.ip_address);
      expect(wrapper.find('[data-test="sessionStartedAt-field"]').text()).toEqual('Monday, May 18th 2020, 12:30:28 pm');
      expect(wrapper.find('[data-test="sessionLastSeen-field"]').text()).toEqual('Monday, May 18th 2020, 12:30:30 pm');
    });
  });

  ///////
  // In this case, it is checking the rendering of components when
  // the session has not been recorded and the device is online.
  ///////

  describe('Recorded is false and device is online', () => {
    beforeEach(() => {
      timezoneMock.register('UTC');

      wrapper = shallowMount(SessionDetails, {
        store: storeRecordedFalse,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $route: {
            params: {
              id: session.uid,
            },
          },
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
    // Data and Props checking
    //////

    it('Compare data with default value', () => {
      expect(wrapper.vm.uid).toEqual(session.uid);
      expect(wrapper.vm.session).toEqual({ ...session, recorded: false });
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.sessionPlayDialog).toEqual(false);
      expect(wrapper.vm.sessionCloseDialog).toEqual(false);
      expect(wrapper.vm.hide).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="sessionPlay-component"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="sessionClose-component"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="sessionDeleteRecord-component"]').exists()).toBe(false);
    });

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="sessionUid-field"]').text()).toEqual(session.uid);
      expect(wrapper.find('[data-test="sessionUser-field"]').text()).toEqual(session.username);
      expect(wrapper.find('[data-test="sessionIpAddress-field"]').text()).toEqual(session.ip_address);
      expect(wrapper.find('[data-test="sessionStartedAt-field"]').text()).toEqual('Monday, May 18th 2020, 12:30:28 pm');
      expect(wrapper.find('[data-test="sessionLastSeen-field"]').text()).toEqual('Monday, May 18th 2020, 12:30:30 pm');
    });
  });

  ///////
  // In this case, it is checking the rendering of components when
  // the session has not been recorded and the device is offline.
  ///////

  describe('Recorded is false and device is offline', () => {
    beforeEach(() => {
      timezoneMock.register('UTC');

      wrapper = shallowMount(SessionDetails, {
        store: storeRecordedFalseAndOffline,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $route: {
            params: {
              id: session.uid,
            },
          },
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
    // Data and Props checking
    //////

    it('Receive data in props', () => {
      expect(wrapper.vm.uid).toEqual(session.uid);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.uid).toEqual(session.uid);
      expect(wrapper.vm.session).toEqual({
        ...session, device: { online: false }, active: false, recorded: false,
      });
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.sessionPlayDialog).toEqual(false);
      expect(wrapper.vm.sessionCloseDialog).toEqual(false);
      expect(wrapper.vm.hide).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="sessionPlay-component"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="sessionClose-component"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="sessionDeleteRecord-component"]').exists()).toBe(false);
    });

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="sessionUid-field"]').text()).toEqual(session.uid);
      expect(wrapper.find('[data-test="sessionUser-field"]').text()).toEqual(session.username);
      expect(wrapper.find('[data-test="sessionIpAddress-field"]').text()).toEqual(session.ip_address);
      expect(wrapper.find('[data-test="sessionStartedAt-field"]').text()).toEqual('Monday, May 18th 2020, 12:30:28 pm');
      expect(wrapper.find('[data-test="sessionLastSeen-field"]').text()).toEqual('Monday, May 18th 2020, 12:30:30 pm');
    });
  });
});
