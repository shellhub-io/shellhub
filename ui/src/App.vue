<template>
  <v-app>
    <v-navigation-drawer :clipped="clipped" v-model="drawer" enable-resize-watcher app dark>
      <v-toolbar class="primary darken-1">
        <img
          src="http://vma.isocked.com/static/m.png"
          height="36"
          alt="ShellHub"
        >
        <v-toolbar-title class="ml-0 pl-3">
          <span class="hidden-sm-and-down">ShellHub</span>
        </v-toolbar-title>
      </v-toolbar>

      <v-divider></v-divider>
      <v-list class="grey--text">
        <v-list-tile
          v-for="item in items"
          :key="item.title"
          :to="item.path"
          active-class="grey--text text--lighten-4 v-list-active"
        >
          <v-list-tile-action>
            <v-icon class="grey--text text--lighten-2">{{ item.icon }}</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>{{ item.title }}</v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>
      </v-list>
    </v-navigation-drawer>
    <v-toolbar app color="primary"></v-toolbar>
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