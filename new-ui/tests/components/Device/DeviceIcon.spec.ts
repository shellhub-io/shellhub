import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import DeviceIcon from "../../../src/components/Devices/DeviceIcon.vue";
import routes from "../../../src/router";

const iconName = "alpine";
const defaultIcon = "fl-alpine";

const iconsMap = {
  alpine: "fl-alpine",
  arch: "fl-archlinux",
  centos: "fl-centos",
  coreos: "fl-coreos",
  debian: "fl-debian",
  devuan: "fl-devuan",
  elementary: "fl-elementary",
  fedora: "fl-fedora",
  freebsd: "fl-freebsd",
  gentoo: "fl-gentoo",
  linuxmint: "fl-linuxmint",
  mageia: "fl-mageia",
  manjaro: "fl-manjaro",
  mandriva: "fl-mandriva",
  nixos: "fl-nixos",
  opensuse: "fl-opensuse",
  rhel: "fl-redhat",
  sabayon: "fl-sabayon",
  slackware: "fl-slackware",
  ubuntu: "fl-ubuntu",
  raspbian: "fl-raspberry-pi",
  "ubuntu-core": "fl-ubuntu",
  ubuntucore: "fl-ubuntu",
  void: "fl-void",
};

describe("DeviceIcon", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();
    wrapper = mount(DeviceIcon, {
      global: {
        plugins: [vuetify, routes],
        stubs: ["router-link"],
      },
      props: {
        icon: iconName,
      },
      data() {
        return {
          deviceIcon: iconsMap,
        };
      },
    });
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

  it("Receive data in props", () => {
    expect(wrapper.props("icon")).toBe(iconName);
  });

  it("Compare data with default value", () => {
    expect(wrapper.vm.deviceIcon).toEqual(iconsMap);
  });

  //////
  // HTML validation
  //////

  it("Renders the template with data", () => {
    const vuetify = createVuetify();

    wrapper = mount(DeviceIcon, {
      global: {
        plugins: [vuetify, routes],
        stubs: ["router-link"],
      },
      props: {
        icon: "alpine",
      },
      data() {
        return {
          deviceIcon: iconsMap,
        };
      },
      computed: {
        iconName: () => defaultIcon,
      },
      shallow: true,
    });
    const wrapperClasses = wrapper.find('[data-test="type-icon"]').classes();
    expect(wrapperClasses[0]).toBe(defaultIcon);
  });

  //////
  // In this case, the other icons are tested.
  //////

  Object.keys(iconsMap).forEach((iconKey) => {
    const vuetify = createVuetify();

    wrapper = mount(DeviceIcon, {
      global: {
        plugins: [vuetify, routes],
        stubs: ["router-link"],
      },
      props: { icon: iconKey },
      shallow: true,
    });
    const wrapperClasses = wrapper.find('[data-test="type-icon"]').classes();

    expect(wrapperClasses[0]).toBe(
      iconsMap[iconKey],
    );
  });
});

describe("DeviceIcon icon not found", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();
    wrapper = mount(DeviceIcon, {
      global: {
        plugins: [vuetify, routes],
        stubs: ["router-link"],
      },
      props: {
        icon: "not-found",
      },
      data() {
        return {
          deviceIcon: iconsMap,
        };
      },
    });
  });

  ///////
  // Data and Props checking
  //////

  it("Receive data in props", () => {
    expect(wrapper.props("icon")).toBe("not-found");
  });

  it("Compare data with default value", () => {
    expect(wrapper.vm.deviceIcon).toEqual(iconsMap);
  });

  it("Renders the template with data", () => {
    const icon = wrapper.find('[data-test="type-icon"]');
    expect(icon.exists()).toBeTruthy();
    expect(icon.classes()).toContain("fl-tux");
  });
});
