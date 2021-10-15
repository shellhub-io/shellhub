import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import NamespaceDelete from '@/components/app_bar/namespace/NamespaceDelete';

describe('NamespaceDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const tenant = 'xxxxxx';
  const owner = true;

  const namespace = {
    name: 'namespace3',
    owner: 'user1',
    member_names: ['user6', 'user7', 'user8'],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484715',
  };

  const inactiveBilling = {
    active: false,
    current_period_end: 0,
    customer_id: '',
    subscription_id: '',
    payment_method_id: '',
  };

  const activeBilling = {
    active: true,
    current_period_end: 12121,
    customer_id: 'cus_123',
    subscription_id: 'subs_123',
    payment_method_id: 'pm_123',
  };

  const text = `This action cannot be undone. This will permanently delete the
           ${namespace.name} and its related data.`;

  const mockCallWithout = jest.fn();
  const mockCallWith = jest.fn();

  const storeOwnerWithoutSubscription = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      stateBilling: inactiveBilling.active,
      billing: inactiveBilling,
      owner,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
      'billing/active': (state) => state.stateBilling,
      'billing/get': (state) => state.billing,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'namespaces/remove': () => {},
      'auth/logout': () => {},
      'billing/getSubscription': mockCallWithout,
      'snackbar/showSnackbarErrorLoading': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  const storeOwnerWithSubscription = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      stateBilling: activeBilling.active,
      billing: activeBilling,
      owner,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
      'billing/active': (state) => state.stateBilling,
      'billing/get': (state) => state.billing,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'namespaces/remove': () => {},
      'auth/logout': () => {},
      'billing/getSubscription': mockCallWith,
      'snackbar/showSnackbarErrorLoading': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  ///////
  // In this case, the rendering of the dialog is checked. In which
  // case with the input data it cannot take place.
  ///////

  describe('Button', () => {
    beforeEach(() => {
      wrapper = mount(NamespaceDelete, {
        store: storeOwnerWithoutSubscription,
        localVue,
        stubs: ['fragment'],
        propsData: { nsTenant: tenant },
        vuetify,
        mocks: {
          $env: {
            billingEnable: false,
          },
          $stripe: {
            elements: () => ({
              create: () => ({
                mount: () => null,
              }),
            }),
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

    ///////
    // Data and Props checking
    //////

    it('Receives data in props', () => {
      expect(wrapper.vm.nsTenant).toEqual(tenant);
    });
    it('Compare data with the default', () => {
      expect(wrapper.vm.name).toEqual(namespace.name);
      expect(wrapper.vm.dialog).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      expect(wrapper.find('[data-test="delete-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="namespaceDelete-dialog"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="contentSubscription-p"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="close-btn"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="remove-btn"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, when the user owns the keys and the focus of
  // the test is dialog rendering.
  ///////

  describe('Dialog without subscription', () => {
    beforeEach(() => {
      wrapper = mount(NamespaceDelete, {
        store: storeOwnerWithoutSubscription,
        localVue,
        stubs: ['fragment'],
        propsData: { nsTenant: tenant },
        vuetify,
        mocks: {
          $env: {
            billingEnable: false,
          },
          $stripe: {
            elements: () => ({
              create: () => ({
                mount: () => null,
              }),
            }),
          },
        },
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

    ///////
    // Data and Props checking
    //////

    it('Receives data in props', () => {
      expect(wrapper.vm.nsTenant).toEqual(tenant);
    });
    it('Compare data with the default', () => {
      expect(wrapper.vm.name).toEqual(namespace.name);
      expect(wrapper.vm.dialog).toEqual(true);
    });

    //////
    // Call actions
    //////

    it('Call actions', () => {
      expect(mockCallWithout).not.toHaveBeenCalled();
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="delete-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="namespaceDelete-dialog"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="contentSubscription-p"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="close-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="remove-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="content-text"]').text()).toEqual(text);
    });
  });

  ///////
  // In this case, when the user owns the keys and the focus of
  // the test is dialog rendering.
  ///////

  describe('Dialog with subscription', () => {
    beforeEach(() => {
      wrapper = mount(NamespaceDelete, {
        store: storeOwnerWithSubscription,
        localVue,
        stubs: ['fragment'],
        propsData: { nsTenant: tenant },
        vuetify,
        mocks: {
          $env: {
            billingEnable: true,
          },
          $stripe: {
            elements: () => ({
              create: () => ({
                mount: () => null,
              }),
            }),
          },
        },
      });

      wrapper.setData({ dialog: true, amountDue: 5 });
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

    it('Receives data in props', () => {
      expect(wrapper.vm.nsTenant).toEqual(tenant);
    });
    it('Compare data with the default', () => {
      expect(wrapper.vm.name).toEqual(namespace.name);
      expect(wrapper.vm.dialog).toEqual(true);
    });

    //////
    // Call actions
    //////

    it('Call actions', () => {
      expect(mockCallWith).toHaveBeenCalled();
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="delete-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="namespaceDelete-dialog"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="contentSubscription-p"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="close-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="remove-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="content-text"]').text()).toEqual(text);
    });
  });
});
