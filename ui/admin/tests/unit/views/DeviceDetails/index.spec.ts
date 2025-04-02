import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import DeviceDetails from "../../../../src/views/DeviceDetails.vue";

type DeviceDetailsWrapper = VueWrapper<InstanceType<typeof DeviceDetails>>;

const deviceDetail = {
  uid: "cb1533e2e683aec21aee89b24ac4604b1a1955930362d33fb22e4e03fac52c75",
  name: "08-97-98-68-7a-97",
  identity: {
    mac: "08:97:98:68:7a:97",
  },
  info: {
    id: "ubuntu",
    pretty_name: "Ubuntu 20.04.4 LTS",
    version: "latest",
    arch: "amd64",
    platform: "docker",
  },
  public_key: `-----BEGIN RSA PUBLIC KEY-----
  MIIBCgKCAQEAx2C95p3s9OpwHdSwV8xP5dS39jGBCM+VMChiqJViaaVoJJ2tTK/i
  zCEH6+jAuKSfvXjM3jql59RD0o7lFqd9bixiGN8/KvXZ/6hlDrdKniatIGHmGw6z
  N9EfKbTqJh0vHX/yRzRWlfAlLHoWjg0lV+Y6RpAiV1u6Gd4ZnDyz62u82fpQYqLu
  IFrhfOP52qbVZHMT6Vn/q8U26wysrDlVbF1k8RDR79Ib9i1Bu3mBPn0r5AEJOpQQ
  NqODe3Wjntgy8i0/iFaUV+9K17u50Pmm4uPfVfMEPmZSXpAwfpgWFPFInA9hLefq
  9XLjOj93MwVWN4iXLbLOoLI/9MQw5zZSYQIDAQAB
  -----END RSA PUBLIC KEY-----
  `,
  tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  last_seen: "2022-06-06T18:51:53.813Z",
  online: true,
  namespace: "dev",
  status: "accepted",
  created_at: "2022-04-13T11:43:25.218Z",
  remote_addr: "172.22.0.1",
  position: {
    latitude: 0,
    longitude: 0,
  },
  tags: ["dev"],
};

const mockRoute = {
  params: {
    id: "cb1533e2e683aec21aee89b24ac4604b1a1955930362d33fb22e4e03fac52c75",
  },
};

describe("Device Details", () => {
  const store = createStore({
    state: {
      device: deviceDetail,
    },
    getters: {
      "devices/get": () => deviceDetail,
    },
    actions: {
      "devices/get": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });
  let wrapper: DeviceDetailsWrapper;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(DeviceDetails, {
      global: {
        plugins: [[store, key], vuetify, routes],
        mocks: {
          $route: mockRoute,
        },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Has the correct data", async () => {
    expect(wrapper.vm.device).toEqual(deviceDetail);
  });

  it("Render the correct title", () => {
    expect(wrapper.find("h1").text()).toEqual("Device Details");
  });

  it("Should render the props of the user in the Screen", () => {
    expect(wrapper.find(`[data-test='${deviceDetail.uid}']`).text()).toContain(deviceDetail.uid);
    expect(wrapper.find(`[data-test='${deviceDetail.name}']`).text()).toContain(deviceDetail.name);
    expect(wrapper.find(`[data-test='${deviceDetail.identity.mac}']`).text()).toContain(deviceDetail.identity.mac);
    expect(wrapper.find(`[data-test='${deviceDetail.info.id}']`).text()).toContain(deviceDetail.info.id);
    expect(wrapper.find(`[data-test='${deviceDetail.tenant_id}']`).text()).toContain(deviceDetail.tenant_id);
    expect(wrapper.find(`[data-test='${deviceDetail.online}']`).text()).toContain(deviceDetail.online);
    expect(wrapper.find(`[data-test='${deviceDetail.tags}']`).text()).toContain(deviceDetail.tags);
    expect(wrapper.find(`[data-test='${deviceDetail.namespace}']`).text()).toContain(deviceDetail.namespace);
    expect(wrapper.find(`[data-test='${deviceDetail.status}']`).text()).toContain(deviceDetail.status);
  });
});
