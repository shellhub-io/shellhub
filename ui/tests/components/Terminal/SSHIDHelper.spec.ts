import { DOMWrapper, flushPromises, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import SSHIDHelper from "@/components/Terminal/SSHIDHelper.vue";
import { mountComponent } from "@tests/utils/mount";

describe("SSHIDHelper", () => {
  let wrapper: VueWrapper<InstanceType<typeof SSHIDHelper>>;
  let dialog: DOMWrapper<HTMLElement>;

  const sshid = "namespace.70-85-c2-08-60-2a@staging.shellhub.io";

  const mountWrapper = (modelValue = true) => {
    wrapper = mountComponent(SSHIDHelper, {
      props: { sshid, modelValue },
      attachTo: document.body,
    });
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  it("renders the dialog when modelValue is true", () => {
    expect(dialog.text()).toContain("What is an SSHID?");
  });

  it("does not render the dialog when modelValue is false", () => {
    wrapper.unmount();
    mountWrapper(false);
    expect(dialog.text()).not.toContain("What is an SSHID?");
  });

  it("displays the SSHID explanation text", () => {
    expect(dialog.text()).toContain("The SSHID is a unique identifier");
    expect(dialog.text()).toContain("Use it in scripts, CI/CD pipelines");
  });

  it("displays all 5 SSH examples", () => {
    const expansionPanels = dialog.findAll(".v-expansion-panel");
    expect(expansionPanels).toHaveLength(5);
  });

  it.each([
    {
      index: 0,
      title: "Interactive SSH Session",
      description: "Connect to your device and get an interactive shell",
      command: `ssh <username>@${sshid}`,
    },
    {
      index: 1,
      title: "Execute Remote Command",
      description: "Run a command on the device and see the output",
      command: `ssh <username>@${sshid} "ls -la"`,
    },
    {
      index: 2,
      title: "Upload File (SCP)",
      description: "Copy a file from your local machine to the device",
      command: `scp file.txt <username>@${sshid}:/path/to/destination/`,
    },
    {
      index: 3,
      title: "Download File (SCP)",
      description: "Copy a file from the device to your local machine",
      command: `scp <username>@${sshid}:/path/to/file.txt ./`,
    },
    {
      index: 4,
      title: "Port Forwarding",
      description: "Forward a local port to a port on the device",
      command: `ssh -L 8080:localhost:80 <username>@${sshid}`,
    },
  ])("displays $title example with description and command", async ({ index, title, description, command }) => {
    expect(dialog.text()).toContain(title);
    expect(dialog.text()).toContain(description);

    // Open the expansion panel
    const panels = dialog.findAll(".v-expansion-panel");
    await panels[index].find(".v-expansion-panel-title").trigger("click");
    await flushPromises();

    // Find the CopyCommandField within this panel and verify the command
    const copyCommandField = panels[index].findComponent({ name: "CopyCommandField" });
    const input = copyCommandField.find("input");
    expect(input.element.value).toBe(command);
  });

  it("closes dialog when close button is clicked", async () => {
    const closeButton = dialog.find('[data-test="close-btn"]');
    await closeButton.trigger("click");
    await flushPromises();

    expect(wrapper.emitted("update:modelValue")).toBeTruthy();
    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
  });

  it("displays expansion panels with correct icons", () => {
    const panels = dialog.findAll(".v-expansion-panel");

    expect(panels[0].find(".v-icon").classes()).toContain("mdi-console");
    expect(panels[1].find(".v-icon").classes()).toContain("mdi-play-circle-outline");
    expect(panels[2].find(".v-icon").classes()).toContain("mdi-upload");
    expect(panels[3].find(".v-icon").classes()).toContain("mdi-download");
    expect(panels[4].find(".v-icon").classes()).toContain("mdi-lan-connect");
  });
});
