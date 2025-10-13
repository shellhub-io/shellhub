import { setActivePinia, createPinia } from "pinia";
import { mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it } from "vitest";
import WelcomeSecondScreen from "@/components/Welcome/WelcomeSecondScreen.vue";
import useAuthStore from "@/store/modules/auth";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("Welcome Second Screen", () => {
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  authStore.tenantId = "fake-tenant-data";

  const wrapper = mount(WelcomeSecondScreen, { global: { plugins: [createVuetify(), SnackbarPlugin] } });

  it("Renders the component", () => { expect(wrapper.html()).toMatchSnapshot(); });

  it("Renders the correct command", async () => {
    const expectedCommand = "curl -sSf http://localhost:3000/install.sh | TENANT_ID=fake-tenant-data SERVER_ADDRESS=http://localhost sh";
    expect(wrapper.find('[data-test="command-field"] input').attributes("value")).toBe(expectedCommand);
  });
});
