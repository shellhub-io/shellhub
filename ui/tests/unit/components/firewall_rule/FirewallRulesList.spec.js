import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import FirewallRulesList from '@/components/firewall_rules/FirewallRulesList';

describe('FirewallRulesList', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const numberFirewalls = 2;
  const firewalls = [
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
      id: '5f1996c84d2190a22d5857cc',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      priority: 3,
      action: 'allow',
      active: false,
      source_ip: '127.0.0.1',
      username: 'shellhub',
      hostname: 'shellhub',
    },
  ];
  const owner = true;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      firewalls,
      numberFirewalls,
      owner,
    },
    getters: {
      'firewallrules/list': (state) => state.firewalls,
      'firewallrules/getNumberFirewalls': (state) => state.numberFirewalls,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'firewallrules/fetch': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(FirewallRulesList, {
      store,
      localVue,
      stubs: ['fragment'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.getFirewallRules).toEqual(firewalls);
    expect(wrapper.vm.getNumberFirewallRules).toEqual(numberFirewalls);
  });
  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="dataTable-field"]');
    const dataTableProps = dt.vm.$options.propsData;
    expect(dataTableProps.items).toHaveLength(numberFirewalls);
  });
});
