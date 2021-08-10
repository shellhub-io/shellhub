import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import KeyDelete from '@/components/public_key/KeyDelete';

describe('KeyDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const isOwner = true;
  const fingerprint = 'b7:25:f8';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
    },
    actions: {
      'publickeys/remove': () => {},
    },
  });

  ///////
  // in this case, when the user owns the namespace and the focus of
  // the test is icon rendering. The action is public.
  ///////

  describe('Icon', () => {
    beforeEach(() => {
      const action = 'public';

      wrapper = mount(KeyDelete, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { fingerprint, action },
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
      expect(wrapper.vm.fingerprint).toEqual(fingerprint);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.isOwner).toEqual(isOwner);
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
      expect(wrapper.find('[data-test="keyDelete-card"]').exists()).toEqual(false);
    });
  });

  ///////
  // in this case, when the user owns the namespace and the focus of
  // the test is icon rendering. The action is private.
  ///////

  describe('Icon', () => {
    beforeEach(() => {
      const action = 'private';

      wrapper = mount(KeyDelete, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { fingerprint, action },
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
      expect(wrapper.vm.fingerprint).toEqual(fingerprint);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.isOwner).toEqual(isOwner);
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
      expect(wrapper.find('[data-test="keyDelete-card"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, when the user owns the keys and the focus of
  // the test is dialog rendering. The action is public.
  ///////

  describe('Dialog opened', () => {
    beforeEach(() => {
      const action = 'public';

      wrapper = mount(KeyDelete, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { fingerprint, action },
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
      expect(wrapper.vm.fingerprint).toEqual(fingerprint);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.isOwner).toEqual(isOwner);
    });
    it('Process data in methods', () => {
      wrapper.vm.close();

      expect(wrapper.vm.dialog).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="keyDelete-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="close-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="Remove-btn"]').exists()).toEqual(true);
    });
  });

  ///////
  // In this case, when the user owns the keys and the focus of
  // the test is dialog rendering. The action is private.
  ///////

  describe('Dialog opened', () => {
    beforeEach(() => {
      const action = 'private';

      wrapper = mount(KeyDelete, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { fingerprint, action },
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
      expect(wrapper.vm.fingerprint).toEqual(fingerprint);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.isOwner).toEqual(isOwner);
    });
    it('Process data in methods', () => {
      wrapper.vm.close();

      expect(wrapper.vm.dialog).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="keyDelete-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="close-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="Remove-btn"]').exists()).toEqual(true);
    });
  });
});
