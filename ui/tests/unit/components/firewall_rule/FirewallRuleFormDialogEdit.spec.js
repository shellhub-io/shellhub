import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import Vuetify from 'vuetify';
import FirewallRuleFormDialog from '@/components/firewall_rule/FirewallRuleFormDialogEdit';
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
      description: 'Dialog closed',
      props: {
        firewallRule,
        show: false,
      },
      data: {
        dialog: false,
        state: stateRuleFirewall,
        ruleFirewallLocal: firewallRule,
      },
      template: {
        'edit-icon': true,
        'edit-title': true,
        'firewallRuleForm-card': false,
      },
      templateText: {
        'edit-title': 'Edit',
      },
    },
    {
      description: 'Dialog opened',
      props: {
        firewallRule,
        show: true,
      },
      data: {
        dialog: false,
        state: stateRuleFirewall,
        ruleFirewallLocal: firewallRule,
      },
      template: {
        'edit-icon': true,
        'edit-title': true,
        'firewallRuleForm-card': true,
        'text-title': true,
        'priority-field': true,
        'action-field': true,
        'source_ip-field': true,
        'username-field': true,
        'hostname-field': true,
        'cancel-btn': true,
        'edit-btn': true,
      },
      templateText: {
        'edit-title': 'Edit',
        'text-title': 'Edit Firewall Rule',
        'cancel-btn': 'Cancel',
        'edit-btn': 'Edit',
      },
    },
  ];

  const storeVuex = () => new Vuex.Store({
    namespaced: true,
    state: { },
    getters: { },
    actions: {
      'firewallrules/post': () => {},
      'firewallrules/put': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        wrapper = mount(FirewallRuleFormDialog, {
          store: storeVuex(),
          localVue,
          stubs: ['fragment'],
          propsData: {
            firewallRule: test.props.firewallRule,
            show: test.props.show,
          },
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
