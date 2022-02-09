import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import Vuetify from 'vuetify';
import FirewallRuleFormDialogAdd from '@/components/firewall_rule/FirewallRuleFormDialogAdd';
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

  const ruleFirewall = {
    action: '',
    active: true,
    hostname: '',
    priority: '',
    source_ip: '',
    username: '',
  };

  const tests = [
    {
      description: 'Dialog closed',
      role: {
        type: 'owner',
        permission: true,
      },
      data: {
        dialog: false,
        action: 'create',
        ruleFirewall,
        state: stateRuleFirewall,
      },
      computed: {
        hasAuthorization: true,
      },
      template: {
        'add-btn': true,
        'firewallRuleForm-card': false,
      },
      templateText: {
        'add-btn': 'Add Rule',
      },
    },
    {
      description: 'Dialog closed',
      role: {
        type: 'operator',
        permission: false,
      },
      data: {
        dialog: false,
        action: 'create',
        ruleFirewall,
        state: stateRuleFirewall,
      },
      computed: {
        hasAuthorization: false,
      },
      template: {
        'add-btn': true,
        'firewallRuleForm-card': false,
      },
      templateText: {
        'add-btn': 'Add Rule',
      },
    },
    {
      description: 'Dialog opened',
      role: {
        type: 'owner',
        permission: true,
      },
      data: {
        dialog: true,
        action: 'create',
        ruleFirewall,
        state: stateRuleFirewall,
      },
      computed: {
        hasAuthorization: true,
      },
      template: {
        'add-btn': true,
        'firewallRuleForm-card': true,
        'text-title': true,
        'cancel-btn': true,
        'create-btn': true,
      },
      templateText: {
        'add-btn': 'Add Rule',
        'text-title': 'New Rule',
        'cancel-btn': 'Cancel',
        'create-btn': 'Create',
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
    describe(`${test.description} - ${test.role.type}`, () => {
      beforeEach(async () => {
        wrapper = mount(FirewallRuleFormDialogAdd, {
          store: storeVuex(test.role.type),
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
      it('Renders template with expected text', () => {
        Object.keys(test.templateText).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).text()).toContain(test.templateText[item]);
        });
      });
    });
  });
});
