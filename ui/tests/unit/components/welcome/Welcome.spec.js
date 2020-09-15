import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Welcome from '@/components/welcome/Welcome';

describe('Welcome', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const show = true;
  const tenant = '';
  const stats = '';
  const devicePending = '';

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
      'stats/get': () => {
      },
      'devices/accept': () => {
      },
      'notifications/fetch': () => {
      },
      'modals/showSnackbarErrorAction': () => {
      },
      'modals/showSnackbarErrorDefault': () => {
      },
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

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
