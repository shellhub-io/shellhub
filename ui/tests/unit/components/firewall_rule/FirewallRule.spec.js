import Vuex from 'vuex';
import { mount, createLocalVue, config } from '@vue/test-utils';
import Vuetify from 'vuetify';
import Router from 'vue-router';
import FirewallRule from '@/components/firewall_rule/FirewallRule';
import { actions, authorizer } from '../../../../src/authorizer';

config.mocks = {
  $env: {
    isCloud: true,
  },
};

describe('FirewallRule', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  const router = new Router();
  localVue.use(Vuex);
  localVue.use(Router);

  let wrapper;

  const numberFirewallsEqualZero = 0;
  const numberFirewallsGreaterThanZero = 1;

  const storeWithoutFirewalls = new Vuex.Store({
    namespaced: true,
    state: {
      numberFirewalls: numberFirewallsEqualZero,
    },
    getters: {
      'firewallrules/getNumberFirewalls': (state) => state.numberFirewalls,
    },
    actions: {
      'boxs/setStatus': () => {},
      'firewallrules/resetPagePerpage': () => {},
      'firewallrules/refresh': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  const storeWithFirewalls = new Vuex.Store({
    namespaced: true,
    state: {
      numberFirewalls: numberFirewallsGreaterThanZero,
    },
    getters: {
      'firewallrules/getNumberFirewalls': (state) => state.numberFirewalls,
    },
    actions: {
      'boxs/setStatus': () => {},
      'firewallrules/resetPagePerpage': () => {},
      'firewallrules/refresh': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  ///////
  // In this case, the rendering of the component that shows the
  // message when it does not have access to the device is tested.
  ///////

  describe('Without firewall rules', () => {
    beforeEach(() => {
      wrapper = mount(FirewallRule, {
        store: storeWithoutFirewalls,
        localVue,
        stubs: ['fragment'],
        router,
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.hasFirewallRule).toEqual(false);
      expect(wrapper.vm.showBoxMessage).toEqual(true);
    });
    it('Compare data with the default', () => {
      expect(wrapper.vm.show).toEqual(true);
      expect(wrapper.vm.showHelp).toEqual(false);
      expect(wrapper.vm.firewallRuleCreateShow).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="FirewallRuleFormDialogAdd-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="boxMessageFirewall-component"]').exists()).toBe(true);
    });
  });

  ///////
  // In this case, it is tested when there is already a registered
  // firewall.
  ///////

  describe('With firewall rules', () => {
    beforeEach(() => {
      wrapper = mount(FirewallRule, {
        store: storeWithFirewalls,
        localVue,
        stubs: ['fragment'],
        router,
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.hasFirewallRule).toEqual(true);
      expect(wrapper.vm.showBoxMessage).toEqual(false);
    });
    it('Compare data with the default and defined value', () => {
      expect(wrapper.vm.show).toEqual(true);
      expect(wrapper.vm.showHelp).toEqual(false);
      expect(wrapper.vm.firewallRuleCreateShow).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="FirewallRuleFormDialogAdd-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="boxMessageFirewall-component"]').exists()).toBe(false);
    });
  });
});
