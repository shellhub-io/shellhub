<template>
  <WindowDialog
    v-model="showDialog"
    transition="dialog-bottom-transition"
    title="What is an SSHID?"
    icon="mdi-ssh"
    @close="close"
  >
    <v-card-text class="pa-6">
      <p class="text-body-2 mb-4">
        The SSHID is a unique identifier that allows you to connect to your device from anywhere.
        Use it in scripts, CI/CD pipelines, automations, or your local terminal.
        It works exactly like traditional SSH.
      </p>

      <v-divider class="mb-4" />

      <div class="text-subtitle-2 font-weight-bold mb-3">
        Examples
      </div>

      <v-expansion-panels
        variant="accordion"
        elevation="0"
        class="bg-v-theme-surface"
      >
        <v-expansion-panel
          v-for="(example, index) in examples"
          :key="index"
          class="bg-v-theme-surface"
        >
          <v-expansion-panel-title>
            <div class="d-flex align-center w-100">
              <v-icon
                :icon="example.icon"
                size="large"
                class="mr-3"
                color="primary"
              />
              <div class="flex-grow-1">
                <div class="text-subtitle-1 font-weight-medium">
                  {{ example.title }}
                </div>
                <div class="text-body-2 text-medium-emphasis">
                  {{ example.description }}
                </div>
              </div>
            </div>
          </v-expansion-panel-title>
          <v-expansion-panel-text>
            <div class="pa-4">
              <CopyCommandField
                :command="example.command"
                :hide-details="true"
                density="compact"
              />
            </div>
          </v-expansion-panel-text>
        </v-expansion-panel>
      </v-expansion-panels>
    </v-card-text>

    <template #footer>
      <v-card-actions class="d-flex justify-end w-100">
        <v-btn
          data-test="close-btn"
          @click="close"
        >
          Close
        </v-btn>
      </v-card-actions>
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";
import CopyCommandField from "@/components/CopyCommandField.vue";

interface Props {
  sshid: string;
}

const props = defineProps<Props>();

const showDialog = defineModel<boolean>({ required: true });

const username = "<username>";

const examples = computed(() => [
  {
    title: "Interactive SSH Session",
    description: "Connect to your device and get an interactive shell",
    icon: "mdi-console",
    command: `ssh ${username}@${props.sshid}`,
  },
  {
    title: "Execute Remote Command",
    description: "Run a command on the device and see the output",
    icon: "mdi-play-circle-outline",
    command: `ssh ${username}@${props.sshid} "ls -la"`,
  },
  {
    title: "Upload File (SCP)",
    description: "Copy a file from your local machine to the device",
    icon: "mdi-upload",
    command: `scp file.txt ${username}@${props.sshid}:/path/to/destination/`,
  },
  {
    title: "Download File (SCP)",
    description: "Copy a file from the device to your local machine",
    icon: "mdi-download",
    command: `scp ${username}@${props.sshid}:/path/to/file.txt ./`,
  },
  {
    title: "Port Forwarding",
    description: "Forward a local port to a port on the device (e.g., access device's web server)",
    icon: "mdi-lan-connect",
    command: `ssh -L 8080:localhost:80 ${username}@${props.sshid}`,
  },
]);

const close = () => {
  showDialog.value = false;
};

defineExpose({ showDialog });
</script>
