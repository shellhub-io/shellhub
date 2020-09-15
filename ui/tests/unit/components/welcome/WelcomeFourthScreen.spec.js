import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import WelcomeFourthScreen from '@/components/welcome/WelcomeFourthScreen';

describe('WelcomeFourthScreen', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const uid = 'a582b47a42d';
  const username = 'user';
  const password = 'user';

  beforeEach(() => {
    wrapper = shallowMount(WelcomeFourthScreen, {
      localVue,
      stubs: ['fragment'],
      propsData: { uid, username, password },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
