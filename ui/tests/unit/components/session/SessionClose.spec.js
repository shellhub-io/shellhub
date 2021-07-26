import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import SessionClose from '@/components/session/SessionClose';
import Vuetify from 'vuetify';

describe('SessionClose', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const owner = true;
  const uid = '8c354a00';
  const device = 'a582b47a';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      owner,
    },
    getters: {
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'sessions/close': () => {},
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
      wrapper = mount(SessionClose, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { uid, device },
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
      expect(wrapper.vm.device).toEqual(device);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.session).toEqual({ uid, device_uid: device });
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
        expect(wrapper.find('[data-test="text-tooltip"]').text()).toEqual('Close');
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="sessionClose-card"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is dialog rendering.
  ///////

  describe('Dialog opened', () => {
    beforeEach(() => {
      wrapper = mount(SessionClose, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { uid, device },
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
      expect(wrapper.vm.device).toEqual(device);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.session).toEqual({ uid, device_uid: device });
    });

    //////
    // HTML validation
    //////
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="sessionClose-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="close-btn"]').exists()).toEqual(true);
    });
  });
});
