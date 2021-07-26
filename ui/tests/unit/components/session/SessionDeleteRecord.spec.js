import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import SessionDeleteRecord from '@/components/session/SessionDeleteRecord';
import Vuetify from 'vuetify';

describe('SessionDeleteRecord', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const owner = true;
  const uid = '8c354a00';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      owner,
    },
    getters: {
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'sessions/deleteSessionLogs': () => {},
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
      wrapper = mount(SessionDeleteRecord, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { uid },
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
      expect(wrapper.vm.uid).toEqual(uid);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
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
        expect(wrapper.find('[data-test="text-tooltip"]').text()).toEqual('Delete session record');
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="sessionDeleteRecord-card"]').exists()).toEqual(false);
    });
  });

  ///////
  // in this case, when the user owns the namespace and the focus of
  // the test is dialog rendering.
  ///////

  describe('Dialog opened', () => {
    beforeEach(() => {
      wrapper = mount(SessionDeleteRecord, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { uid },
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
      expect(wrapper.vm.uid).toEqual(uid);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="sessionDeleteRecord-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="delete-btn"]').exists()).toEqual(true);
    });
  });
});
