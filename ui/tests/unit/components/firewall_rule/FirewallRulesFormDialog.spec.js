import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import FirewallRulesFormDialog from '@/components/firewall_rules/FirewallRulesFormDialog';

describe('FirewallRulesFormDialog', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const firewallRule = {
    id: '5f1996c84d2190a22d5857bb',
    tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
    priority: 4,
    action: 'allow',
    active: true,
    source_ip: '127.0.0.1',
    username: 'shellhub',
    hostname: 'shellhub',
  };
  const createRule = true;
  const state = [
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
    },
    getters: {
    },
    actions: {
      'firewallrules/post': () => {
      },
      'firewallrules/put': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(FirewallRulesFormDialog, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { firewallRule, createRule },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.firewallRule).toEqual(firewallRule);
    expect(wrapper.vm.createRule).toEqual(createRule);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
    expect(wrapper.vm.state).toEqual(state);
  });
});
