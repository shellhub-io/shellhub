import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import TagDelete from '@/components/setting/tag/TagDelete';

describe('TagDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const isOwner = true;
  const tagName = 'tag';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
    },
    actions: {
      'tags/remove': () => {},
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
      wrapper = mount(TagDelete, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { tagName },
        vuetify,
      });
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      document.body.setAttribute('data-app', true);
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Receive data in props', () => {
      expect(wrapper.vm.tagName).toEqual(tagName);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.isOwner).toEqual(isOwner);
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
      expect(helpIcon.text()).toEqual('delete');
      requestAnimationFrame(() => {
        expect(wrapper.find('[data-test="text-span"]').text()).toEqual('Remove');
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="tagDelete-card"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is dialog rendering.
  ///////

  describe('Dialog', () => {
    beforeEach(() => {
      wrapper = mount(TagDelete, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { tagName },
        vuetify,
      });

      wrapper.setData({ dialog: true });
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      document.body.setAttribute('data-app', true);
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Receive data in props', () => {
      expect(wrapper.vm.tagName).toEqual(tagName);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.isOwner).toEqual(isOwner);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="tagDelete-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="close-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="remove-btn"]').exists()).toEqual(true);
    });
  });
});
