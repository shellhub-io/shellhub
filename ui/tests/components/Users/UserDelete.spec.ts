import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import UserDelete from "@/components/User/UserDelete.vue";
import { usersApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import { router } from "@/router";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

type UserDeleteWrapper = VueWrapper<InstanceType<typeof UserDelete>>;

describe("User Delete", () => {
  let wrapper: UserDeleteWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const authStore = useAuthStore();
  const mockUsersApi = new MockAdapter(usersApi.getAxios());

  beforeEach(async () => {
    wrapper = mount(UserDelete, {
      global: {
        plugins: [router, vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", async () => {
    wrapper.vm.showDialog = true;
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="user-delete-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="delete-user-btn"]').exists()).toBe(true);
  });

  it("Successfully Delete User", async () => {
    mockUsersApi.onDelete("http://localhost:3000/api/user").reply(200);

    const storeSpy = vi.spyOn(authStore, "deleteUser");

    wrapper.vm.showDialog = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="delete-user-btn"]').trigger("click");

    expect(storeSpy).toHaveBeenCalled();
  });

  it("Fails to add Delete User", async () => {
    mockUsersApi.onDelete("http://localhost:3000/api/user").reply(400);

    wrapper.vm.showDialog = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="delete-user-btn"]').trigger("click");
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete account.");
  });
});
