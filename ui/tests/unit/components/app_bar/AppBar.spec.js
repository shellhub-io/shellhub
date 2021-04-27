import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import AppBar from '@/components/app_bar/AppBar';
import router from '@/router/index';

describe('AppBar', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const tenant = 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx';
  const isLoggedIn = true;
  const numberNamespaces = 1;
  const drawer = true;
  const isMobile = false;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isLoggedIn,
      numberNamespaces,
      tenant,
      isMobile,
    },
    getters: {
      'auth/isLoggedIn': (state) => state.isLoggedIn,
      'namespaces/getNumberNamespaces': (state) => state.numberNamespaces,
      'auth/tenant': (state) => state.tenant,
      'mobile/isMobile': (state) => state.isMobile,
    },
    actions: {
      'auth/logout': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(AppBar, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { drawer },
      router,
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Renders the template with data', async () => {
    expect(wrapper.find('[data-test="Settings"]').exists()).toEqual(true);
    expect(wrapper.find('[data-test="Logout"]').exists()).toEqual(true);
  });
});
