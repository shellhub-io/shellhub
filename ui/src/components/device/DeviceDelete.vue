<template>
  <fragment>
    <v-tooltip bottom>
      <template #activator="{ on }">
        <v-icon
          v-on="on"
          @click="dialog = !dialog"
        >
          delete
        </v-icon>
      </template>
      <span>Remove</span>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card>
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          You are about to remove this device
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            @click="dialog=!dialog"
          >
            Close
          </v-btn>

          <v-btn
            color="red darken-1"
            text
            @click="remove();"
          >
            Remove
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

export default {
  name: 'DeviceDelete',

  props: {
    uid: {
      type: String,
      required: true,
    },

    redirect: {
      type: Boolean,
    },
  },

  data() {
    return {
      dialog: false,
    };
  },

  methods: {
    async remove() {
      await this.$store.dispatch('devices/remove', this.uid);
      this.dialog = !this.dialog;
      if (this.redirect) {
        this.$router.push('/devices');
      }
      this.$emit('update');
    },
  },
};

</script>
