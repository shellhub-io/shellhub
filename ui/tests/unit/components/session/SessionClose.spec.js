import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SessionClose from '@/components/session/SessionClose';

describe('SessionClose', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  const uid = 'cfe65e6a847ffafd62de34fc75c1faf0ebdbb477d';
  const device = 'a1aa7d06927cc3b30cdf0d0db8c39b488891576';

  const store = new Vuex.Store({
    namespaced: true,
    actions: {
      'sessions/close': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(SessionClose, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, device },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
