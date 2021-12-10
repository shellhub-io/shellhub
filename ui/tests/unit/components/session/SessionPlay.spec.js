import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import SessionPlay from '@/components/session/SessionPlay';
import { actions, authorizer } from '../../../../src/authorizer';

describe('SessionPlay', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const role = ['owner', 'operator'];

  const hasAuthorization = {
    owner: true,
    operator: false,
  };

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
      description: 'Icon',
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
        action: 'play',
      },
      computed: {
        length: 0,
        nowTimerDisplay: 0,
      },
      template: {
        'sessionPlay-card': false,
        'close-btn': false,
        'pause-icon': false,
        'play-icon': false,
        'time-slider': false,
        'speed-select': false,
      },
    },
    {
      description: 'Dialog play paused',
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
        action: 'play',
      },
      computed: {
        length: 1,
        nowTimerDisplay: '00:00',
      },
      template: {
        'sessionPlay-card': true,
        'close-btn': true,
        'pause-icon': true,
        'play-icon': true,
        'time-slider': true,
        'speed-select': true,
      },
    },
    {
      description: 'Dialog play not paused',
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
        action: 'play',
      },
      computed: {
        length: 1,
        nowTimerDisplay: '00:00',
      },
      template: {
        'sessionPlay-card': true,
        'close-btn': true,
        'pause-icon': false,
        'play-icon': true,
        'time-slider': true,
        'speed-select': true,
      },
    },
  ];

  const storeVuex = (session, currentrole) => new Vuex.Store({
    namespaced: true,
    state: {
      session,
      currentrole,
    },
    getters: {
      'sessions/get': (state) => state.session,
      'auth/role': (state) => state.currentrole,
    },
    actions: {
      'sessions/getLogSession': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(async () => {
          wrapper = mount(SessionPlay, {
            store: storeVuex(test.variables.session, currentrole),
            localVue,
            stubs: ['fragment'],
            propsData: {
              uid: test.props.uid,
              recorded: test.props.recorded,
              show: test.props.show,
            },
            vuetify,
            mocks: {
              $authorizer: authorizer,
              $actions: actions,
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
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentrole]);
        });

        //////
        // HTML validation
        //////

        it('Renders the template with data', () => {
          if (hasAuthorization[currentrole]) {
            Object.keys(test.template).forEach((item) => {
              expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
            });
          } else if (!test.props.show) {
            Object.keys(test.template).forEach((item) => {
              expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
            });
          }
        });
      });
    });
  });
});
