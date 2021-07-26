import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import FirewallRuleFormDialog from '@/components/firewall_rule/FirewallRuleFormDialog';
import Vuetify from 'vuetify';

describe('FirewallRuleFormDialog', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const isOwner = true;
  const createRule = true;

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

  const stateFirewallRule = [
    {
      id: 'allow',
      name: 'allow',
    },
    {
      id: 'deny',
      name: 'deny',
    },
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
    },
    actions: {
      'firewallrules/post': () => {},
      'firewallrules/put': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  ///////
  // in this case, when the user owns the namespace and the focus of
  // the test is button rendering. Add firewall rule
  ///////

  describe('Button', () => {
    beforeEach(() => {
      wrapper = mount(FirewallRuleFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { firewallRule, createRule },
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

    it('Receive data in props', () => {
      expect(wrapper.vm.firewallRule).toEqual(firewallRule);
      expect(wrapper.vm.createRule).toEqual(createRule);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.state).toEqual(stateFirewallRule);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="add-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="firewallRuleForm-card"]').exists()).toEqual(false);
    });
  });

  //////
  // In this case, when the user owns the namespace and the focus of
  // the test is icon rendering. Editing firewall rule
  //////

  describe('Button', () => {
    beforeEach(() => {
      wrapper = mount(FirewallRuleFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { firewallRule, createRule: !createRule },
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

    it('Receive data in props', () => {
      expect(wrapper.vm.firewallRule).toEqual(firewallRule);
      expect(wrapper.vm.createRule).toEqual(!createRule);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.state).toEqual(stateFirewallRule);
    });

    //////
    // HTML validation
    //////

    it('Show message tooltip to user owner', async (done) => {
      const icons = wrapper.findAll('.v-icon');
      const helpIcon = icons.at(0);
      helpIcon.trigger('mouseenter');
      await wrapper.vm.$nextTick();

      expect(icons.length).toBe(1);
      requestAnimationFrame(() => {
        expect(wrapper.find('[data-test="text-tooltip"]').text()).toEqual('Edit');
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="add-btn"]').exists()).toEqual(false);
      expect(wrapper.find('[firewallRuleForm-card]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is dialog rendering. Creating firewall rule
  ///////

  describe('Button', () => {
    beforeEach(() => {
      wrapper = mount(FirewallRuleFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { firewallRule, createRule },
        vuetify,
      });

      wrapper.setData({ dialog: true });
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

    it('Receive data in props', () => {
      expect(wrapper.vm.firewallRule).toEqual(firewallRule);
      expect(wrapper.vm.createRule).toEqual(createRule);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.state).toEqual(stateFirewallRule);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="firewallRuleForm-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="create-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="edit-btn"]').exists()).toEqual(false);
    });
  });

  //////
  // In this case, when the user owns the namespace and the focus of
  // the test is dialog rendering. Editing firewall rule
  //////

  describe('Button', () => {
    beforeEach(() => {
      wrapper = mount(FirewallRuleFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { firewallRule, createRule: !createRule },
        vuetify,
      });

      wrapper.setData({ dialog: true });
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

    it('Receive data in props', () => {
      expect(wrapper.vm.firewallRule).toEqual(firewallRule);
      expect(wrapper.vm.createRule).toEqual(!createRule);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.state).toEqual(stateFirewallRule);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="firewallRuleForm-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="create-btn"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="edit-btn"]').exists()).toEqual(true);
    });
  });
});
