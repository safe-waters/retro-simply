<template>
  <div class="h-100 d-flex flex-column">
    <h1 class="text-center text-dark display-5">
      {{ column.title }}
    </h1>
    <span
      @click="addNewRetroCard"
      class="fas fa-plus ps-1 mb-1 text-dark"
      :class="{ grow: this.canAddNewRetroCard }"
      :style="{
        cursor: this.canAddNewRetroCard ? 'pointer' : '',
      }"
      >&nbsp;&nbsp;Add new card</span
    >
    <span
      @click="addNewGroup"
      class="fas fa-plus ps-1 mb-1 text-dark"
      :class="{ grow: this.canAddNewGroup }"
      :style="{
        cursor: this.canAddNewGroup ? 'pointer' : '',
      }"
      >&nbsp;&nbsp;Add new group</span
    >
    <div class="border rounded overflow-auto bg-light" style="flex-grow: 1">
      <draggable
        v-model="groups"
        :group="column.id + 'group'"
        handle=".handle"
        :disabled="!isDraggable"
      >
        <retro-group
          v-for="group in column.groups"
          :key="group.id"
          :group="group"
          :isDraggable="isDraggable"
          :cardStyle="column.cardStyle"
        ></retro-group>
      </draggable>
    </div>
  </div>
</template>

<script>
import RetroGroup from "./RetroGroup.vue";
import { mapGetters } from "vuex";
import { v4 as uuidv4 } from "uuid";
import draggable from "vuedraggable";

export default {
  name: "RetroColumn",
  props: ["column"],
  computed: {
    canAddNewRetroCard: function () {
      let defaultRetroCards = this.column.groups.filter(
        (group) => group.id === "default"
      )[0].retroCards;

      return (
        this.connected &&
        (defaultRetroCards.every((retroCard) => !retroCard.isEditable) ||
          defaultRetroCards.length === 0)
      );
    },
    canAddNewGroup: function () {
      return (
        this.connected && this.column.groups.every((group) => !group.isEditable)
      );
    },
    isDraggable: function () {
      return this.canAddNewRetroCard && this.canAddNewGroup;
    },
    groups: {
      get() {
        return this.column.groups;
      },
      set(groups) {
        let payload = {
          columnId: this.column.id,
          groups: groups,
        };
        this.$store.commit("switchGroups", payload);
      },
    },
    ...mapGetters(["connected"]),
  },
  methods: {
    addNewGroup: function () {
      if (this.canAddNewGroup) {
        let group = {
          id: uuidv4(),
          columnId: this.column.id,
          isEditable: true,
          title: "",
          retroCards: [],
        };

        this.$store.commit("addNewGroup", group);
      }
    },
    addNewRetroCard: function () {
      if (this.canAddNewRetroCard) {
        let card = {
          columnId: this.column.id,
          id: uuidv4() + "-pk-0",
          message: "",
          numVotes: 0,
          isEditable: true,
          groupId: "default",
          isDeleted: false,
          lastModified: Date.now(),
        };
        this.$store.commit("addNewRetroCard", card);
      }
    },
  },
  components: {
    "retro-group": RetroGroup,
    draggable: draggable,
  },
};
</script>

<style scoped>
.ghost {
  display: none;
}

.grow {
  transition: all 0.1s;
}

.grow:hover {
  transform: scale(1.02);
}

.grow:active {
  transform: scale(1.02);
  opacity: 0.7;
}
</style>