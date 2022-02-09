import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import NamespaceMemberList from '@/components/namespace/NamespaceMemberList';
import { actions, authorizer } from '../../../../src/authorizer';

describe('NamespaceMemberList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const tenantGlobal = 'xxxxxxxx';

  const namespace = {
    name: 'nsxxx',
    members: [{ username: 'user1', role: 'owner' }, { username: 'user2', role: 'administrator' }, { username: 'user3', role: 'observer' }],
  };

  const headers = [
    {
      text: 'Username',
      value: 'username',
      align: 'start',
      sortable: false,
    },
    {
      text: 'Role',
      value: 'role',
      align: 'center',
      sortable: false,
    },
    {
      text: 'Actions',
      value: 'actions',
      align: 'end',
      sortable: false,
    },
  ];

  const tests = [
    {
      description: 'List data when user has owner role',
      variables: {
        tenant: tenantGlobal,
      },
      role: {
        type: 'owner',
        permission: true,
      },
      props: {
        namespace,
      },
      data: {
        menu: false,
        namespaceMemberFormShow: [],
        namespaceMemberDeleteShow: [],
        editMemberAction: 'editMember',
        removeMemberAction: 'removeMember',
        headers,
      },
      computed: {
        tenant: tenantGlobal,
        members: namespace.members,
        hasAuthorizationEditMember: true,
        hasAuthorizationRemoveMember: true,
      },
    },
    {
      description: 'List data when user has observer role',
      variables: {
        tenant: tenantGlobal,
      },
      role: {
        type: 'observer',
        permission: false,
      },
      props: {
        namespace,
      },
      data: {
        menu: false,
        namespaceMemberFormShow: [],
        namespaceMemberDeleteShow: [],
        editMemberAction: 'editMember',
        removeMemberAction: 'removeMember',
        headers,
      },
      computed: {
        tenant: tenantGlobal,
        members: namespace.members,
        hasAuthorizationEditMember: false,
        hasAuthorizationRemoveMember: false,
      },
    },
  ];

  const storeVuex = (tenant, currentRole) => new Vuex.Store({
    namespaced: true,
    state: {
      tenant,
      currentRole,
    },
    getters: {
      'auth/tenant': (state) => state.tenant,
      'auth/role': (state) => state.currentRole,
    },
    actions: {
      'snackbar/showSnackbarErrorAssociation': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'namespaces/get': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = shallowMount(NamespaceMemberList, {
          store: storeVuex(
            test.variables.tenant,
            test.role.type,
          ),
          localVue,
          stubs: ['fragment'],
          propsData: { namespace: test.props.namespace },
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
        Object.keys(test.computed).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.computed[item]);
        });
      });

      // ///////
      // HTML validation
      ///////

      it('Renders the template with data', () => {
        const dt = wrapper.find('[data-test="dataTable-field"]');
        const dataTableProps = dt.vm.$options.propsData;

        expect(dataTableProps.items).toHaveLength(namespace.members.length);
      });
    });
  });
});
