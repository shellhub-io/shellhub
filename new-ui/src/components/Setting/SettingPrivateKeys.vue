<template>
  <v-container>
    <v-row align="center" justify="center" class="mt-4">
      <v-col sm="8">
        <v-card class="pb-0 elevation-0">
          <div class="d-flex pa-0 align-center bg-v-theme-surface">
            <v-spacer />
            <v-spacer />

            <span>
              <PrivateKeyAdd data-test="privateKeyFormDialogFirst-component" />
            </span>
          </div>

          <PrivateKeyList />
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import PrivateKeyAdd from "../PrivateKeys/PrivateKeyAdd.vue";
import PrivateKeyList from "../PrivateKeys/PrivateKeyList.vue";

export default defineComponent({
  setup() {
    return {
      //
    };
  },
  components: { PrivateKeyAdd, PrivateKeyList },
});
</script>
