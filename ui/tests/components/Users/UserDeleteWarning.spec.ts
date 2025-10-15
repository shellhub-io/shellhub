import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, flushPromises, DOMWrapper } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import UserDeleteWarning from "@/components/User/UserDeleteWarning.vue";
import useAuthStore from "@/store/modules/auth";
import { envVariables } from "@/envVariables";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

const mockSnackbar = {
  showError: vi.fn(),
};

describe("UserDeleteWarning", () => {
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const vuetify = createVuetify();
  authStore.username = "testuser";
  envVariables.isCommunity = true;

  it("Shows community instructions and CLI command", async () => {
    mount(UserDeleteWarning, {
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: { modelValue: true },
    });
    const dialog = new DOMWrapper(document.body);

    const cliCommand = dialog.find('[data-test="copy-command-field"] input');
    await flushPromises();

    expect(dialog.text()).toContain("Community instances");
    expect(cliCommand.attributes("value")).toContain("./bin/cli user delete testuser");
    expect(dialog.find('[data-test="docs-link"]').exists()).toBe(true);
  });

  it("Shows enterprise instructions when not community", async () => {
    envVariables.isCommunity = false;
    envVariables.isEnterprise = true;

    mount(UserDeleteWarning, {
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: { modelValue: true },
    });

    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    expect(dialog.text()).toContain("Enterprise instances");
    expect(dialog.text()).toContain("Admin Console");
  });
});
