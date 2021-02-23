import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import BoxMessage from '@/components/box/BoxMessage';
import Vuetify from 'vuetify';

describe('BoxMessage', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  let wrapper;

  const typeMessage = {
    session: 'session',
    firewall: 'firewall',
  };

  const items = {
    session:
    {
      icon: 'history',
      title: 'Session',
      text: [
        'An SSH session is created when a connection is made to any registered device.',
      ],
      textWithLink: [
        `If you don't know how to connect to your devices, please follow this guide
            See More.`,
      ],
    },
    firewall:
    {
      icon: 'security',
      title: 'Firewall Rule',
      text: [
        `ShellHub provides flexible firewall for filtering SSH connections.
              It gives a fine-grained control over which SSH connections reach the devices.`,
        `Using Firewall Rules you can deny or allow SSH connections from specific
              IP address to a specific or a group of devices using a given username.`,
      ],
      textWithLink: [],
    },
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
    },
    getters: {
    },
    actions: {
      'boxs/setStatus': () => {
      },
      'firewallrules/refresh': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(BoxMessage, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
      propsData: { typeMessage: typeMessage.session },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Process data in methods Session', () => {
    const sessionWrapper = shallowMount(BoxMessage, {
      localVue,
      stubs: ['fragment'],
      vuetify,
      propsData: { typeMessage: typeMessage.session },
    });

    const title = `Looks like you don't have any ${items.session.title}`;

    expect(sessionWrapper.find('[data-test="boxMessage-icon"]').text()).toEqual(items.session.icon);
    expect(sessionWrapper.find('[data-test="boxMessage-title"]').text()).toEqual(title);
    Object.keys(items.session.text).forEach((index) => {
      expect(sessionWrapper.find(`[data-test="${index}-boxMessage-text"]`).text()).toEqual(items.session.text[index]);
    });

    const lenSessionText = (items.session.text).length;
    Object.keys(items.session.textWithLink).forEach((index) => {
      expect(sessionWrapper.find(`[data-test="${lenSessionText + parseInt(index, 10)}-boxMessage-text"]`).text()).toEqual(items.session.textWithLink[index]);
    });
  });
  it('Process data in methods FirewallRule', () => {
    const firewallWrapper = shallowMount(BoxMessage, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
      propsData: { typeMessage: typeMessage.firewall },
    });

    const title = `Looks like you don't have any ${items.firewall.title}`;

    expect(firewallWrapper.find('[data-test="boxMessage-icon"]').text()).toEqual(items.firewall.icon);
    expect(firewallWrapper.find('[data-test="boxMessage-title"]').text()).toEqual(title);
    Object.keys(items.firewall.text).forEach((index) => {
      expect(firewallWrapper.find(`[data-test="${index}-boxMessage-text"]`).text()).toEqual(items.firewall.text[index]);
    });

    const lenFirewallText = (items.firewall.text).length;
    Object.keys(items.firewall.textWithLink).forEach((index) => {
      expect(firewallWrapper.find(`[data-test="${lenFirewallText + parseInt(index, 10)}-boxMessage-text"]`).text()).toEqual(items.firewall.textWithLink[index]);
    });
  });
});
