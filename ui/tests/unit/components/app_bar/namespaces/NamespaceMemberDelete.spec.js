import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import Vuetify from 'vuetify';
import NamespaceMemberDelete from '@/components/app_bar/namespace/NamespaceMemberDelete';
import { actions, authorizer } from '../../../../../src/authorizer';
import '@/vee-validate';

describe('NamespaceMemberDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const accessType = ['owner', 'operator'];

  const hasAuthorization = {
    owner: true,
    operator: false,
  };

  const members = [
    {
      id: 'xxxxxxxx',
      type: 'owner',
      username: 'user1',
    },
    {
      id: 'xxxxxxxy',
      type: 'observer',
      username: 'user2',
    },
  ];

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
      description: 'Button',
      variables: {
        namespace: namespaceGlobal,
        dialog: false,
      },
      props: {
        member: members[0],
      },
      data: {
        dialog: false,
        action: 'removeMember',
      },
      template: {
        'removeMember-btn': true,
        'namespaceMemberDelete-dialog': false,
        'close-btn': false,
        'remove-btn': false,
      },
    },
    {
      description: 'Dialog',
      variables: {
        namespace: namespaceGlobal,
        dialog: true,
      },
      props: {
        member: members[0],
      },
      data: {
        dialog: true,
        action: 'removeMember',
      },
      template: {
        'removeMember-btn': true,
        'namespaceMemberDelete-dialog': true,
        'close-btn': true,
        'remove-btn': true,
      },
    },
  ];

  const storeVuex = (namespace, currentAccessType, tenant) => new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      currentAccessType,
      tenant,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
      'auth/accessType': (state) => state.currentAccessType,
      'auth/tenant': (state) => state.tenant,
    },
    actions: {
      'namespaces/removeUser': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    accessType.forEach((currentAccessType) => {
      describe(`${test.description} ${currentAccessType}`, () => {
        beforeEach(() => {
          wrapper = mount(NamespaceMemberDelete, {
            store: storeVuex(
              test.variables.namespace,
              currentAccessType,
              test.variables.namespace.tenant_id,
            ),
            localVue,
            stubs: ['fragment'],
            propsData: { member: test.props.member },
            vuetify,
            mocks: {
              $authorizer: authorizer,
              $actions: actions,
            },
          });

          wrapper.setData({ dialog: test.variables.dialog });
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

        it('Receive data in props', () => {
          Object.keys(test.props).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.props[item]);
          });
        });
        it('Compare data with default value', () => {
          Object.keys(test.data).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.data[item]);
          });
        });
        it('Process data in the computed', () => {
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentAccessType]);
        });

        //////
        // HTML validation
        //////

        it('Renders the template with data', () => {
          Object.keys(test.template).forEach((item) => {
            expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
          });
        });
      });
    });
  });
});
