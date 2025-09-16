<template>
  <v-tooltip bottom anchor="bottom">
    <template v-slot:activator="{ props }">
      <v-icon
        @click="showDialog = true"
        tag="button"
        dark
        v-bind="props"
        tabindex="0"
        aria-label="Edit Namespace"
        data-test="dialog-btn"
      >mdi-pencil
      </v-icon>
    </template>
    <span>Edit</span>
  </v-tooltip>

  <BaseDialog v-model="showDialog" transition="dialog-bottom-transition">
    <v-card>
      <v-card-title class="text-h5 pb-2"> Edit Namespace </v-card-title>
      <v-divider />

      <form @submit.prevent="onSubmit">
        <v-card-text>
          <v-container>
            <v-row>
              <v-col cols="12">
                <v-container>
                  <v-text-field
                    v-model="name"
                    label="Name"
                    required
                    :error-messages="nameError"
                    color="primary"
                    variant="underlined"
                    data-test="name-text"
                  />

                  <v-text-field
                    v-model="maxDevices"
                    label="Maximum Devices"
                    required
                    type="number"
                    :min="-1"
                    :error-messages="maxDevicesError"
                    color="primary"
                    variant="underlined"
                    data-test="maxDevices-text"
                  />
                  <div class="d-flex align-center justify-center">
                    <span class="mr-4 text-body-1">Session record:</span>
                    <v-switch
                      v-model="sessionRecord"
                      :error-messages="sessionRecordError"
                      color="primary"
                      hide-details
                    />
                  </div>
                </v-container>
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>
        <v-card-actions class="pa-4">
          <v-spacer />
          <v-btn class="mr-2" @click="showDialog = false" type="reset"> Cancel </v-btn>
          <v-btn color="primary" type="submit" class="mr-4"> Save </v-btn>
        </v-card-actions>
      </form>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import { ref } from "vue";
import * as yup from "yup";
import useNamespacesStore from "@admin/store/modules/namespaces";
import useSnackbar from "@/helpers/snackbar";
import { IAdminNamespace } from "../../interfaces/INamespace";
import BaseDialog from "@/components/BaseDialog.vue";

const props = defineProps<{ namespace: IAdminNamespace }>();

const snackbar = useSnackbar();
const namespacesStore = useNamespacesStore();
const showDialog = ref(false);

const { value: name, errorMessage: nameError } = useField<string | undefined>(
  "name",
  yup.string().required(),
  { initialValue: props.namespace.name },
);

const { value: maxDevices, errorMessage: maxDevicesError } = useField<number | undefined>(
  "maxDevices",
  yup.number().required(),
  { initialValue: props.namespace.max_devices },
);

const { value: sessionRecord, errorMessage: sessionRecordError } = useField<boolean>(
  "sessionRecord",
  yup.boolean(),
  { initialValue: props.namespace.settings.session_record || false },
);

const hasErrors = () => nameError.value || maxDevicesError.value || sessionRecordError.value;

const onSubmit = async () => {
  if (!hasErrors()) {
    await namespacesStore.updateNamespace({
      ...props.namespace as IAdminNamespace,
      name: name.value as string,
      max_devices: Number(maxDevices.value),
      settings: { session_record: sessionRecord.value },
    });
    await namespacesStore.fetchNamespaceList();
    snackbar.showSuccess("Namespace updated successfully.");
    showDialog.value = false;
  } else {
    snackbar.showError("Please fill in all required fields.");
  }
};

defineExpose({ name, maxDevices, sessionRecord, onSubmit });
</script>
