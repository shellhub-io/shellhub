import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import FirewallRulesFormDialog from '@/components/firewall_rules/FirewallRulesFormDialog';

describe('FirewallRulesFormDialog', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  const firewallRule = [
    {
      id: '5f1996c84d2190a22d5857bb',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      priority: 4,
      action: 'allow',
      active: true,
      source_ip: '127.0.0.1',
      username: 'shellhub',
      hostname: 'shellhub',
    },
    {
      id: '5f1996fe4d2190a22d5857c1',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      priority: 18,
      action: 'deny',
      active: true,
      source_ip: '127.0.0.2',
      username: 'shellhub',
      hostname: '.*',
    }];
  const createRule = true;

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
      props: { firewallRule, createRule },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
