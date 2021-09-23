import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import Router from 'vue-router';
import ValidationAccount from '@/views/ValidationAccount';

describe('ValidationAccount', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  const router = new Router();
  localVue.use(Vuex);
  localVue.use(Router);

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
      vuetify,
      mocks: {
        $success: {
          validationAccount: 'validation account',
        },
        $errors: {
          snackbar: {
            validationAccount: 'validation account',
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
});
