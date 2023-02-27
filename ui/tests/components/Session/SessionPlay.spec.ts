import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import SessionPlay from "../../../src/components/Sessions/SessionPlay.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";
import { envVariables } from "../../../src/envVariables";

const sessionGlobal = [
  {
    uid: "1a0536ab",
    message: "\u001b]0;shellhub@shellhub: ~\u0007shellhub@shellhub:~$ ",
    tenant_id: "xxxxxxxx",
    time: "2020-09-24T18:32:04.559Z",
    width: 110,
    height: 23,
  },
];

const tests = [
  {
    description: "Dialog closed",
    variables: {
      session: [],
      recorded: false,
      paused: false,
      dialog: false,
    },
    props: {
      uid: "8c354a00",
      recorded: false,
      notHasAuthorization: false,
    },
    data: {
      showDialog: false,
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
      "play-icon": true,
      "play-title": true,
      "sessionPlay-card": false,
    },
    templateText: {
      "play-title": "Play",
    },
  },
  {
    description: "Dialog opened with play paused",
    variables: {
      session: sessionGlobal,
      recorded: true,
      paused: false,
      dialog: true,
    },
    props: {
      uid: "8c354a00",
      recorded: true,
      notHasAuthorization: false,
    },
    data: {
      showDialog: true,
      currentTime: 0,
      totalLength: 0,
      endTimerDisplay: 0,
      getTimerNow: 0,
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
      nowTimerDisplay: "00:00",
    },
    template: {
      "play-icon": true,
      "play-title": true,
      "sessionPlay-card": true,
      "close-btn": true,
      "text-title": true,
      "pause-icon": true,
      "time-slider": true,
      "speed-select": true,
    },
    templateText: {
      "play-title": "Play",
      "text-title": "Watch Session",
    },
  },
  {
    description: "Dialog opened with play not paused",
    variables: {
      session: sessionGlobal,
      recorded: true,
      paused: false,
      show: true,
      dialog: true,
    },
    props: {
      uid: "8c354a00",
      recorded: true,
      notHasAuthorization: false,
    },
    data: {
      showDialog: true,
      currentTime: 0,
      totalLength: 0,
      endTimerDisplay: 0,
      getTimerNow: 0,
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
      nowTimerDisplay: "00:00",
    },
    template: {
      "play-icon": true,
      "play-title": true,
      "sessionPlay-card": true,
      "close-btn": true,
      "text-title": true,
      "pause-icon": false,
      "time-slider": true,
      "speed-select": true,
    },
    templateText: {
      "play-title": "Play",
      "text-title": "Watch Session",
    },
  },
];

const store = (session: typeof sessionGlobal) => createStore({
  state: {
    session,
  },
  getters: {
    "sessions/get": (state) => state.session,
  },
  actions: {
    "sessions/getLogSession": vi.fn(),

    "snackbar/showSnackbarErrorLoading": vi.fn(),
  },
});

describe("SessionPlay", () => {
  let wrapper:VueWrapper<InstanceType<typeof SessionPlay>>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        wrapper = mount(SessionPlay, {
          global: {
            plugins: [[store(test.variables.session), key], routes, vuetify],
          },
          props: {
            uid: test.props.uid,
            recorded: test.props.recorded,
            notHasAuthorization: test.props.notHasAuthorization,
          },
          shallow: true,
        });
        wrapper.vm.logs = test.variables.session;
        wrapper.vm.paused = test.variables.paused;

        await flushPromises();

        envVariables.isEnterprise = true;
      });

      ///////
      // Component Rendering
      //////

      it("Is a Vue instance", () => {
        expect(wrapper).toBeTruthy();
      });
      it("Renders the component", () => {
        expect(wrapper.html()).toMatchSnapshot();
      });

      ///////
      // Data and Props checking
      //////

      it("Data is defined", () => {
        expect(wrapper.vm.$data).toBeDefined();
      });
      it("Receive data in props", () => {
        expect(wrapper.vm.recorded).toEqual(test.props.recorded);
        expect(wrapper.vm.uid).toEqual(test.props.uid);
        expect(wrapper.vm.notHasAuthorization).toEqual(test.props.notHasAuthorization);
      });
      it("Compare data with default value", () => {
        expect(wrapper.vm.currentTime).toEqual(test.data.currentTime);
        expect(wrapper.vm.totalLength).toEqual(test.data.totalLength);
        expect(wrapper.vm.endTimerDisplay).toEqual(test.data.endTimerDisplay);
        expect(wrapper.vm.getTimerNow).toEqual(test.data.getTimerNow);
        expect(wrapper.vm.paused).toEqual(test.data.paused);
        expect(wrapper.vm.previousPause).toEqual(test.data.previousPause);
        expect(wrapper.vm.sliderChange).toEqual(test.data.sliderChange);
        expect(wrapper.vm.speedList).toEqual(test.data.speedList);
        expect(wrapper.vm.logs).toEqual(test.data.logs);
        expect(wrapper.vm.frames).toEqual(test.data.frames);
        expect(wrapper.vm.defaultSpeed).toEqual(test.data.defaultSpeed);
        expect(wrapper.vm.transition).toEqual(test.data.transition);
      });

      // todo html
    });
  });
});
