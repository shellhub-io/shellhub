import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import App from '@/App';
import router from '@/router/index';

describe('App', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  let wrapper;

  const tenant = 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx';
  const isLoggedIn = true;
  const isMobile = false;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      tenant,
      isLoggedIn,
      isMobile,
    },
    getters: {
      'auth/tenant': (state) => state.tenant,
      'auth/isLoggedIn': (state) => state.isLoggedIn,
      'mobile/isMobile': (state) => state.isMobile,
    },
    actions: {
      'auth/logout': () => {
      },
      'privatekeys/fetch': () => {
      },
      'mobile/setIsMobileStatus': () => {
      },
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
  it('Renders the template with data', async () => {
    expect(wrapper.find('[data-test="dashboard"]').exists()).toEqual(true);
    expect(wrapper.find('[data-test="devices"]').exists()).toEqual(true);
    expect(wrapper.find('[data-test="history"]').exists()).toEqual(true);
  });
});
