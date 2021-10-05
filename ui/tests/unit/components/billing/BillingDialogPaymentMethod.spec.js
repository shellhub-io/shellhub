import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import BillingDialogPaymentMethod from '@/components/billing/BillingDialogPaymentMethod';

describe('BillingDialogPaymentMethod', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const typeOperation = 'subscription';
  const hasSpinner = false;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      hasSpinner,
    },
    getters: {
      'spinner/getStatus': (state) => state.hasSpinner,
    },
    actions: {
      'billing/subscritionPaymentMethod': () => {},
      'billing/updatePaymentMethod': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  ///////
  // In this case, it's testing the button rendering.
  ///////

  describe('Button', () => {
    beforeEach(() => {
      wrapper = mount(BillingDialogPaymentMethod, {
        localVue,
        store,
        stubs: ['fragment'],
        propsData: { typeOperation },
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

    /// //
    // Data and Props checking
    /// /

    it('Receive data in props', () => {
      expect(wrapper.vm.typeOperation).toBe(typeOperation);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.card).toEqual(null);
      expect(wrapper.vm.elementError).toEqual(null);
      expect(wrapper.vm.elements).toEqual(null);
      expect(wrapper.vm.lockButton).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      expect(wrapper.find('[data-test="show-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="BillingDialogPaymentMethod-dialog"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="confirm-btn"]').exists()).toBe(false);
    });
  });

  ///////
  // In this case, it's testing the subscription.
  ///////

  describe('Dialog Subscription', () => {
    beforeEach(() => {
      wrapper = mount(BillingDialogPaymentMethod, {
        localVue,
        store,
        stubs: ['fragment'],
        propsData: { typeOperation },
        vuetify,
      });

      wrapper.setData({ dialog: true });
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

    /// //
    // Data and Props checking
    /// /

    it('Receive data in props', () => {
      expect(wrapper.vm.typeOperation).toBe(typeOperation);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.card).toEqual(null);
      expect(wrapper.vm.elementError).toEqual(null);
      expect(wrapper.vm.elements).toEqual(null);
      expect(wrapper.vm.lockButton).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      expect(wrapper.find('[data-test="show-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="BillingDialogPaymentMethod-dialog"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="confirm-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="text-cardTitle"]').text()).toEqual('Create subscription');
    });
  });

  ///////
  // In this case, it's testing the update subscription.
  ///////

  describe('Dialog Update', () => {
    const typeOperationUpdate = 'update';

    beforeEach(() => {
      wrapper = mount(BillingDialogPaymentMethod, {
        localVue,
        store,
        stubs: ['fragment'],
        propsData: { typeOperation: typeOperationUpdate },
        vuetify,
      });

      wrapper.setData({ dialog: true });
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

    /// //
    // Data and Props checking
    /// /

    it('Receive data in props', () => {
      expect(wrapper.vm.typeOperation).toBe(typeOperationUpdate);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.card).toEqual(null);
      expect(wrapper.vm.elementError).toEqual(null);
      expect(wrapper.vm.elements).toEqual(null);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      expect(wrapper.find('[data-test="show-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="BillingDialogPaymentMethod-dialog"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="confirm-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="text-cardTitle"]').text()).toEqual('Update payment method');
    });
  });
});
