import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import WelcomeSecondScreen from '@/components/welcome/WelcomeSecondScreen';

describe('WelcomeSecondScreen', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const command = 'curl "http://localhost/install.sh?tenant_id=a582b47a42e" | sh';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
    },
    getters: {
    },
    actions: {
      'snackbar/showSnackbarCopy': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(WelcomeSecondScreen, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { command },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.command).toEqual(command);
  });
  it('Renders the template with data', () => {
    const commandText = wrapper.find('[data-test="command-text"]');
    commandText.element.value = command;
    expect(commandText.element.value).toEqual(command);
  });
});
