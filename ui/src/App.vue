<template>
  <v-app>
    <v-navigation-drawer :clipped="clipped" v-model="drawer" :mini-variant="true" enable-resize-watcher app dark>
      <v-app-bar class="primary darken-1">
        <img
          src="http://vma.isocked.com/static/m.png"
          height="36"
          alt="ShellHub"
        >
        <v-toolbar-title class="ml-0 pl-3">
          <span class="hidden-sm-and-down">ShellHub</span>
        </v-toolbar-title>
      </v-app-bar>

      <v-divider></v-divider>
      <v-list class="grey--text">
        <v-list-item
          v-for="item in items"
          :key="item.title"
          :to="item.path"
          active-class="grey--text text--lighten-4 v-list-active"
        >
          <v-list-item-action>
            <v-icon class="grey--text text--lighten-2">{{ item.icon }}</v-icon>
          </v-list-item-action>

          <v-list-item-content>
            <v-list-item-title>{{ item.title }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>
    <v-app-bar app color="primary" dark>
      <v-spacer></v-spacer>
      <v-btn icon>
        <font-awesome-icon icon="sign-out-alt" @click="logout()">sign-out-alt</font-awesome-icon>
      </v-btn>
    </v-app-bar>
    <v-content>
      <v-container fluid>
        <router-view></router-view>
      </v-container>
    </v-content>
  </v-app>
</template>

<script>
export default {
  name: "App",

  methods: {
    logout() {
      this.$store.dispatch("auth/logout").then(() => {
        this.$router.push("/login");
      });
    }
  },

  data() {
    return {
      drawer: true,
      clipped: false,
      items: [
        {
          icon: "dashboard",
          title: "Dashboard",
          path: "/"
        },
        {
          icon: "devices",
          title: "Device Fleet",
          path: "/devices"
        },
        {
          icon: "people",
          title: "Sessions",
          path: "/sessions"
        }
      ]
    };
  }
};
</script>

<style>
.v-list-active {
  border-left: 4px solid var(--v-primary-base);
}
</style>