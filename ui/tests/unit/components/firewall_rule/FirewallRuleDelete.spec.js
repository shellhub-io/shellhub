import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import FirewallRuleDelete from '@/components/firewall_rule/FirewallRuleDelete';
import Vuetify from 'vuetify';

describe('FirewallRuleDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const isOwner = true;
  const id = '5f1996c8';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
    },
    actions: {
      'firewallrules/remove': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  ///////
  // in this case, when the user owns the namespace and the focus of
  // the test is icon rendering.
  ///////

  describe('Icon', () => {
    beforeEach(() => {
      wrapper = mount(FirewallRuleDelete, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { id },
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

    it('Receive data in props', () => {
      expect(wrapper.vm.id).toEqual(id);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
    });
    it('Process data in methods', () => {
      wrapper.vm.close();
      expect(wrapper.vm.dialog).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Show message tooltip to user owner', async (done) => {
      const icons = wrapper.findAll('.v-icon');
      const helpIcon = icons.at(0);
      helpIcon.trigger('mouseenter');
      await wrapper.vm.$nextTick();

      expect(icons.length).toBe(1);
      requestAnimationFrame(() => {
        expect(wrapper.find('[data-test="text-tooltip"]').text()).toEqual('Remove');
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="firewallRuleDelete-card"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is dialog rendering.
  ///////

  describe('Dialog opened', () => {
    beforeEach(() => {
      wrapper = mount(FirewallRuleDelete, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { id },
        vuetify,
      });

      wrapper.setData({ dialog: true });
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

    it('Receive data in props', () => {
      expect(wrapper.vm.id).toEqual(id);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
    });
    it('Process data in methods', () => {
      wrapper.vm.close();
      expect(wrapper.vm.dialog).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="firewallRuleDelete-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="remove-btn"]').exists()).toEqual(true);
    });
  });
});
