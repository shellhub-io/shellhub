import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import SessionPlay from '@/components/session/SessionPlay';

describe('SessionPlay', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const sessionGlobal = [
    {
      uid: '1a0536ab',
      message: '\u001b]0;shellhub@shellhub: ~\u0007shellhub@shellhub:~$ ',
      tenant_id: 'xxxxxxxx',
      time: '2020-09-24T18:32:04.559Z',
      width: 110,
      height: 23,
    },
  ];

  const tests = [
    {
      description: 'Dialog closed',
      variables: {
        session: [],
        recorded: false,
        paused: false,
        dialog: false,
      },
      props: {
        uid: '8c354a00',
        recorded: false,
        show: false,
      },
      data: {
        dialog: false,
        currentTime: 0,
        totalLength: 0,
        endTimerDisplay: 0,
        getTimerNow: 0,
        paused: false,
        previousPause: false,
        sliderChange: false,
        speedList: [0.5, 1, 1.5, 2, 4],
        logs: [],
        frames: [],
        defaultSpeed: 1,
        transition: false,
      },
      computed: {
        length: 0,
        nowTimerDisplay: 0,
      },
      template: {
        'play-icon': true,
        'play-title': true,
        'sessionPlay-card': false,
      },
      templateText: {
        'play-title': 'Play',
      },
    },
    {
      description: 'Dialog opened with play paused',
      variables: {
        session: sessionGlobal,
        recorded: true,
        paused: false,
        dialog: true,
      },
      props: {
        uid: '8c354a00',
        recorded: true,
        show: true,
      },
      data: {
        dialog: true,
        currentTime: 0,
        totalLength: 0,
        endTimerDisplay: 0,
        getTimerNow: '00:00',
        paused: false,
        previousPause: false,
        sliderChange: false,
        speedList: [0.5, 1, 1.5, 2, 4],
        logs: sessionGlobal,
        frames: [],
        defaultSpeed: 1,
        transition: false,
      },
      computed: {
        length: 1,
        nowTimerDisplay: '00:00',
      },
      template: {
        'play-icon': true,
        'play-title': true,
        'sessionPlay-card': true,
        'close-btn': true,
        'text-title': true,
        'pause-icon': true,
        'time-slider': true,
        'speed-select': true,
      },
      templateText: {
        'play-title': 'Play',
        'text-title': 'Watch Session',
      },
    },
    {
      description: 'Dialog opened with play not paused',
      variables: {
        session: sessionGlobal,
        recorded: true,
        paused: true,
        show: true,
        dialog: true,
      },
      props: {
        uid: '8c354a00',
        recorded: true,
        show: true,
      },
      data: {
        dialog: true,
        currentTime: 0,
        totalLength: 0,
        endTimerDisplay: 0,
        getTimerNow: '00:00',
        paused: true,
        previousPause: false,
        sliderChange: false,
        speedList: [0.5, 1, 1.5, 2, 4],
        logs: sessionGlobal,
        frames: [],
        defaultSpeed: 1,
        transition: false,
      },
      computed: {
        length: 1,
        nowTimerDisplay: '00:00',
      },
      template: {
        'play-icon': true,
        'play-title': true,
        'sessionPlay-card': true,
        'close-btn': true,
        'text-title': true,
        'pause-icon': false,
        'time-slider': true,
        'speed-select': true,
      },
      templateText: {
        'play-title': 'Play',
        'text-title': 'Watch Session',
      },
    },
  ];

  const storeVuex = (session) => new Vuex.Store({
    namespaced: true,
    state: {
      session,
    },
    getters: {
      'sessions/get': (state) => state.session,
    },
    actions: {
      'sessions/getLogSession': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        wrapper = mount(SessionPlay, {
          store: storeVuex(test.variables.session),
          localVue,
          stubs: ['fragment'],
          propsData: {
            uid: test.props.uid,
            recorded: test.props.recorded,
            show: test.props.show,
          },
          vuetify,
          mocks: {
            $env: {
              isEnterprise: true,
            },
          },
        });
        wrapper.setData({ logs: test.variables.session });
        wrapper.setData({ paused: test.variables.paused });
        wrapper.setData({ dialog: test.variables.dialog });

        await flushPromises();
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
      // Data checking
      //////

      it('Receive data in props', () => {
        Object.keys(test.props).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.props[item]);
        });
      });
      it('Compare data with default value', () => {
        Object.keys(test.data).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.data[item]);
        });
      });
      it('Process data in the computed', () => {
        Object.keys(test.computed).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.computed[item]);
        });
      });

      //////
      // HTML validation
      //////

      it('Renders the template with data', () => {
        Object.keys(test.template).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
        });
      });
      it('Renders template with expected text', () => {
        Object.keys(test.templateText).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).text()).toContain(test.templateText[item]);
        });
      });
    });
  });
});
