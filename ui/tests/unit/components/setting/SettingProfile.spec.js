import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SettingProfile from '@/components/setting/SettingProfile';

describe('SettingProfile', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const store = new Vuex.Store({
    namespaced: true,
    actions: {
      'users/put': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(SettingProfile, {
      store,
      localVue,
      stubs: ['fragment'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
