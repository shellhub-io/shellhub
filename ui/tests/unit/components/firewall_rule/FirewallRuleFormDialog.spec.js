import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import Vuetify from 'vuetify';
import FirewallRuleFormDialog from '@/components/firewall_rule/FirewallRuleFormDialog';
import { actions, authorizer } from '../../../../src/authorizer';
import '@/vee-validate';

describe('FirewallRuleFormDialog', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const role = ['owner', 'operator'];

  const hasAuthorization = {
    owner: true,
    operator: false,
  };

  const stateRuleFirewall = [
    {
      id: 'allow',
      name: 'allow',
    },
    {
      id: 'deny',
      name: 'deny',
    },
  ];

  const ruleFirewallLocal = {
    action: '',
    active: true,
    hostname: '',
    priority: '',
    source_ip: '',
    username: '',
  };

  const firewallRule = {
    id: '5f1996c8',
    tenant_id: 'xxxxxxxx',
    priority: 4,
    action: 'allow',
    active: true,
    source_ip: '00.00.00',
    username: 'shellhub',
    hostname: 'shellhub',
  };

  const tests = [
    {
      description: 'Button add firewall rule',
      props: {
        firewallRule: {},
        createRule: true,
        show: false,
      },
      data: {
        ruleFirewallLocal,
        state: stateRuleFirewall,
      },
      template: {
        'add-btn': true,
        'firewallRuleForm-card': false,
        'cancel-btn': false,
        'create-btn': false,
        'edit-btn': false,
      },
    },
    {
      description: 'Button edit firewall rule',
      variables: {
        createRule: false,
      },
      props: {
        firewallRule: {},
        createRule: false,
        show: false,
      },
      data: {
        ruleFirewallLocal: {},
        state: stateRuleFirewall,
      },
      template: {
        'add-btn': false,
        'firewallRuleForm-card': false,
        'cancel-btn': false,
        'create-btn': false,
        'edit-btn': false,
      },
    },
    {
      description: 'Dialog creating firewall rule',
      variables: {
        createRule: true,
      },
      props: {
        firewallRule: {},
        createRule: true,
        show: true,
      },
      data: {
        ruleFirewallLocal,
        state: stateRuleFirewall,
      },
      template: {
        'add-btn': true,
        'firewallRuleForm-card': true,
        'cancel-btn': true,
        'create-btn': true,
        'edit-btn': false,
      },
    },
    {
      description: 'Dialog editing firewall rule',
      variables: {
        createRule: false,
      },
      props: {
        firewallRule,
        createRule: false,
        show: true,
      },
      data: {
        ruleFirewallLocal: firewallRule,
        state: stateRuleFirewall,
      },
      template: {
        'add-btn': false,
        'firewallRuleForm-card': true,
        'cancel-btn': true,
        'create-btn': false,
        'edit-btn': true,
      },
    },
  ];

  const storeVuex = (currentrole) => new Vuex.Store({
    namespaced: true,
    state: {
      currentrole,
    },
    getters: {
      'auth/role': (state) => state.currentrole,
    },
    actions: {
      'firewallrules/post': () => {},
      'firewallrules/put': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(async () => {
          wrapper = mount(FirewallRuleFormDialog, {
            store: storeVuex(currentrole),
            localVue,
            stubs: ['fragment'],
            propsData: {
              firewallRule: test.props.firewallRule,
              createRule: test.props.createRule,
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
          Object.keys(test.template).forEach((item) => {
            if (hasAuthorization[currentrole]) {
              expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
            }
          });
        });
      });
    });
  });
});
