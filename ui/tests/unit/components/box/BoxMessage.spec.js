import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import BoxMessage from '@/components/box/BoxMessage';

describe('BoxMessage', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  let wrapper;

  const typeMessage = {
    device: 'device',
    session: 'session',
    firewall: 'firewall',
    publicKey: 'publicKey',
  };

  const items = {
    device:
    {
      icon: 'devices',
      title: 'Device',
      text: [
        'In order to register a device on ShellHub, you need to install ShellHub agent onto it.',
      ],
      textWithLink: [
        `The easiest way to install ShellHub agent is with our automatic one-line installation
          script, which works with all Linux distributions that have Docker installed and
          properly set up See More.`,
      ],
    },
    session:
    {
      icon: 'history',
      title: 'Session',
      text: [
        'An SSH session is created when a connection is made to any registered device.',
      ],
      textWithLink: [
        `Please follow our guide on how to connect to your devices
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
    publicKey:
    {
      icon: 'vpn_key',
      title: 'Public Keys',
      text: [
        'You can connect to your devices using password-based logins, but we strongly recommend using SSH key pairs instead.',
        'SSH keys are more secure than passwords and can help you log in without having to remember long passwords.',
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
      'publickeys/refresh': () => {
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
  it('Process data in methods Device', () => {
    wrapper = shallowMount(BoxMessage, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
      propsData: { typeMessage: typeMessage.device },
    });

    const title = `Looks like you don't have any ${items.device.title}`;
    const lenDeviceText = (items.device.text).length;

    expect(wrapper.find('[data-test="boxMessage-icon"]').text()).toEqual(items.device.icon);
    expect(wrapper.find('[data-test="boxMessage-title"]').text()).toEqual(title);
    Object.keys(items.device.text).forEach((index) => {
      expect(wrapper.find(`[data-test="${index}-boxMessage-text"]`).text()).toEqual(items.device.text[index]);
    });
    Object.keys(items.firewall.textWithLink).forEach((index) => {
      expect(wrapper.find(`[data-test="${lenDeviceText + parseInt(index, 10)}-boxMessage-text"]`).text()).toEqual(items.device.textWithLink[index]);
    });
  });
  it('Process data in methods Session', () => {
    wrapper = shallowMount(BoxMessage, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
      propsData: { typeMessage: typeMessage.session },
    });

    const title = `Looks like you don't have any ${items.session.title}`;
    const lenSessionText = (items.session.text).length;

    expect(wrapper.find('[data-test="boxMessage-icon"]').text()).toEqual(items.session.icon);
    expect(wrapper.find('[data-test="boxMessage-title"]').text()).toEqual(title);
    Object.keys(items.session.text).forEach((index) => {
      expect(wrapper.find(`[data-test="${index}-boxMessage-text"]`).text()).toEqual(items.session.text[index]);
    });
    Object.keys(items.session.textWithLink).forEach((index) => {
      expect(wrapper.find(`[data-test="${lenSessionText + parseInt(index, 10)}-boxMessage-text"]`).text()).toEqual(items.session.textWithLink[index]);
    });
  });
  it('Process data in methods FirewallRule', () => {
    wrapper = shallowMount(BoxMessage, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
      propsData: { typeMessage: typeMessage.firewall },
    });

    const title = `Looks like you don't have any ${items.firewall.title}`;
    const lenFirewallText = (items.firewall.text).length;

    expect(wrapper.find('[data-test="boxMessage-icon"]').text()).toEqual(items.firewall.icon);
    expect(wrapper.find('[data-test="boxMessage-title"]').text()).toEqual(title);
    Object.keys(items.firewall.text).forEach((index) => {
      expect(wrapper.find(`[data-test="${index}-boxMessage-text"]`).text()).toEqual(items.firewall.text[index]);
    });
    Object.keys(items.firewall.textWithLink).forEach((index) => {
      expect(wrapper.find(`[data-test="${lenFirewallText + parseInt(index, 10)}-boxMessage-text"]`).text()).toEqual(items.firewall.textWithLink[index]);
    });
  });
  it('Process data in methods Public Key', () => {
    wrapper = shallowMount(BoxMessage, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
      propsData: { typeMessage: typeMessage.publicKey },
    });

    const title = `Looks like you don't have any ${items.publicKey.title}`;
    const lenPublicKeyText = (items.publicKey.text).length;

    expect(wrapper.find('[data-test="boxMessage-icon"]').text()).toEqual(items.publicKey.icon);
    expect(wrapper.find('[data-test="boxMessage-title"]').text()).toEqual(title);
    Object.keys(items.publicKey.text).forEach((index) => {
      expect(wrapper.find(`[data-test="${index}-boxMessage-text"]`).text()).toEqual(items.publicKey.text[index]);
    });
    Object.keys(items.firewall.textWithLink).forEach((index) => {
      expect(wrapper.find(`[data-test="${lenPublicKeyText + parseInt(index, 10)}-boxMessage-text"]`).text()).toEqual(items.publicKey.textWithLink[index]);
    });
  });
});
