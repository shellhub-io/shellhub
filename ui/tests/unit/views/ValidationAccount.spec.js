import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import ValidationAccount from '@/views/ValidationAccount';
import router from '@/router/index';

describe('ValidationAccount', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const store = new Vuex.Store({
    state: {
    },
    getters: {
    },
    actions: {
      'users/validationAccount': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(ValidationAccount, {
      store,
      stubs: ['fragment'],
      localVue,
      router,
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
});
