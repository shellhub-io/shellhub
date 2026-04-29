<template>
  <v-switch
    v-model="settingEnabled"
    hide-details
    inset
    :disabled="!canUpdateSetting"
    color="primary"
    :data-test="dataTest"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";
import type { INamespaceSettings } from "@/interfaces/INamespace";

type NamespaceToggleSetting = keyof Pick<
  INamespaceSettings,
  | "allow_password"
  | "allow_public_key"
  | "allow_root"
  | "allow_empty_passwords"
  | "allow_tty"
  | "allow_tcp_forwarding"
  | "allow_web_endpoints"
  | "allow_sftp"
  | "allow_agent_forwarding"
>;

const props = defineProps<{
  tenantId: string;
  settingKey: NamespaceToggleSetting;
  permission: Parameters<typeof hasPermission>[0];
  label: string;
  dataTest: string;
}>();

const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();

const defaultValues: Record<NamespaceToggleSetting, boolean> = {
  allow_password: true,
  allow_public_key: true,
  allow_root: true,
  allow_empty_passwords: true,
  allow_tty: true,
  allow_tcp_forwarding: true,
  allow_web_endpoints: true,
  allow_sftp: true,
  allow_agent_forwarding: true,
};

const updateSetting = async (value: boolean) => {
  try {
    await namespacesStore.editNamespace({
      tenant_id: props.tenantId,
      settings: {
        [props.settingKey]: value,
      } as Partial<INamespaceSettings>,
    });

    snackbar.showSuccess(`${props.label} was successfully ${value ? "enabled" : "disabled"}.`);
  } catch (error: unknown) {
    snackbar.showError(`Failed to update ${props.label.toLowerCase()}.`);
    handleError(error);
  }
};

const settingEnabled = computed({
  get: () => namespacesStore.currentNamespace.settings?.[props.settingKey] ?? defaultValues[props.settingKey],
  set: (value: boolean) => {
    void updateSetting(value);
  },
});

const canUpdateSetting = hasPermission(props.permission);
</script>
