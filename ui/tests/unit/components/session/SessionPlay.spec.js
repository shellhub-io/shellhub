import Vuex from 'vuex';
import { mount, createLocalVue, config } from '@vue/test-utils';
import SessionPlay from '@/components/session/SessionPlay';
import Vuetify from 'vuetify';

config.mocks = {
  $env: {
    isEnterprise: true,
  },
};

describe('SessionPlay', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const isOwner = true;
  const recorded = true;
  const speedList = [0.5, 1, 1.5, 2, 4];

  const session = [
    {
      uid: '1a0536ab',
      message: '\u001b]0;shellhub@shellhub: ~\u0007shellhub@shellhub:~$ ',
      tenant_id: 'xxxxxxxx',
      time: '2020-09-24T18:32:04.559Z',
      width: 110,
      height: 23,
    },
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
      session,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
      'sessions/get': (state) => state.session,
    },
    actions: {
      'sessions/getLogSession': () => {
      },
    },
  });

  ///////
  // in this case, when the user owns the namespace and the focus of
  // the test is icon rendering.
  ///////

  describe('Icon', () => {
    beforeEach(() => {
      wrapper = mount(SessionPlay, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { uid: session.uid, recorded },
        mocks: ['$env'],
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
      expect(wrapper.vm.uid).toEqual(session.uid);
      expect(wrapper.vm.recorded).toEqual(recorded);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.currentTime).toEqual(0);
      expect(wrapper.vm.totalLength).toEqual(0);
      expect(wrapper.vm.endTimerDisplay).toEqual(0);
      expect(wrapper.vm.getTimerNow).toEqual(0);
      expect(wrapper.vm.paused).toEqual(false);
      expect(wrapper.vm.previousPause).toEqual(false);
      expect(wrapper.vm.sliderChange).toEqual(false);
      expect(wrapper.vm.speedList).toEqual(speedList);
      expect(wrapper.vm.logs).toEqual([]);
      expect(wrapper.vm.frames).toEqual([]);
      expect(wrapper.vm.defaultSpeed).toEqual(1);
      expect(wrapper.vm.transition).toEqual(false);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.length).toEqual(0);
      expect(wrapper.vm.nowTimerDisplay).toEqual(0);
      expect(wrapper.vm.isOwner).toEqual(true);
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
        expect(wrapper.find('[data-test="text-tooltip"]').text()).toEqual('Play');
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="render-fragment"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="sessionPlay-card"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, you are testing the rendering of the dialog.
  ///////

  describe('Dialog opened', () => {
    beforeEach(() => {
      wrapper = mount(SessionPlay, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { uid: session.uid, recorded },
        mocks: ['$env'],
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
      expect(wrapper.vm.uid).toEqual(session.uid);
      expect(wrapper.vm.recorded).toEqual(recorded);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.currentTime).toEqual(0);
      expect(wrapper.vm.totalLength).toEqual(0);
      expect(wrapper.vm.endTimerDisplay).toEqual(0);
      expect(wrapper.vm.getTimerNow).toEqual(0);
      expect(wrapper.vm.paused).toEqual(false);
      expect(wrapper.vm.previousPause).toEqual(false);
      expect(wrapper.vm.sliderChange).toEqual(false);
      expect(wrapper.vm.speedList).toEqual(speedList);
      expect(wrapper.vm.logs).toEqual([]);
      expect(wrapper.vm.frames).toEqual([]);
      expect(wrapper.vm.defaultSpeed).toEqual(1);
      expect(wrapper.vm.transition).toEqual(false);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.length).toEqual(0);
      expect(wrapper.vm.nowTimerDisplay).toEqual(0);
      expect(wrapper.vm.isOwner).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="render-fragment"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="sessionPlay-card"]').exists()).toEqual(true);
    });
  });

  ///////
  // In this case, you are testing the rendering when the owner is
  // not the owner.
  ///////

  describe('Dialog opened', () => {
    beforeEach(() => {
      wrapper = mount(SessionPlay, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { uid: session.uid, recorded },
        mocks: {
          $env: {
            isEnterprise: false,
          },
        },
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
      expect(wrapper.vm.uid).toEqual(session.uid);
      expect(wrapper.vm.recorded).toEqual(recorded);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.currentTime).toEqual(0);
      expect(wrapper.vm.totalLength).toEqual(0);
      expect(wrapper.vm.endTimerDisplay).toEqual(0);
      expect(wrapper.vm.getTimerNow).toEqual(0);
      expect(wrapper.vm.paused).toEqual(false);
      expect(wrapper.vm.previousPause).toEqual(false);
      expect(wrapper.vm.sliderChange).toEqual(false);
      expect(wrapper.vm.speedList).toEqual(speedList);
      expect(wrapper.vm.logs).toEqual([]);
      expect(wrapper.vm.frames).toEqual([]);
      expect(wrapper.vm.defaultSpeed).toEqual(1);
      expect(wrapper.vm.transition).toEqual(false);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.length).toEqual(0);
      expect(wrapper.vm.nowTimerDisplay).toEqual(0);
      expect(wrapper.vm.isOwner).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="render-fragment"]').exists()).toEqual(false);
    });
  });
});
