import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import BoxMessage from "../../../src/components/Box/BoxMessage.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "boxs/setStatus": vi.fn(),
    "firewallrules/refresh": vi.fn(),
    "publickeys/refresh": vi.fn(),
    "snackbar/showSnackbarErrorLoading": vi.fn(),
  },
});

describe("BoxMessage Render", () => {
  let wrapper: VueWrapper<InstanceType<typeof BoxMessage>>;

  const typeMessage = {
    device: "device",
    session: "session",
    firewall: "firewall",
    publicKey: "publicKey",
  };

  const items = {
    device: {
      icon: "mdi-cellphone-link",
      title: "Device",
      text: [
        "In order to register a device on ShellHub, you need to install ShellHub agent onto it.",
      ],
      textWithLink: [
        `The easiest way to install ShellHub agent is with our automatic one-line installation
          script, which works with all Linux distributions that have Docker installed and
          properly set up See More.`,
      ],
    },
    session: {
      icon: "mdi-history",
      title: "Session",
      text: [
        "An SSH session is created when a connection is made to any registered device.",
      ],
      textWithLink: [
        `Please follow our guide on how to connect to your devices
            See More.`,
      ],
    },
    firewall: {
      icon: "mdi-security",
      title: "Firewall Rule",
      text: [
        `ShellHub provides flexible firewall for filtering SSH connections.
              It gives a fine-grained control over which SSH connections reach the devices.`,
        `Using Firewall Rules you can deny or allow SSH connections from specific
              IP address to a specific or a group of devices using a given username.`,
      ],
      textWithLink: [],
    },
    publicKey: {
      icon: "mdi-key",
      title: "Public Keys",
      text: [
        "You can connect to your devices using password-based logins, but we strongly recommend using SSH key pairs instead.",
        "SSH keys are more secure than passwords and can help you log in without having to remember long passwords.",
      ],
      textWithLink: [],
    },
  };

  it("Is a Vue instance", () => {
    const vuetify = createVuetify();

    wrapper = mount(BoxMessage, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      props: {
        typeMessage: typeMessage.device,
      },
    });
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    const vuetify = createVuetify();

    wrapper = mount(BoxMessage, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      props: {
        typeMessage: typeMessage.device,
      },
    });
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Process data in methods Device", () => {
    const vuetify = createVuetify();

    wrapper = mount(BoxMessage, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      props: {
        typeMessage: typeMessage.device,
      },
    });

    const title = `Looks like you don't have any ${items.device.title}`;
    const lenDeviceText = items.device.text.length;

    expect(wrapper.find('[data-test="boxMessage-icon"]').classes()[0]).toEqual(
      items.device.icon,
    );
    expect(wrapper.find('[data-test="boxMessage-title"]').text()).toEqual(
      title,
    );

    Object.keys(items.device.text).forEach((index: string) => {
      expect(
        wrapper.find(`[data-test="${index}-boxMessage-text"]`).text(),
      ).toEqual(items.device.text[index]);
    });
    Object.keys(items.firewall.textWithLink).forEach((index: string) => {
      expect(
        wrapper
          .find(
            `[data-test="${
              lenDeviceText + parseInt(index, 10)
            }-boxMessage-text"]`,
          )
          .text(),
      ).toEqual(items.device.textWithLink[index]);
    });
  });

  it("Process data in methods Session", () => {
    const vuetify = createVuetify();

    wrapper = mount(BoxMessage, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      props: {
        typeMessage: typeMessage.session,
      },
    });

    const title = `Looks like you don't have any ${items.session.title}`;
    const lenSessionText = items.session.text.length;

    expect(wrapper.find('[data-test="boxMessage-icon"]').classes()[0]).toEqual(
      items.session.icon,
    );
    expect(wrapper.find('[data-test="boxMessage-title"]').text()).toEqual(
      title,
    );

    Object.keys(items.session.text).forEach((index: string) => {
      expect(
        wrapper.find(`[data-test="${index}-boxMessage-text"]`).text(),
      ).toEqual(items.session.text[index]);
    });
    Object.keys(items.firewall.textWithLink).forEach((index: string) => {
      expect(
        wrapper
          .find(
            `[data-test="${
              lenSessionText + parseInt(index, 10)
            }-boxMessage-text"]`,
          )
          .text(),
      ).toEqual(items.device.textWithLink[index]);
    });
  });

  it("Process data in methods Firewall", () => {
    const vuetify = createVuetify();

    wrapper = mount(BoxMessage, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      props: {
        typeMessage: typeMessage.firewall,
      },
    });

    const title = `Looks like you don't have any ${items.firewall.title}`;
    const lenFirewallText = items.firewall.text.length;

    expect(wrapper.find('[data-test="boxMessage-icon"]').classes()[0]).toEqual(
      items.firewall.icon,
    );
    expect(wrapper.find('[data-test="boxMessage-title"]').text()).toEqual(
      title,
    );

    Object.keys(items.firewall.text).forEach((index: string) => {
      expect(
        wrapper.find(`[data-test="${index}-boxMessage-text"]`).text(),
      ).toEqual(items.firewall.text[index]);
    });
    Object.keys(items.firewall.textWithLink).forEach((index: string) => {
      expect(
        wrapper
          .find(
            `[data-test="${
              lenFirewallText + parseInt(index, 10)
            }-boxMessage-text"]`,
          )
          .text(),
      ).toEqual(items.firewall.textWithLink[index]);
    });
  });

  it("Process data in methods PublicKey", () => {
    const vuetify = createVuetify();

    wrapper = mount(BoxMessage, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      props: {
        typeMessage: typeMessage.publicKey,
      },
    });

    const title = `Looks like you don't have any ${items.publicKey.title}`;
    const lenPublicKeyText = items.publicKey.text.length;

    expect(wrapper.html()).toMatchSnapshot();

    expect(wrapper.find('[data-test="boxMessage-icon"]').classes()[0]).toEqual(
      items.publicKey.icon,
    );
    expect(wrapper.find('[data-test="boxMessage-title"]').text()).toEqual(
      title,
    );

    Object.keys(items.publicKey.text).forEach((index: string) => {
      expect(
        wrapper.find(`[data-test="${index}-boxMessage-text"]`).text(),
      ).toEqual(items.publicKey.text[index]);
    });

    Object.keys(items.publicKey.textWithLink).forEach((index: string) => {
      expect(
        wrapper
          .find(
            `[data-test="${
              lenPublicKeyText + parseInt(index, 10)
            }-boxMessage-text"]`,
          )
          .text(),
      ).toEqual(items.publicKey.textWithLink[index]);
    });
  });
});
