import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import App from '@/App';
import router from '@/router/index';

import Vuetify from 'vuetify';

describe('App', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  let wrapper;

  const tenant = 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx';
  const isLoggedIn = true;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      tenant,
      isLoggedIn,
    },
    getters: {
      'auth/tenant': (state) => state.tenant,
      'auth/isLoggedIn': (state) => state.isLoggedIn,
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
