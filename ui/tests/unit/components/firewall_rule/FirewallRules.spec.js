import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import FirewallRules from '@/components/firewall_rules/FirewallRules';
import Vuetify from 'vuetify';

describe('FirewallRules', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  let wrapper;

  const isOwner = true;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
    },
    actions: {
      'firewallrules/refresh': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
    },
  });

  const store2 = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner: !isOwner,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
    },
    actions: {
      'firewallrules/refresh': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = mount(FirewallRules, {
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
    expect(wrapper.vm.isOwner).toEqual(isOwner);
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="firewall-dialog-field"]').exists()).toBe(true);
  });
  it('Hides dialogs when the user is not the owner', () => {
    const wrapper2 = mount(FirewallRules, {
      store: store2,
      localVue,
      stubs: ['fragment'],
      vuetify,
    });
    expect(wrapper2.find('[data-test="firewall-dialog-field"]').exists()).toBe(false);
  });
});
