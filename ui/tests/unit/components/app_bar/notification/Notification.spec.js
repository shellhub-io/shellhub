import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Notification from '@/components/app_bar/notification/Notification';

describe('Notification', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      notifications: [],
      numberNotifications: 0,
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
      stubs: ['fragment'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
