import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import FirewallRuleDelete from '@/components/firewall_rule/FirewallRuleDelete';

describe('FirewallRuleDelete', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const isOwner = true;
  const id = '5f1996c84d2190a22d5857bb';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
    },
    actions: {
      'firewallrules/remove': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(FirewallRuleDelete, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { id },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.id).toEqual(id);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
  });
});
