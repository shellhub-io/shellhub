import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import FirewallRuleList from '@/components/firewall_rule/FirewallRuleList';

describe('FirewallRuleList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const status = true;
  const numberFirewalls = 2;

  const firewalls = [
    {
      id: '5f1996c8',
      tenant_id: 'xxxxxxxx',
      priority: 4,
      action: 'allow',
      active: true,
      source_ip: '00.00.00',
      username: 'shellhub',
      hostname: 'shellhub',
    },
    {
      id: '5f1996c8',
      tenant_id: 'xxxxxxxx',
      priority: 3,
      action: 'allow',
      active: false,
      source_ip: '00.00.00',
      username: 'shellhub',
      hostname: 'shellhub',
    },
  ];

  const headers = [
    {
      text: 'Active',
      value: 'active',
      align: 'center',
    },
    {
      text: 'Priority',
      value: 'priority',
      align: 'center',
    },
    {
      text: 'Action',
      value: 'action',
      align: 'center',
    },
    {
      text: 'Source IP',
      value: 'source_ip',
      align: 'center',
    },
    {
      text: 'Username',
      value: 'username',
      align: 'center',
    },
    {
      text: 'Hostname',
      value: 'hostname',
      align: 'center',
    },
    {
      text: 'Actions',
      value: 'actions',
      align: 'center',
    },
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      firewalls,
      numberFirewalls,
      status,
    },
    getters: {
      'firewallrules/list': (state) => state.firewalls,
      'firewallrules/getNumberFirewalls': (state) => state.numberFirewalls,
      'boxs/getStatus': (state) => state.status,
    },
    actions: {
      'firewallrules/fetch': () => {},
      'boxs/setStatus': () => {},
      'snackbar/showSnackbarErrorAssociation': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  beforeEach(() => {
    wrapper = mount(FirewallRuleList, {
      store,
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

  it('Compare data with default value', () => {
    expect(wrapper.vm.headers).toEqual(headers);
    expect(wrapper.vm.showHelp).toEqual(false);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.getFirewallRules).toEqual(firewalls);
    expect(wrapper.vm.getNumberFirewallRules).toEqual(numberFirewalls);
  });

  //////
  // HTML validation
  //////

  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="firewallRuleList-dataTable"]');
    const dataTableProps = dt.vm.$options.propsData;

    expect(dataTableProps.items).toHaveLength(numberFirewalls);
    expect(wrapper.find('[data-test="firewallRuleEdit-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewallRuleDelete-component"]').exists()).toBe(true);
  });
});
