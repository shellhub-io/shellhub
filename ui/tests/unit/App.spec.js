import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import Router from 'vue-router';
import App from '@/App';

describe('App', () => {
  const localVue = createLocalVue();
  const router = new Router();
  localVue.use(Vuex);
  localVue.use(Router);
  const vuetify = new Vuetify();

  let wrapper;

  const tenant = 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx';
  const isLoggedIn = true;
  const isMobile = false;
  const hasSpinner = false;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      tenant,
      isLoggedIn,
      isMobile,
      hasSpinner,
    },
    getters: {
      'auth/tenant': (state) => state.tenant,
      'auth/isLoggedIn': (state) => state.isLoggedIn,
      'mobile/isMobile': (state) => state.isMobile,
      'spinner/getStatus': (state) => state.hasSpinner,
    },
    actions: {
      'auth/logout': () => {},
      'privatekeys/fetch': () => {},
      'mobile/setIsMobileStatus': () => {},
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(App, {
      store,
      localVue,
      stubs: ['fragment'],
      mocks: {
        $env: (isEnterprise) => isEnterprise,
      },
      router,
      vuetify,
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Renders the template with components', async () => {
    expect(wrapper.find('[data-test="deviceWarning-component"]').exists()).toEqual(true);
  });
  it('Renders the template with data', async () => {
    expect(wrapper.find('[data-test="dashboard"]').exists()).toEqual(true);
    expect(wrapper.find('[data-test="devices"]').exists()).toEqual(true);
    expect(wrapper.find('[data-test="history"]').exists()).toEqual(true);
  });
});
