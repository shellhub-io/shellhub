<template>
  <v-btn class="mr-6" @click="dialog = !dialog" v-bind="$attrs">Export CSV</v-btn>

  <v-dialog v-model="dialog" max-width="400" transition="dialog-bottom-transition">
    <v-card>
      <v-card-title class="text-h5 pb-2"> Export users data </v-card-title>
      <v-divider />
      <v-form @submit.prevent="onSubmit">
        <v-card-text>
          <v-container>
            <v-radio-group v-model="selected">
              <v-row no-gutters class="first-row">
                <v-col class="pt-8" cols="12">
                  <v-radio label="Namespaces with more than:" value="moreThan" />
                </v-col>
              </v-row>
              <v-row no-gutters class="d-flex justify-center align-center ml-3 mt-2">
                <v-text-field
                  v-model.number="gtNumberOfNamespaces"
                  type="number"
                  label="namespaces"
                  density="comfortable"
                  variant="outlined"
                  color="primary"
                  :min="0"
                  hide-details
                />
              </v-row>

              <v-row no-gutters class="first-row">
                <v-col class="pt-8" cols="12">
                  <v-radio label="Namespaces with more than:" value="equalTo" />
                </v-col>
              </v-row>
              <v-row no-gutters class="d-flex justify-center align-center ml-3 mt-2">
                <v-text-field
                  v-model.number="eqNumberOfNamespaces"
                  type="number"
                  label="namespaces"
                  color="primary"
                  density="comfortable"
                  variant="outlined"
                  :min="0"
                  hide-details
                />
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

<script lang="ts">
import { defineComponent, ref } from "vue";
import { saveAs } from "file-saver";
import { useStore } from "../../store";
import { INotificationsError, INotificationsSuccess } from "../../interfaces/INotifications";

export default defineComponent({
  setup() {
    const store = useStore();
    const dialog = ref(false);
    const selected = ref("moreThan");
    const gtNumberOfNamespaces = ref(0);
    const eqNumberOfNamespaces = ref(0);

    const generateEncodedFilter = (encodeFilter: string) => {
      let filter;
      switch (encodeFilter) {
        case "moreThan":
          filter = [
            {
              type: "property",
              params: {
                name: "namespaces",
                operator: "gt",
                value: String(gtNumberOfNamespaces.value),
              },
            },
          ];
          break;
        case "equalTo":
          filter = [
            {
              type: "property",
              params: {
                name: "namespaces",
                operator: "eq",
                value: eqNumberOfNamespaces.value,
              },
            },
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
        await store.dispatch("users/setFilterUsers", encodedFilter);
        const response = await store.dispatch("users/exportUsersToCsv");
        const blob = new Blob([response], { type: "content-disposition" });

        if (selected.value === "moreThanN") saveAs(blob, `users_more_than_${gtNumberOfNamespaces.value}_namespaces.csv`);
        else saveAs(blob, `users_exactly_${eqNumberOfNamespaces.value}_namespaces.csv`);

        store.dispatch("snackbar/showSnackbarSuccessAction", INotificationsSuccess.exportUsers);
      } catch {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.exportUsers);
      }
    };

    return {
      dialog,
      selected,
      gtNumberOfNamespaces,
      eqNumberOfNamespaces,
      onSubmit,
    };
  },
});
</script>
