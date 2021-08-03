import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import NamespaceNewMember from '@/components/app_bar/namespace/NamespaceNewMember';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import '@/vee-validate';

describe('NamespaceNewMember', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  document.body.setAttribute('data-app', true);
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  let wrapper;

  const tenant = 'xxxxx';

  const store = new Vuex.Store({
    namespaced: true,
    actions: {
      'namespaces/adduser': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
      'snackbar/showSnackbarErrorAction': () => {
      },
    },
  });

  ///////
  // In this case, the rendering of the button is checked.
  ///////

  describe('Button', () => {
    beforeEach(() => {
      wrapper = mount(NamespaceNewMember, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { nsTenant: tenant },
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

    it('Receive data in props', () => {
      expect(wrapper.vm.nsTenant).toEqual(tenant);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.username).toEqual('');
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.tenant).toEqual(tenant);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      expect(wrapper.find('[data-test="add-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="namespaceNewMember-dialog"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, the focus of the test is dialog rendering.
  ///////

  describe('Button', () => {
    beforeEach(() => {
      wrapper = mount(NamespaceNewMember, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { nsTenant: tenant },
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

    ///////
    // Data and Props checking
    //////

    it('Receive data in props', () => {
      expect(wrapper.vm.nsTenant).toEqual(tenant);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.username).toEqual('');
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.tenant).toEqual(tenant);
    });

    //////
    // HTML validation
    //////

    it('Show validation messages', async () => {
      wrapper.setData({ namespaceName: '' });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerUsername;

      await validator.validate();
      expect(validator.errors[0]).toBe('This field is required');
    });
    it('Renders the template with data', async () => {
      expect(wrapper.find('[data-test="add-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="namespaceNewMember-dialog"]').exists()).toEqual(true);
    });
  });
});
