import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Notification from '@/components/app_bar/notification/Notification';

describe('Notification', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const numberNotifications = 2;
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

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      notifications,
      numberNotifications,
    },
    getters: {
      'notifications/list': (state) => state.notifications,
      'notifications/getNumberNotifications': (state) => state.numberNotifications,
    },
    actions: {
      'notifications/fetch': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(Notification, {
      store,
      localVue,
      stubs: ['fragment', 'router-link'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Renders the template with data', () => {
    Object.keys(notifications).forEach((field) => {
      expect(wrapper.find(`[data-test="${notifications[field].uid}-field"]`).text()).toEqual(notifications[field].name);
      expect(wrapper.find(`[data-test="${notifications[field].uid}-btn"]`).exists()).toEqual(true);
    });
  });
});
