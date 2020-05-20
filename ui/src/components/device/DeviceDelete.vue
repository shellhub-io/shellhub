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
      <span>Delete</span>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="290"
    >
      <v-card>
        <v-card-title class="headline">
          Are you sure?
        </v-card-title>

        <v-card-text>
          Your device is going to be deleted
        </v-card-text>
        
        <v-card-actions>
          <v-spacer />

          <v-btn
            color="green darken-1"
            text
            @click="dialog = !dialog"
          >
            Cancel
          </v-btn>

          <v-btn
            color="green darken-1"
            text
            @click="remove();"
          >
            Ok
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
    uid:{
      type: String,
      required: true
    },
  },

  data() {
    return {
      dialog: false,
    };
  },

  methods:{
    async remove() {
      await this.$store.dispatch('devices/remove', this.uid);
    },
  }  
};
</script>