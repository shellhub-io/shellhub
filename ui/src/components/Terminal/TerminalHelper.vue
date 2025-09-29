<template>
  <WindowDialog
    v-model="showDialog"
    @close="close"
    transition="dialog-bottom-transition"
    title="Generate a SSH command line"
    icon="mdi-console"
  >
    <v-card-text class="pa-6">
      <v-text-field
        class="mb-6"
        v-model="username"
        label="Username"
        hint="Enter an existing user on the device"
        persistent-placeholder
        persistent-hint
        density="compact"
        data-test="username-input"
      />

      <CopyWarning :copied-item="'Command'">
        <template #default="{ copyText }">
          <v-text-field
            :model-value="commandLine"
            @click:append-inner="copyText(commandLine)"
            append-inner-icon="mdi-content-copy"
            hint="Run this command on your Terminal"
            class="code"
            label="Command Line"
            readonly
            density="compact"
            persistent-placeholder
            persistent-hint
            variant="outlined"
            data-test="command-field"
          />
        </template>
      </CopyWarning>

      <v-checkbox
        v-if="showCheckbox"
        v-model="dispenseHelper"
        label="Always copy SSHID directly and skip this helper"
        density="compact"
        hide-details
        data-test="dispense-checkbox"
      />
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
import { computed, ref, watch } from "vue";
import CopyWarning from "@/components/User/CopyWarning.vue";
import WindowDialog from "../WindowDialog.vue";

interface Props {
  sshid: string;
  userId?: string;
  showCheckbox?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  userId: "",
  showCheckbox: false,
});

const showDialog = defineModel({ default: false });
const username = ref("");
const dispenseHelper = ref(false);
const LS_KEY = "dispenseTerminalHelper";

const getDispensedUsers = (): string[] => {
  try {
    return JSON.parse(localStorage.getItem(LS_KEY) || "[]");
  } catch { return []; }
};

const setDispensedUsers = (users: string[]) => {
  localStorage.setItem(LS_KEY, JSON.stringify(users));
};

const commandLine = computed(() => {
  const trimmedUsername = username.value.trim();
  return trimmedUsername ? `ssh ${trimmedUsername}@${props.sshid}` : "";
});

watch(dispenseHelper, (isDispensed) => {
  const users = getDispensedUsers();
  const userIndex = users.indexOf(props.userId);

  if (isDispensed && userIndex === -1) {
    users.push(props.userId);
  } else if (!isDispensed && userIndex !== -1) {
    users.splice(userIndex, 1);
  }

  setDispensedUsers(users);
});

watch(showDialog, (isOpen) => {
  if (isOpen) {
    const users = getDispensedUsers();
    dispenseHelper.value = users.includes(props.userId);
  }
});

const close = () => {
  showDialog.value = false;
  username.value = "";
};

defineExpose({ showDialog });
</script>

<style scoped lang="scss">
.code ::v-deep(.v-field__input) {
  font-family: monospace;
  font-size: 85%;
}
</style>
