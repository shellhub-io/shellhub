import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import AccountCreated from '@/components/account/AccountCreated';

describe('Account Created', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const show = false;
  const email = 'email@email.com';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
    },
    getters: {
    },
    actions: {
      'users/resendEmail': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  describe('Doesn\'t render component', () => {
    beforeEach(() => {
      wrapper = shallowMount(AccountCreated, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { show, email },
      });
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      document.body.setAttribute('data-app', true);
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Receive data in props', () => {
      expect(wrapper.vm.show).toEqual(show);
      expect(wrapper.vm.email).toEqual(email);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="accountCreated-card"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="resendEmail-btn"]').exists()).toEqual(false);
    });
  });

  describe('Render component', () => {
    beforeEach(() => {
      wrapper = shallowMount(AccountCreated, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { show: !show, email },
      });
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      document.body.setAttribute('data-app', true);
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Receive data in props', () => {
      expect(wrapper.vm.show).toEqual(!show);
      expect(wrapper.vm.email).toEqual(email);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="accountCreated-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="resendEmail-btn"]').exists()).toEqual(true);
    });
  });
});
