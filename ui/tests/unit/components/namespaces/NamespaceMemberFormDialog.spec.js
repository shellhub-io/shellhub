import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import NamespaceMemberFormDialog from '@/components/namespace/NamespaceMemberFormDialog';
import { actions, authorizer } from '../../../../src/authorizer';
import '@/vee-validate';

describe('NamespaceMemberFormDialog', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const role = ['owner', 'administrator', 'operator', 'observer'];

  const hasAuthorization = {
    owner: true,
    administrator: true,
    operator: false,
    observer: false,
  };

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

  const memberLocal = {
    id: '',
    selectedRole: '',
    username: '',
  };

  const memberLocalEdit = {
    id: 'xxxxxxxy',
    role: 'observer',
    selectedRole: 'observer',
    username: 'user2',
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
      description: 'Icon add user',
      variables: {
        namespaceGlobal,
      },
      props: {
        member: {},
        addUser: true,
        show: false,
      },
      data: {
        username: '',
        selectedRole: '',
        memberLocal,
        items: ['administrator', 'operator', 'observer'],
      },
      template: {
        'add-btn': true,
        'edit-list': false,
        'edit-icon': false,
        'namespaceNewMember-dialog': false,
        'dialogClose-btn': false,
        'dialogAdd-btn': false,
        'dialogEdit-btn': false,
      },
    },
    {
      description: 'Icon edit user',
      variables: {
        namespaceGlobal,
      },
      props: {
        member: members[1],
        addUser: false,
        show: false,
      },
      data: {
        username: '',
        selectedRole: '',
        memberLocal: memberLocalEdit,
        items: ['administrator', 'operator', 'observer'],
      },
      template: {
        'add-btn': false,
        'edit-list': true,
        'edit-icon': true,
        'namespaceNewMember-dialog': false,
        'dialogClose-btn': false,
        'dialogAdd-btn': false,
        'dialogEdit-btn': false,
      },
    },
    {
      description: 'Dialog create user',
      variables: {
        namespaceGlobal,
      },
      props: {
        member: {},
        addUser: true,
        show: true,
      },
      data: {
        username: '',
        selectedRole: '',
        memberLocal,
        items: ['administrator', 'operator', 'observer'],
      },
      template: {
        'add-btn': true,
        'edit-list': false,
        'edit-icon': false,
        'namespaceNewMember-dialog': true,
        'dialogClose-btn': true,
        'dialogAdd-btn': true,
        'dialogEdit-btn': false,
      },
    },
    {
      description: 'Dialog edit user',
      variables: {
        namespaceGlobal,
        dialog: true,
      },
      props: {
        member: members[1],
        addUser: false,
        show: true,
      },
      data: {
        username: '',
        selectedRole: '',
        memberLocal: memberLocalEdit,
        items: ['administrator', 'operator', 'observer'],
      },
      template: {
        'add-btn': false,
        'edit-list': true,
        'edit-icon': true,
        'namespaceNewMember-dialog': true,
        'dialogClose-btn': true,
        'dialogAdd-btn': false,
        'dialogEdit-btn': true,
      },
    },
  ];

  const storeVuex = (namespace, currentrole) => new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      currentrole,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
      'auth/role': (state) => state.currentrole,
    },
    actions: {
      'namespaces/adduser': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = mount(NamespaceMemberFormDialog, {
            store: storeVuex(test.variables.namespaceGlobal, currentrole),
            localVue,
            stubs: ['fragment'],
            propsData: {
              member: test.props.member,
              addUser: test.props.addUser,
              show: test.props.show,
            },
            vuetify,
            mocks: {
              $authorizer: authorizer,
              $actions: actions,
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
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentrole]);
        });

        //////
        // HTML validation
        //////

        it('Renders the template with data', () => {
          if (hasAuthorization[currentrole]) {
            Object.keys(test.template).forEach((item) => {
              expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
            });
          } else if (!test.props.show) {
            Object.keys(test.template).forEach((item) => {
              expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
            });
          }
        });

        if (test.props.show) {
          if (hasAuthorization[currentrole] && !test.props.addUser) {
            it('Show validation messages', async () => {
              wrapper.setData({ memberLocal: { username: '' } });
              await flushPromises();

              const validator = wrapper.vm.$refs.providerUsername;

              await validator.validate();
              expect(validator.errors[0]).toBe('This field is required');
            });
          }
        }
      });
    });
  });
});
