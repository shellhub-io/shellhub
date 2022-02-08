import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import NamespaceMemberFormDialogAdd from '@/components/namespace/NamespaceMemberFormDialogAdd';
import { actions, authorizer } from '../../../../src/authorizer';
import '@/vee-validate';

describe('NamespaceMemberFormDialogAdd', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const members = [
    {
      id: 'xxxxxxxx',
      role: 'owner',
      username: 'user1',
    },
    {
      id: 'xxxxxxxy',
      role: 'observer',
      username: 'user2',
    },
  ];

  const member = {
    selectedRole: '',
    username: '',
  };

  const namespaceGlobal = {
    name: 'namespace',
    owner: 'user1',
    members,
    tenant_id: 'xxxxxxxx',
    devices_count: 0,
    max_devices: 0,
  };

  const tests = [
    {
      description: 'Button add user has authorization',
      variables: {
        namespaceGlobal,
      },
      role: {
        type: 'owner',
        permission: true,
      },
      data: {
        username: '',
        selectedRole: '',
        action: 'addMember',
        dialog: false,
        member,
        items: ['administrator', 'operator', 'observer'],
      },
      computed: {
        hasAuthorization: true,
      },
      template: {
        'addMember-btn': true,
        'namespaceNewMember-dialog': false,
        'close-btn': false,
        'add-btn': false,
      },
    },
    {
      description: 'Button add user has no authorization',
      variables: {
        namespaceGlobal,
      },
      role: {
        type: 'operator',
        permission: false,
      },
      data: {
        username: '',
        selectedRole: '',
        action: 'addMember',
        dialog: false,
        member,
        items: ['administrator', 'operator', 'observer'],
      },
      computed: {
        hasAuthorization: false,
      },
      template: {
        'addMember-btn': true,
        'namespaceNewMember-dialog': false,
        'close-btn': false,
        'add-btn': false,
      },
    },
    {
      description: 'dialog add user has authorization',
      variables: {
        namespaceGlobal,
      },
      role: {
        type: 'owner',
        permission: true,
      },
      data: {
        username: '',
        selectedRole: '',
        action: 'addMember',
        dialog: true,
        member,
        items: ['administrator', 'operator', 'observer'],
      },
      computed: {
        hasAuthorization: true,
      },
      template: {
        'addMember-btn': true,
        'namespaceNewMember-dialog': true,
        'close-btn': true,
        'add-btn': true,
      },
    },
  ];

  const storeVuex = (namespace, currentRole) => new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      currentRole,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
      'auth/role': (state) => state.currentRole,
    },
    actions: {
      'namespaces/adduser': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description} - ${test.role.type}`, () => {
      beforeEach(() => {
        wrapper = mount(NamespaceMemberFormDialogAdd, {
          store: storeVuex(
            test.variables.namespaceGlobal,
            test.role.type,
          ),
          localVue,
          stubs: ['fragment'],
          vuetify,
          mocks: {
            $authorizer: authorizer,
            $actions: actions,
          },
        });

        wrapper.setData({ dialog: test.data.dialog });
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
      // Data checking
      //////

      it('Compare data with default value', () => {
        Object.keys(test.data).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.data[item]);
        });
      });
      it('Process data in the computed', () => {
        Object.keys(test.computed).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.computed[item]);
        });
      });

      //////
      // HTML validation
      //////

      it('Renders the template with data', () => {
        Object.keys(test.template).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
        });
      });

      if (test.role.permission && test.data.show) {
        it('Show validation messages', async () => {
          wrapper.setData({ memberLocal: { username: '' } });
          await flushPromises();

          const validator = wrapper.vm.$refs.providerUsername;

          await validator.validate();
          expect(validator.errors[0]).toBe('This field is required');
        });
      }
    });
  });
});
