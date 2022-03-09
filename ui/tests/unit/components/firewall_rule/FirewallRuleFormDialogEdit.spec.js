import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import FirewallRuleFormDialog from '@/components/firewall_rule/FirewallRuleFormDialogEdit';
import '@/vee-validate';

describe('FirewallRuleFormDialog', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  const usernameFieldChoices = [
    {
      filterName: 'all',
      filterText: 'Define rule to all users',
    },
    {
      filterName: 'username',
      filterText: 'Restrict access using a regexp for username',
    },
  ];

  const filterFieldChoices = [
    {
      filterName: 'all',
      filterText: 'Define rule to all devices',
    },
    {
      filterName: 'hostname',
      filterText: 'Restrict rule with a regexp for hostname',
    },
    {
      filterName: 'tags',
      filterText: 'Restrict rule by device tags',
    },
  ];

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

  const firewallRuleProps = {
    id: '5f1996c8',
    tenant_id: 'xxxxxxxx',
    priority: 4,
    source_ip: '00.00.00',
    username: '.*',
    filter: {
      hostname: 'hostname',
    },
    action: 'allow',
    active: true,
  };

  const firewallRuleData = {
    priority: 0,
    source_ip: '',
    username: '',
    filter: {
    },
    policy: '',
    status: '',
  };

  const tests = [
    {
      description: 'Dialog closed',
      props: {
        firewallRule: firewallRuleProps,
        show: false,
      },
      data: {
        hostnameField: '',
        state: stateRuleFirewall,
        choiceFilter: 'all',
        choiceUsername: 'all',
        usernameFieldChoices,
        filterFieldChoices,
        ruleFirewallLocal: firewallRuleData,
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
        firewallRule: firewallRuleProps,
        show: true,
      },
      data: {
        state: stateRuleFirewall,
        hostnameField: '',
        choiceFilter: 'all',
        choiceUsername: 'all',
        usernameFieldChoices,
        filterFieldChoices,
        ruleFirewallLocal: firewallRuleData,
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
        'filter-field': true,
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
          mocks: {
            $errors: {
              snackbar: {
                firewallRuleEditing: '',
              },
            },
          },
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
      Object.keys(test.data).forEach((item) => {
        it(`Compare data ${item} with its default value`, () => {
          expect(wrapper.vm[item]).toEqual(test.data[item]);
        });
      });

      //////
      // HTML validation
      //////

      Object.keys(test.template).forEach((item) => {
        it(`Renders the template ${item} with data`, () => {
          expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
        });
      });

      Object.keys(test.templateText).forEach((item) => {
        it(`Renders template ${item} with expected text`, () => {
          expect(wrapper.find(`[data-test="${item}"]`).text()).toContain(test.templateText[item]);
        });
      });
    });
  });

  describe('Updating data checks', () => {
    it('Should setLocalVariable populate ruleFirewallLocal for all hostname', async () => {
      await wrapper.setProps({
        firewallRule: {
          priority: 1,
          source_ip: '.*',
          filter: {
            hostname: '.*',
          },
          username: '.*',
          status: 'active',
          policy: 'allow',
        },
      });

      wrapper.vm.setLocalVariable();

      expect(wrapper.vm.choiceFilter).toBe('all');
      expect(wrapper.vm.choiceUsername).toBe('all');
      expect(wrapper.vm.choiceIP).toBe('all');
      expect(wrapper.vm.usernameField).toBe('');
      expect(wrapper.vm.ipField).toBe('');
      expect(wrapper.vm.hostnameField).toStrictEqual('');
    });

    it('Should setLocalVariable populate ruleFirewallLocal for tag filter', async () => {
      await wrapper.setProps({
        firewallRule: {
          priority: 1,
          source_ip: '2',
          filter: {
            tags: ['tag1', 'tag2'],
          },
          username: 'user',
          status: 'active',
          policy: 'allow',
        },
      });

      wrapper.vm.setLocalVariable();

      expect(wrapper.vm.choiceFilter).toBe('tags');
      expect(wrapper.vm.choiceUsername).toBe('username');
      expect(wrapper.vm.usernameField).toBe('user');
      expect(wrapper.vm.tagChoices).toStrictEqual(['tag1', 'tag2']);
    });

    it('Should setLocalVariable populate ruleFirewallLocal for hostname filter', async () => {
      await wrapper.setProps({
        firewallRule: {
          priority: 1,
          source_ip: '2',
          filter: {
            hostname: 'name',
          },
          username: '.*',
          status: 'active',
          policy: 'allow',
        },
      });

      await flushPromises();
      wrapper.vm.setLocalVariable();

      expect(wrapper.vm.choiceFilter).toBe('hostname');
      expect(wrapper.vm.choiceUsername).toBe('all');
      expect(wrapper.vm.hostnameField).toStrictEqual('name');
    });

    it('Should showDialog enabled call setLocalVariable', () => {
      const setMock = jest.spyOn(wrapper.vm, 'setLocalVariable');
      wrapper.vm.$options.watch.showDialog.call(wrapper.vm, true);

      expect(setMock).toHaveBeenCalled();
    });

    it('Should select restriction update ruleFirewallLocal filter tags', async () => {
      const rfl = wrapper.vm.ruleFirewallLocal;

      const tags = ['tag1', 'tag2'];
      await wrapper.setData({ choiceFilter: 'tags', tagChoices: tags });
      await flushPromises();

      wrapper.vm.selectRestriction();
      expect(wrapper.vm.ruleFirewallLocal).toStrictEqual({ ...rfl, filter: { tags } });
    });

    it('Should select restriction update ruleFirewallLocal filter ip_address', async () => {
      const rfl = wrapper.vm.ruleFirewallLocal;

      const ipAddr = '123';
      await wrapper.setData({ choiceIP: 'ipDetails', ipField: ipAddr });
      await flushPromises();

      wrapper.vm.selectRestriction();
      expect(wrapper.vm.ruleFirewallLocal).toStrictEqual({ ...rfl, source_ip: ipAddr });
    });

    it('Should select restriction update ruleFirewallLocal for field username', async () => {
      const rfl = wrapper.vm.ruleFirewallLocal;

      const uf = 'user';
      await wrapper.setData({ choiceUsername: 'username', usernameField: uf });
      await flushPromises();

      wrapper.vm.selectRestriction();
      expect(wrapper.vm.ruleFirewallLocal).toStrictEqual({ ...rfl, username: uf });
    });

    it('Should select restriction update ruleFirewallLocal filter hostname', async () => {
      const rfl = wrapper.vm.ruleFirewallLocal;

      const hostname = 'hostname';
      await wrapper.setData({ choiceFilter: 'hostname', hostnameField: hostname });
      await flushPromises();

      wrapper.vm.selectRestriction();
      expect(wrapper.vm.ruleFirewallLocal).toStrictEqual({ ...rfl, filter: { hostname } });
    });

    it('Should edit method call selectRestriction', () => {
      const constrMock = jest.spyOn(wrapper.vm, 'selectRestriction');
      wrapper.vm.edit();

      expect(constrMock).toHaveBeenCalled();
    });

    it('Should display tags field textfield when choice is tags', async () => {
      wrapper = mount(FirewallRuleFormDialog, {
        store: storeVuex(),
        localVue,
        stubs: ['fragment'],
        propsData: {
          firewallRule: firewallRuleProps,
          show: true,
        },
        vuetify,
      });

      await wrapper.setData({ choiceFilter: 'tags' });
      expect(wrapper.find('[data-test="tags-selector"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="hostname-field"]').exists()).toBe(false);
    });
  });
});
