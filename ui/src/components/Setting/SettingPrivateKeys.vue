<template>
  <v-container>
    <v-row align="center" justify="center" class="mt-4">
      <v-col sm="8">
        <v-alert
          class="text-subtitle-2"
          type="info"
          title="ShellHub does not save your private keys"
          text="Your private keys remain confidential and are not shared with the ShellHub server.
            Instead, they are securely stored in your browser's local storage."
          color="primary"
          variant="tonal"
          data-test="privacy-policy-alert"
        />
      </v-col>
    </v-row>
    <v-row align="center" justify="center" class="mt-4">
      <v-col sm="8">
        <v-card class="pb-0 elevation-0 bg-v-theme-surface">
          <v-row>
            <v-col>
              <h3>Private Keys</h3>
            </v-col>
            <v-col>
              <div class="d-flex pa-0 align-center bg-v-theme-surface">
                <v-spacer />
                <v-spacer />

                <span>
                  <PrivateKeyAdd data-test="privateKeyFormDialogFirst-component" />
                </span>
              </div>
            </v-col>
          </v-row>

          <PrivateKeyList />
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import PrivateKeyAdd from "../PrivateKeys/PrivateKeyAdd.vue";
import PrivateKeyList from "../PrivateKeys/PrivateKeyList.vue";
</script>
