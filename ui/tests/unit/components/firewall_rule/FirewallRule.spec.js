import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import FirewallRule from '@/components/firewall_rule/FirewallRule';
import Vuetify from 'vuetify';

describe('FirewallRule', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

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
    // Data and Props checking
    //////

    it('Process data in the computed', () => {
      expect(wrapper.vm.hasFirewallRule).toEqual(false);
      expect(wrapper.vm.showBoxMessage).toEqual(true);
    });
    it('Compare data with the default', () => {
      expect(wrapper.vm.show).toEqual(true);
      expect(wrapper.vm.showHelp).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="firewallRuleCreate-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="boxMessageFirewall-component"]').exists()).toBe(true);
    });
  });

  ///////
  // In this case, it is tested when there is already a registered
  // firewall.
  ///////

  describe('Without firewall rules', () => {
    beforeEach(() => {
      wrapper = mount(FirewallRule, {
        store: storeWithFirewalls,
        localVue,
        stubs: ['fragment'],
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
    // Data and Props checking
    //////

    it('Process data in the computed', () => {
      expect(wrapper.vm.hasFirewallRule).toEqual(true);
      expect(wrapper.vm.showBoxMessage).toEqual(false);
    });
    it('Compare data with the default and defined value', () => {
      expect(wrapper.vm.show).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="firewallRuleCreate-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="boxMessageFirewall-component"]').exists()).toBe(false);
    });
  });
});
