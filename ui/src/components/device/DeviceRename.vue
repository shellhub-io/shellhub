<template>
  <fragment>
    <v-tooltip bottom>
      <template v-slot:activator="{ on }">
        <v-icon
          v-on="on"
          @click="dialog = !dialog"
        >
          mdi-file-edit
        </v-icon>
      </template>
      <span>Edit</span>
    </v-tooltip>
    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card>
        {{ hostname }}
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>
export default {
  name: 'DeviceRename',

  props: {
    hostname: {
      type: String,
      required: true
    }, 
  },

  data() {
    return {
      dialog: false,
      editName: '',
    };
  },
  methods : {
    save() {
      this.$store.dispatch('devices/rename', {
        uid: this.device.uid,
        name: this.editName
      });

      this.device.name = this.editName;
    }
  }
};
</script>
