import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import FirewallRuleList from '@/components/firewall_rule/FirewallRuleList';
import { actions, authorizer } from '../../../../src/authorizer';

describe('FirewallRuleList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const statusGlobal = true;
  const numberFirewallsGlobal = 2;

  const firewallsGlobal = [
    {
      id: '5f1996c8',
      tenant_id: 'xxxxxxxx',
      priority: 4,
      action: 'allow',
      active: true,
      source_ip: '00.00.00',
      username: 'shellhub',
      hostname: 'shellhub',
    },
    {
      id: '5f1996c8',
      tenant_id: 'xxxxxxxx',
      priority: 3,
      action: 'allow',
      active: false,
      source_ip: '00.00.00',
      username: 'shellhub',
      hostname: 'shellhub',
    },
  ];

  const headers = [
    {
      text: 'Active',
      value: 'active',
      align: 'center',
    },
    {
      text: 'Priority',
      value: 'priority',
      align: 'center',
    },
    {
      text: 'Action',
      value: 'action',
      align: 'center',
    },
    {
      text: 'Source IP',
      value: 'source_ip',
      align: 'center',
    },
    {
      text: 'Username',
      value: 'username',
      align: 'center',
    },
    {
      text: 'Hostname',
      value: 'hostname',
      align: 'center',
    },
    {
      text: 'Actions',
      value: 'actions',
      align: 'center',
    },
  ];

  const tests = [
    {
      description: 'List data when user has owner role',
      role: {
        type: 'owner',
        permission: true,
      },
      variables: {
        firewallsGlobal,
        numberFirewallsGlobal,
        statusGlobal,
      },
      data: {
        firewallRuleEditShow: [false, false],
        firewallRuleDeleteShow: [false, false],
        removeAction: 'remove',
        headers,
      },
      computed: {
        getFirewallRules: firewallsGlobal,
        getNumberFirewallRules: numberFirewallsGlobal,
        hasAuthorizationFormDialogRemove: true,
      },
    },
    {
      description: 'List data when user has operator role',
      role: {
        type: 'operator',
        permission: false,
      },
      variables: {
        firewallsGlobal,
        numberFirewallsGlobal,
        statusGlobal,
      },
      data: {
        firewallRuleEditShow: [false, false],
        firewallRuleDeleteShow: [false, false],
        removeAction: 'remove',
        headers,
      },
      computed: {
        getFirewallRules: firewallsGlobal,
        getNumberFirewallRules: numberFirewallsGlobal,
        hasAuthorizationFormDialogRemove: false,
      },
    },
  ];

  const storeVuex = (firewalls, numberFirewalls, status, currentrole) => new Vuex.Store({
    namespaced: true,
    state: {
      firewalls,
      numberFirewalls,
      status,
      currentrole,
    },
    getters: {
      'firewallrules/list': (state) => state.firewalls,
      'firewallrules/getNumberFirewalls': (state) => state.numberFirewalls,
      'boxs/getStatus': (state) => state.status,
      'auth/role': (state) => state.currentrole,
    },
    actions: {
      'firewallrules/fetch': () => {},
      'boxs/setStatus': () => {},
      'snackbar/showSnackbarErrorAssociation': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(FirewallRuleList, {
          store: storeVuex(
            test.variables.firewallsGlobal,
            test.variables.numberFirewallsGlobal,
            test.variables.statusGlobal,
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
        const dt = wrapper.find('[data-test="firewallRuleList-dataTable"]');
        const dataTableProps = dt.vm.$options.propsData;

        expect(dataTableProps.items).toHaveLength(test.variables.numberFirewallsGlobal);
      });
    });
  });
});
