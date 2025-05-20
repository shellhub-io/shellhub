<template>
  <v-btn @click="dialog = !dialog" class="mr-2" data-test="namespaces-export-btn">Export CSV</v-btn>

  <v-dialog v-model="dialog" max-width="400" transition="dialog-bottom-transition">
    <v-card>
      <v-card-title class="text-h5 pb-2"> Export namespaces data </v-card-title>
      <v-divider />
      <v-form @submit.prevent="onSubmit" data-test="form">
        <v-card-text>
          <v-container>
            <v-radio-group v-model="selected">
              <v-row no-gutters class="first-row">
                <v-col class="pt-8" cols="12">
                  <v-radio label="Namespaces with more than:" value="moreThan" mt="8" />
                </v-col>
              </v-row>
              <v-row no-gutters class="d-flex justify-center align-center mb-4 ml-3">
                <v-col cols="8">
                  <v-slider v-model="numberOfDevices" hide-details :min="0" :max="150" />
                </v-col>
                <v-col cols="4">
                  <span class="ml-4">{{ numberOfDevicesRound }} devices</span>
                </v-col>
              </v-row>
              <v-row class="mb-4">
                <v-col cols="12">
                  <v-radio label="Namespaces with no devices" value="noDevices" />
                </v-col>
              </v-row>
              <v-row class="mb-4">
                <v-col cols="12">
                  <v-radio value="noSession">
                    <template v-slot:label>
                      Namespace with devices but without <br />
                      sessions
                    </template>
                  </v-radio>
                </v-col>
              </v-row>
            </v-radio-group>
          </v-container>
        </v-card-text>

        <v-card-actions class="pa-4">
          <v-spacer />
          <v-btn class="mr-2" color="dark" @click="dialog = false" type="reset"> Cancel </v-btn>
          <v-btn color="dark" type="submit" class="mr-4"> Save </v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { saveAs } from "file-saver";
import useNamespacesStore from "@admin/store/modules/namespaces";
import useSnackbar from "@/helpers/snackbar";

const numberOfDevices = ref(0);
const dialog = ref(false);
const selected = ref("moreThan");
const snackbar = useSnackbar();
const namespacesStore = useNamespacesStore();

const numberOfDevicesRound = computed(() => Math.round(numberOfDevices.value));

const generateEncodedFilter = (encodeFilter: string) => {
  let filter;
  switch (encodeFilter) {
    case "moreThan":
      filter = [
        {
          type: "property",
          params: {
            name: "devices",
            operator: "gt",
            value: String(numberOfDevicesRound.value),
          },
        },
      ];
      break;
    case "noDevices":
      filter = [
        {
          type: "property",
          params: { name: "devices", operator: "eq", value: 0 },
        },
      ];
      break;
    case "noSession":
      filter = [
        {
          type: "property",
          params: { name: "devices", operator: "gt", value: "0" },
        },
        {
          type: "property",
          params: { name: "sessions", operator: "eq", value: 0 },
        },
        { type: "operator", params: { name: "and" } },
      ];
      break;
    default:
      break;
  }
  return btoa(JSON.stringify(filter));
};

const onSubmit = async () => {
  const encodedFilter = generateEncodedFilter(selected.value);
  try {
    await namespacesStore.setFilterNamespaces(encodedFilter);
    const response = await namespacesStore.exportNamespacesToCsv();
    const blob = new Blob([response], { type: "content-disposition" });
    saveAs(
      blob,
      `namespaces_${
        selected.value === "moreThanN"
          ? `more_than_${String(numberOfDevices.value)}_devices`
          : selected.value
      }.csv`,
    );
    snackbar.showSuccess("Namespaces exported successfully.");
  } catch {
    snackbar.showError("Error exporting namespaces.");
  }
};
</script>

<style scoped>
.first-row {
  height: 70px;
}
</style>
