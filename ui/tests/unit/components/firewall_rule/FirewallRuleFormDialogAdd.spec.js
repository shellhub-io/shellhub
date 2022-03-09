import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import Vuetify from 'vuetify';
import flushPromises from 'flush-promises';
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
  const stateRuleFirewall = [
    {
      id: 'allow',
      name: 'Allow',
    },
    {
      id: 'deny',
      name: 'Deny',
    },
  ];

  const ruleFirewall = {
    policy: 'allow',
    status: 'active',
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
        usernameFieldChoices,
        filterFieldChoices,
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
        usernameFieldChoices,
        filterFieldChoices,
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
        usernameFieldChoices,
        filterFieldChoices,
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
        'text-title': 'New Firewall Rule',
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
            $errors: {
              snackbar: {
                firewallRuleCreating: '',
              },
            },
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

  describe('Update data checks', () => {
    it('Should construct filter object for hostname', async () => {
      const rf = wrapper.vm.ruleFirewall;

      await wrapper.setData({ choiceFilter: 'hostname', filterField: 'another' });
      await flushPromises();
      wrapper.vm.constructFilterObject();

      expect(wrapper.vm.ruleFirewall).toStrictEqual({ ...rf, filter: { hostname: 'another' } });
    });

    it('Should construct filter object for tags', async () => {
      const rf = wrapper.vm.ruleFirewall;

      const tags = ['tag1', 'tag2'];
      await wrapper.setData({ choiceFilter: 'tags', tagChoices: tags });
      await flushPromises();

      wrapper.vm.constructFilterObject();

      expect(wrapper.vm.ruleFirewall).toStrictEqual({ ...rf, filter: { tags } });
    });

    it('Should create call constructFilterObject', async () => {
      const constrMock = jest.spyOn(wrapper.vm, 'constructFilterObject');
      wrapper.vm.create();

      expect(constrMock).toHaveBeenCalled();
    });
  });
});
