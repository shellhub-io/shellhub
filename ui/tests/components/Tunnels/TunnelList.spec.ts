import { setActivePinia, createPinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import { createRouter, createWebHistory } from "vue-router";
import { nextTick } from "vue";
import { store, key } from "@/store";
import TunnelList from "@/components/Tunnels/TunnelList.vue";
import { routes } from "@/router";
import { tunnelApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type TunnelListWrapper = VueWrapper<InstanceType<typeof TunnelList>>;

const tunnelResponse = [{
  address: "9a8df9321368d567cfac8679cec7848c",
  namespace: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
  device: "13b0c8ea878e61ff849db69461795006a9594c8f6a6390ce0000100b0c9d7d0a",
  host: "127.0.0.1",
  port: 8080,
}];

describe("Tunnel List", () => {
  let wrapper: TunnelListWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockTunnelsApi = new MockAdapter(tunnelApi.getAxios());

  const router = createRouter({
    history: createWebHistory(),
    routes,
  });

  beforeEach(async () => {
    router.push("/devices/fake-uid");
    mockTunnelsApi.onGet("http://localhost:3000/api/devices/fake-uid/tunnels").reply(200, tunnelResponse);

    wrapper = mount(TunnelList, {
      global: {
        plugins: [[store, key], vuetify, [router], SnackbarPlugin],

      },
      props: {
        deviceUid: "a582b47a42d",
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the table", () => {
    expect(wrapper.find('[data-test="device-tunnels-table"]').exists()).toBe(true);
  });

  it("Renders table headers", () => {
    const headers = wrapper.findAll('[data-test^="device-tunnels-header-"]');
    expect(headers.length).toBe(5);
    expect(headers[0].text()).toBe("Address");
    expect(headers[1].text()).toBe("Host");
    expect(headers[2].text()).toBe("Port");
    expect(headers[3].text()).toBe("Expiration Date");
    expect(headers[4].text()).toBe("Actions");
  });

  it("Renders table rows", async () => {
    const rows = wrapper.findAll('[data-test^="device-tunnel-row-"]');
    rows.forEach((row, index) => {
      expect(row.find("[data-test=\"device-tunnel-url\"]").exists()).toBe(true);
      expect(row.find("[data-test=\"device-tunnel-host\"]").text()).toBe(tunnelResponse[index].host);
      expect(row.find("[data-test=\"device-tunnel-port\"]").text()).toBe(`${tunnelResponse[index].port}`);
      expect(row.find("[data-test^=\"device-tunnel-delete-\"]").exists()).toBe(true);
    });
  });

  it("Renders empty state if no tunnels", async () => {
    await wrapper.setProps({ deviceUid: "fake-uid" });
    mockTunnelsApi.onGet("http://localhost:3000/api/devices/fake-uid/tunnels").reply(200, []);
    await wrapper.vm.getTunnels();
    await nextTick();

    expect(wrapper.find('[data-test="device-tunnels-empty"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-tunnels-empty"]').text()).toContain("No data available");
  });
});
