import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import Vuetify from 'vuetify';
import NamespaceMemberFormDialogEdit from '@/components/namespace/NamespaceMemberFormDialogEdit';
import { actions, authorizer } from '../../../../src/authorizer';
import '@/vee-validate';

describe('NamespaceMemberFormDialogEdit', () => {
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
      description: 'Dialog closed',
      variables: {
        namespaceGlobal,
      },
      props: {
        member: members[1],
        show: false,
      },
      data: {
        username: '',
        selectedRole: '',
        memberLocal: memberLocalEdit,
        items: ['administrator', 'operator', 'observer'],
      },
      template: {
        'remove-icon': true,
        'edit-title': true,
        'namespaceNewMember-dialog': false,
      },
      templateText: {
        'edit-title': 'Edit',
      },
    },
    {
      description: 'Dialog opened',
      variables: {
        namespaceGlobal,
      },
      props: {
        member: members[1],
        show: true,
      },
      data: {
        username: '',
        selectedRole: '',
        memberLocal: memberLocalEdit,
        items: ['administrator', 'operator', 'observer'],
      },
      template: {
        'remove-icon': true,
        'edit-title': true,
        'namespaceNewMember-dialog': true,
        'text-title': true,
        'close-btn': true,
        'edit-btn': true,
      },
      templateText: {
        'edit-title': 'Edit',
        'text-title': 'Update member role',
        'close-btn': 'Close',
        'edit-btn': 'Edit',
      },
    },
  ];

  const storeVuex = (namespace) => new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
    },
    actions: {
      'namespaces/adduser': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(NamespaceMemberFormDialogEdit, {
          store: storeVuex(test.variables.namespaceGlobal),
          localVue,
          stubs: ['fragment'],
          propsData: {
            member: test.props.member,
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

      //////
      // HTML validation
      //////

      it('Renders the template with data', () => {
        Object.keys(test.template).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
        });
      });
      it('Renders template with expected text', () => {
        Object.keys(test.templateText).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).text()).toContain(test.templateText[item]);
        });
      });
    });
  });
});
