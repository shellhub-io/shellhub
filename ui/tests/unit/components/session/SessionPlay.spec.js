import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SessionPlay from '@/components/session/SessionPlay';

describe('SessionPlay', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const isOwner = true;
  const uid = '8c354a00f50';
  const recorded = true;
  const speedList = [0.5, 1, 1.5, 2, 4];

  const session = [
    {
      uid: '1a0536ab37d',
      message: '\u001b]0;shellhub@shellhub: ~\u0007shellhub@shellhub:~$ ',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      time: '2020-09-24T18:32:04.559Z',
      width: 110,
      height: 23,
    },
    {
      uid: '1a0536ab37d',
      message: 'l',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      time: '2020-09-24T18:32:06.181Z',
      width: 110,
      height: 23,
    },
    {
      uid: '1a0536ab37d',
      message: 's',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      time: '2020-09-24T18:32:06.344Z',
      width: 110,
      height: 23,
    },
    {
      uid: '1a0536ab37d',
      message: '\r\n',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      time: '2020-09-24T18:32:06.592Z',
      width: 110,
      height: 23,
    },
    {
      uid: '1a0536ab37d',
      message: '\u001b]0;shellhub@shellhub: ~\u0007shellhub@shellhub:~$ ',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      time: '2020-09-24T18:32:06.61Z',
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

  beforeEach(() => {
    wrapper = shallowMount(SessionPlay, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, recorded },
      mocks: {
        $env: (isEnterprise) => isEnterprise,
      },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.uid).toEqual(uid);
    expect(wrapper.vm.recorded).toEqual(recorded);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.length).toEqual(0);
    expect(wrapper.vm.nowTimerDisplay).toEqual(0);
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
});
