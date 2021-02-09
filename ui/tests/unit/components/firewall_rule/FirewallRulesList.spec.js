import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import FirewallRulesList from '@/components/firewall_rules/FirewallRulesList';
import Vuetify from 'vuetify';

describe('FirewallRulesList', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  let wrapper;
  let wrapper2;

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

  const store2 = new Vuex.Store({
    namespaced: true,
    state: {
      firewalls,
      numberFirewalls,
      owner: false,
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
    wrapper = mount(FirewallRulesList, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
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
  it('Hides dialogs and delete when the user is not the owner', () => {
    wrapper2 = mount(FirewallRulesList, {
      store: store2,
      localVue,
      stubs: ['fragment'],
      vuetify,
    });
    expect(wrapper2.find('[data-test="firewall-dialog-field"]').exists()).toBe(false);
    expect(wrapper2.find('[data-test="firewall-dialog-field-2"]').exists()).toBe(false);
    expect(wrapper2.find('[data-test="firewall-delete-field"]').exists()).toBe(false);
  });
  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="dataTable-field"]');
    const dataTableProps = dt.vm.$options.propsData;
    expect(dataTableProps.items).toHaveLength(numberFirewalls);
    expect(wrapper.find('[data-test="firewall-dialog-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-dialog-field-2"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-delete-field"]').exists()).toBe(true);
  });
});
