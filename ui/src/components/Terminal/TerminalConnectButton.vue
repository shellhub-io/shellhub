<template>
  <v-btn-group
    variant="outlined"
    density="compact"
  >
    <!-- Main connection buttons group -->
    <v-btn-group
      :color="online ? 'success' : 'normal'"
      divided
      density="compact"
      variant="outlined"
      :class="{ 'green-border': online }"
    >
      <v-btn
        :disabled="!online"
        data-test="connect-btn"
        @click="openWebTerminal"
      >
        {{ online ? "Connect" : "Offline" }}
      </v-btn>
      <v-menu>
        <template v-slot:activator="{ props }">
          <v-btn
            :disabled="!online"
            icon="mdi-triangle-small-down"
            v-bind="props"
          />
        </template>
        <v-list>
          <v-list-item
            v-for="item in menu"
            :key="item.title"
            :value="item"
            :data-test="item.title"
            @click="item.method()"
          >
            <div class="d-flex align-center">
              <v-icon :icon="item.icon" class="mr-2" />
              <v-list-item-title>
                {{ item.title }}
              </v-list-item-title>
            </div>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-btn-group>

    <!-- SRDP dedicated button for visibility -->
    <v-btn
      :disabled="!online"
      color="info"
      density="compact"
      prepend-icon="mdi-desktop-classic"
      data-test="srdp-btn"
      @click="openSRDPDialog"
      class="srdp-btn"
    >
      Remote Desktop
    </v-btn>
  </v-btn-group>
  <TerminalDialog
    v-model="showWebTerminal"
    :device-uid
    :device-name
  />
  <TerminalHelper
    v-model="showTerminalHelper"
    :sshid
  />
  <SRDPDialog
    v-model="showSRDPDialog"
    :device-uid
    :device-name
  />
</template>

<script setup lang="ts">
import { reactive, ref } from "vue";
import TerminalDialog from "./TerminalDialog.vue";
import TerminalHelper from "./TerminalHelper.vue";
import SRDPDialog from "../SRDP/SRDPDialog.vue";

defineOptions({
  inheritAttrs: false,
});

defineProps<{
  online: boolean;
  deviceUid: string;
  deviceName: string;
  sshid: string;
}>();

const showWebTerminal = ref(false);

const showTerminalHelper = ref(false);

const showSRDPDialog = ref(false);

const openWebTerminal = () => {
  showWebTerminal.value = true;
};

const openTerminalHelper = () => {
  showTerminalHelper.value = true;
};

const openSRDPDialog = () => {
  showSRDPDialog.value = true;
};

const menu = reactive([
  {
    icon: "mdi-application-outline",
    title: "Connect via web",
    type: "method",
    method: openWebTerminal,
  },
  {
    icon: "mdi-console",
    title: "Connect via terminal",
    type: "method",
    method: openTerminalHelper,
  },
]);

defineExpose({ showWebTerminal, showTerminalHelper, showSRDPDialog });
</script>

<style scoped lang="scss">
@use 'vuetify/settings';
@use 'vuetify/lib/components/VBtn/variables' as btn;

.v-btn-group {
  @each $name, $modifier in btn.$button-density {
    &.v-btn-group--density-#{$name} {
      height: calc(#{settings.$button-height} + #{$modifier}px);
    }
  }
}

.green-border {
  border: 2px solid var(--v-theme-success);
}

.v-btn-group--divided .v-btn:not(:last-child) {
  border-inline-end-color: var(--v-theme-success);
}

.srdp-btn {
  white-space: nowrap;
}

.srdp-btn {
  min-width: 140px;
  font-weight: 500;
}
</style>
