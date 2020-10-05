<template>
  <v-form>
    <v-container>
      <v-row
        align="center"
        justify="center"
        class="mt-4"
      >
        <v-col
          sm="8"
        >
          <div
            class="mt-6 pl-4 pr-4"
          >
            <v-row>
              <v-col
                md="auto"
                class="pr-0"
              >
                <b>
                  <!-- @change="setSessionRecord" -->
                  <v-checkbox
                    v-model="sessionRecord"
                    label="Enable session record"
                    class="pt-0 mt-0"
                  />
                </b>
                Session record is a feature that allows you to check logged activity when
                connecting to a device.
              </v-col>
            </v-row>
          </div>
        </v-col>
      </v-row>
    </v-container>
  </v-form>
</template>

<script>

export default {
  name: 'SettingSecurity',

  computed: {
    sessionRecord: {
      get() {
        return this.$store.getters['security/get'];
      },

      async set(value) {
        try {
          await this.$store.dispatch('security/set', value);
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorDefault');
        }
      },
    },
  },

  async created() {
    try {
      await this.$store.dispatch('security/get');
    } catch {
      this.$store.dispatch('snackbar/showSnackbarErrorDefault');
    }
  },
};
</script>
