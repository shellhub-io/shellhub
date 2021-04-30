import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import FirewallRule from '@/components/firewall_rule/FirewallRule';
import Vuetify from 'vuetify';

describe('FirewallRule', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  let wrapper;

  const numberFirewalls = 0;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      numberFirewalls,
    },
    getters: {
      'firewallrules/getNumberFirewalls': (state) => state.numberFirewalls,
    },
    actions: {
      'boxs/setStatus': () => {
      },
      'firewallrules/resetPagePerpage': () => {
      },
      'firewallrules/refresh': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = mount(FirewallRule, {
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
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="firewall-dialog-field"]').exists()).toBe(true);
  });
});
