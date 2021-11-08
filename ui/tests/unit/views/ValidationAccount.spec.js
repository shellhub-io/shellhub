import Vuex from 'vuex';
import { shallowMount, createLocalVue, config } from '@vue/test-utils';
import Vuetify from 'vuetify';
import Router from 'vue-router';
import ValidationAccount from '@/views/ValidationAccount';

config.mocks = {
  $success: {
    validationAccount: 'validation account',
  },
  $errors: {
    snackbar: {
      validationAccount: 'validation account',
    },
  },
};

describe('ValidationAccount', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  const router = new Router();
  localVue.use(Vuex);
  localVue.use(Router);

  let wrapper;

  const actionsValidationSuccess = {
    'users/validationAccount': () => {},
    'snackbar/showSnackbarSuccessAction': () => {},
    'snackbar/showSnackbarErrorAction': () => {},
  };

  const actionsValidationFalied = {
    'users/validationAccount': () => {
      throw new TypeError();
    },
    'snackbar/showSnackbarSuccessAction': () => {},
    'snackbar/showSnackbarErrorAction': () => {},
  };

  const storeValidationSuccess = new Vuex.Store({
    state: {
    },
    getters: {
    },
    actions: actionsValidationSuccess,
  });

  const storeValidationFalied = new Vuex.Store({
    state: {
    },
    getters: {
    },
    actions: actionsValidationFalied,
  });

  describe('Success to verify account', () => {
    beforeEach(() => {
      wrapper = shallowMount(ValidationAccount, {
        store: storeValidationSuccess,
        stubs: ['fragment'],
        localVue,
        router,
        vuetify,
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
      expect(wrapper.vm.activationProcessingStatus).toEqual('success');
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="processing-cardText"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="success-cardText"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="failed-cardText"]').exists()).toEqual(false);
    });
  });

  describe('Failed to verify account', () => {
    beforeEach(() => {
      wrapper = shallowMount(ValidationAccount, {
        store: storeValidationFalied,
        stubs: ['fragment'],
        localVue,
        router,
        vuetify,
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
      expect(wrapper.vm.activationProcessingStatus).toEqual('failed');
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="processing-cardText"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="success-cardText"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="failed-cardText"]').exists()).toEqual(true);
    });
  });
});
