import { mount, flushPromises, VueWrapper } from "@vue/test-utils";
import { describe, beforeEach, vi, it, expect } from "vitest";
import { createVuetify } from "vuetify";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";
import { store, key } from "@/store";
import Player from "@/components/Sessions/Player.vue";

type PlayerWrapper = VueWrapper<InstanceType<typeof Player>>;

describe("Asciinema Player", () => {
  let wrapper: PlayerWrapper;

  const vuetify = createVuetify();

  // eslint-disable-next-line vue/max-len
  const logsMock = "{\"version\": 2, \"width\": 80, \"height\": 24}\n[0.123, \"r\", \"80x24\"]\n[1.0, \"o\", \"Asciinema player test\"]\n[2.0, \"o\", \"logout\"]";

  const authData = {
    status: "success",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant",
    email: "test@test.com",
    id: "507f1f77bcf86cd799439011",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  const session = true;

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant");

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("security/setSecurity", session);

    wrapper = mount(Player, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
        config: {
          errorHandler: () => { /* ignore global error handler */ },
        },
      },
      props: {
        logs: logsMock,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Renders components", async () => {
    await flushPromises();
    expect(wrapper.find('[data-test="player-container"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="player-controls"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="pause-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="play-btn"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="playback-time"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="time-slider"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="speed-select"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="shortcuts-btn"]').exists()).toBe(true);
  });

  it("Creates player on mount", async () => {
    await flushPromises();
    expect(wrapper.vm.player).toBeDefined();
  });

  // it("Fails to Change Password", async () => {
  //   mockUser.onPatch("http://localhost:3000/api/users").reply(403);

  //   const StoreSpy = vi.spyOn(store, "dispatch");

  //   wrapper.vm.show = true;
  //   await flushPromises();

  //   await wrapper.findComponent('[data-test="password-input"]').setValue("xxxxxx");
  //   await wrapper.findComponent('[data-test="new-password-input"]').setValue("x1x2x3");
  //   await wrapper.findComponent('[data-test="confirm-new-password-input"]').setValue("x1x2x3");

  //   await wrapper.findComponent('[data-test="change-password-btn"]').trigger("click");
  //   await flushPromises();

  //   expect(StoreSpy).toHaveBeenCalledWith("users/patchPassword", {
  //     name: "test",
  //     username: undefined,
  //     email: "test@test.com",
  //     recovery_email: undefined,
  //     currentPassword: "xxxxxx",
  //     newPassword: "x1x2x3",
  //   });

  //   expect(StoreSpy).toHaveBeenCalledWith("snackbar/showSnackbarErrorDefault");
  // });
});
