<template>
  <WindowDialog
    v-model="showDialog"
    transition="dialog-bottom-transition"
    title="Generate a SSH command line"
    icon="mdi-console"
    @close="close"
  >
    <v-card-text class="pa-6">
      <v-text-field
        v-model.trim="username"
        class="mb-6"
        label="Username"
        hint="Enter an existing user on the device"
        persistent-placeholder
        persistent-hint
        density="compact"
        data-test="username-input"
      />

      <CopyCommandField
        :command
        label="Command Line"
        hint="Run this command on your terminal"
        :persistent-hint="true"
        :hide-details="false"
      />

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
import CopyCommandField from "@/components/CopyCommandField.vue";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";

interface Props {
  sshid: string;
  userId?: string;
  showCheckbox?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  userId: "",
  showCheckbox: false,
});

const showDialog = defineModel<boolean>({ required: true });
const username = ref("");
const dispenseHelper = ref(false);
const LS_KEY = "dispenseTerminalHelper";

const getDispensedUsers = (): string[] => {
  try {
    return JSON.parse(localStorage.getItem(LS_KEY) || "[]") as string[];
  } catch { return []; }
};

const setDispensedUsers = (users: string[]) => {
  localStorage.setItem(LS_KEY, JSON.stringify(users));
};

const command = computed(() => {
  const commandUsername = username.value || "<username>";
  return `ssh ${commandUsername}@${props.sshid}`;
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
